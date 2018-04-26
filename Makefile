BINARY :=app
ALLGOFILES := $(shell ls -1 *.go | grep -v _test.go)

all: deps test build
build:
	go build -a -o $(BINARY) -v

.PHONY: test run

deps:
	go get -v -d
test:
	go test -v ./...
run:
	go run $(ALLGOFILES)
build-alpine: deps test
	env GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -a -o $(BINARY) main.go
