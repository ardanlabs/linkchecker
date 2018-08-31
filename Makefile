build:
	cd cmd/linkchecker && go build

run: build
	cmd/linkchecker/linkchecker -host=www.ardanlabs.com

docker:
	cd cmd/linkchecker && GOOS=linux go build
	docker build -t "quay.io/ardanlabs/linkchecker" .
	docker push "quay.io/ardanlabs/linkchecker"

build-linux:
	cd cmd/linkchecker && GOOS=linux CGO_ENABLED=0 go build

test:
	go test -v ./...
