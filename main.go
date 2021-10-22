package main

import (
	"fmt"
	"github.com/asaskevich/EventBus"
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"github.com/jdeng/goheif"
	"github.com/skratchdot/open-golang/open"
	"image/jpeg"
	"net/url"
	"os"
	systempaths "path"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

func reportError(err error) {
	fmt.Println(err)
	os.Exit(1)
}

var (
	progress int
	total    int
	handled  bool
)

func main() {
	progress = 0
	total = 0
	bus := EventBus.New()
	gtk.Init(nil)
	window, err := gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
	if err != nil {
		reportError(err)
	}

	header, _ := gtk.HeaderBarNew()
	header.SetShowCloseButton(true)
	window.SetTitlebar(header)
	window.SetTitle("Heicy")

	window.Connect("destroy", func() {
		gtk.MainQuit()
	})
	mainLabel, _ := gtk.LabelNew("Drop HEIC files here")
	mainLabel.SetMarginTop(110)
	secondaryLabel, _ := gtk.LabelNew("or")
	chooseFilesButton, _ := gtk.ButtonNewWithLabel("Choose files")
	chooseFilesButton.SetMarginBottom(140)
	chooseFilesButton.SetMarginStart(100)
	chooseFilesButton.SetMarginEnd(100)
	chooseFilesButton.Connect("clicked", func() {
		inputFilesChooser, _ := gtk.FileChooserNativeDialogNew("Select files", window, gtk.FILE_CHOOSER_ACTION_OPEN, "_Open", "_Cancel")
		inputFilesChooser.SetSelectMultiple(true)
		response := inputFilesChooser.NativeDialog.Run()
		if gtk.ResponseType(response) == gtk.RESPONSE_ACCEPT {
			filenames, err := inputFilesChooser.GetFilenames()
			if err == nil {
				start(filenames, window, bus)
			}
		}
	})
	mainBox, err := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 6)
	mainBox.PackStart(mainLabel, true, false, 0)
	mainBox.PackStart(secondaryLabel, true, false, 0)
	mainBox.PackStart(chooseFilesButton, true, false, 0)

	targetEntry, _ := gtk.TargetEntryNew("text/plain", gtk.TARGET_OTHER_APP, 420)
	mainBox.DragDestSet(gtk.DEST_DEFAULT_ALL, []gtk.TargetEntry{*targetEntry}, gdk.ACTION_COPY)
	mainBox.Connect("drag-data-received", func(widget interface{}, dragContext interface{}, x int, y int, data *gtk.SelectionData, info interface{}, time interface{}) {
		paths := string(data.GetData())
		normalizedPaths := make([]string, 0)
		for _, p := range strings.Split(paths, "\n") {
			if len(p) > 0 {
				normalizedPaths = append(normalizedPaths, p)
			}
		}
		start(normalizedPaths, window, bus)
	})

	progressBar, _ := gtk.ProgressBarNew()
	progressBar.SetMarginStart(70)
	progressBar.SetMarginEnd(70)
	progressBar.SetSizeRequest(220, 10)
	notesLabel, _ := gtk.LabelNew("1/173 files processed")
	notesLabel.SetJustify(gtk.JUSTIFY_CENTER)
	notesLabel.SetSizeRequest(360, 20)
	convertAgainButton, _ := gtk.ButtonNewWithLabel("Convert again!")
	convertAgainButton.SetMarginStart(100)
	convertAgainButton.SetMarginEnd(100)
	convertAgainButton.SetSizeRequest(160, 20)

	workingLayout, _ := gtk.FixedNew()
	workingLayout.Put(progressBar, 0, 140)
	workingLayout.Put(notesLabel, 0, 170)

	//window.Add(workingLayout)
	convertAgainButton.Connect("clicked", func() {
		progress = 0
		total = 0
		notesLabel.SetText(" 0/0 files processed")
		progressBar.SetFraction(0)
		window.Remove(workingLayout)
		window.Add(mainBox)
		window.ShowAll()

	})
	glib.TimeoutAdd(100, func() bool {
		if handled {
			return true
		}
		handled = true
		fmt.Printf("Progress: %d\n", progress)
		if progress == 0 && total > 0 {
			window.Remove(mainBox)
			window.Add(workingLayout)
			window.ShowAll()
		} else if progress == total && total > 0 {
			workingLayout.Put(convertAgainButton, 0, 220)
			notesLabel.SetText(strconv.Itoa(progress) + "/" + strconv.Itoa(total) + " files processed")
			if progress != 0 && total != 0 {
				progressBar.SetFraction(float64(progress) / float64(total))
			}

			window.ShowAll()
		} else {
			notesLabel.SetText(strconv.Itoa(progress) + "/" + strconv.Itoa(total) + " files processed")
			if progress != 0 && total != 0 {
				progressBar.SetFraction(float64(progress) / float64(total))
			}
		}
		return true
	})

	window.Add(mainBox)
	window.SetResizable(false)
	window.SetDefaultSize(360, 360)
	window.ShowAll()
	runtime.LockOSThread()
	gtk.Main()
}

func start(paths []string, window *gtk.Window, bus EventBus.Bus) {
	outputFolderChooser, _ := gtk.FileChooserNativeDialogNew("Select output folder", window, gtk.FILE_CHOOSER_ACTION_SELECT_FOLDER, "_Open", "_Cancel")
	response := outputFolderChooser.NativeDialog.Run()
	if gtk.ResponseType(response) == gtk.RESPONSE_ACCEPT {
		filenames := outputFolderChooser.GetFilename()
		outputFolderChooser.Hide()
		go convertFiles(paths, filenames, window, bus)
	}
}

func convertFiles(paths []string, destination string, window *gtk.Window, bus EventBus.Bus) {
	if len(paths) == 0 {
		return
	}
	progress = 0
	total = len(paths)
	handled = false
	time.Sleep(time.Second)
	for i, path := range paths {
		progress = i + 1
		handled = false
		if strings.HasSuffix(strings.TrimSpace(strings.ToLower(path)), "heic") || strings.HasSuffix(strings.TrimSpace(strings.ToLower(path)), "heif") {
			normalizedPath, err := url.QueryUnescape(strings.ReplaceAll(strings.TrimSpace(path), "file://", ""))
			handler, err := os.Open(normalizedPath)
			if err == nil {
				outputHandler, _ := os.OpenFile(systempaths.Join(destination, strings.ReplaceAll(filepath.Base(path), lastString(strings.Split(filepath.Base(path), ".")), "png")), os.O_RDWR|os.O_CREATE, 0644)
				exif, err := goheif.ExtractExif(handler)
				if err == nil {
					image, err := goheif.Decode(handler)
					if err == nil {
						w, err := newWriterExif(outputHandler, exif)
						if err == nil {
							err = jpeg.Encode(w, image, nil)
							if err == nil {
								fmt.Printf("Processed %s\n", normalizedPath)
							}
						}
					}
				}
			}
		}
	}
	progress = len(paths)
	total = len(paths)
	handled = false
	open.Run(destination)
}

func lastString(ss []string) string {
	return ss[len(ss)-1]
}
