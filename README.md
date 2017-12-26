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