# readr-restful
RESTful API for the READr site

## Start the server 

Use `config/main_sample.json` as template, put your mySQL connection address, username and password in a file `main.json` and save it in `config/` 

### with go run

```bash
go run *.go
```

### go run ignoring *_test.go files

```bash
go run $(ls -1 *.go | grep -v _test.go)
```

### run the test
```bash
go test ./routes
```

## Using **MAKE** tool

In this project we build some handy tools as alternatives to original go toolsets. These commands are synonyms to for verbose original go tool command, and could be used to build, test, run, or get dependencies.

### run the server
This was set to run all *.go files except _test.go files. You could run the server by simply typing in:

```bash
make run
```

### run the test
It's default to run test in **all** directory.

```bash
make test
```

### install dependencies
Use this to install package.

```bash
make deps
```


### build binary for alpine linux

```bash
make build-alpine
```