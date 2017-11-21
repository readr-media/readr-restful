# readr-restful
RESTful API for the READr site

## Start the server 

### with go run

```bash
go run *.go --sql-user=[USER ACCOUNT] --sql-address=[SQL SERVER ADDR] --sql-auth=[SQL PASSWORD]
```

### go run ignoring *_test.go files

```bash
go run $(ls -1 *.go | grep -v _test.go)  --sql-user=[USER ACCOUNT] --sql-address=[SQL SERVER ADDR] --sql-auth=[SQL PASSWORD]
```