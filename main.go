package main

import (
	"fmt"
	"github.com/gotk3/gotk3/gtk"
	"os"
)

func reportError(err error) {
	fmt.Println(err)
	os.Exit(1)
}

func main() {
	gtk.Init(nil)
	window, err := gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
	if err != nil {
		reportError(err)
	}
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
	mainBox, err := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 6)
	mainBox.PackStart(mainLabel, true, false, 0)
	mainBox.PackStart(secondaryLabel, true, false, 0)
	mainBox.PackStart(chooseFilesButton, true, false, 0)
	window.Add(mainBox)
	window.SetResizable(false)
	window.SetDefaultSize(360, 360)
	window.ShowAll()
	gtk.Main()
}
