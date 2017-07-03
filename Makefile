build:
	cd cmd/linkchecker && go build

run: build
	cmd/linkchecker/linkchecker -host=www.ardanstudios.com
