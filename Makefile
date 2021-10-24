main:
	go get -v
	go build heicy
	rm heicy_amd64/usr/local/bin/heicy
	mv heicy heicy_amd64/usr/local/bin/
	dpkg-deb --build --root-owner-group heicy_amd64