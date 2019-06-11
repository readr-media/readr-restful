BINARY :=app
ALLGOFILES := $(shell ls -1 *.go | grep -v _test.go)

all: deps test build
build:
	go build -a -o $(BINARY) -v

.PHONY: test run

deps:
	go get -v -d
	# cd $(GOPATH)/src/github.com/gin-gonic/gin && git checkout tags/v1.3.0  && cd -
test:
	# disable cgo to avoid gcc not found re-compile error
	env CGO_ENABLED=0 go test -v ./...
run:
	go run $(ALLGOFILES)
build-alpine: deps test
	env GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -a -o $(BINARY) main.go
build-stage: deps
	env GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -a -o $(BINARY) main.go
