# readr-restful
RESTful API for the READr site

# Start the server 

Use `config/main_sample.json` as template, put your mySQL connection address, username and password in a file `main.json` and save it in `config/` 
You could also use a customized config file, like `main_dev.json`, put under directory `config/`.


## with go run

```bash
go run main.go
```

when using customized configuration file:

```bash 
go run main.go -file=main_dev
```

## go run ignoring *_test.go files

```bash
go run $(ls -1 *.go | grep -v _test.go)
```

## run the test
```bash
go test ./routes
```

# Using **MAKE** tool

In this project we build some handy tools as alternatives to original go toolsets. These commands are synonyms to for verbose original go tool command, and could be used to build, test, run, or get dependencies.

## run the server
This was set to run all *.go files except _test.go files. You could run the server by simply typing in:

```bash
make run
```

## run the test
It's default to run test in **all** directory.

```bash
make test
```

## install dependencies
Use this to install package.

```bash
make deps
```


## build binary for alpine linux

```bash
make build-alpine
```

# Request

## Single Post/Member

## GET
```bash
GET /member/[id]
GET /post/[id]
```

## Multiple Posts/Members

## GET
```bash
GET /posts
GET /members
```

### Pagination

```bash
/posts?max_result=50&page=1&sort=-updated_at
/posts?sort=-created_at,author.nickname
```

Default : `max_result=20`, `page=1`, `sort=-updated_at`

Use data keys for ascending sort order. Attach `-` for descending. For example, `sort=updated_by` is sorting using **ascending** `updated_by`, while `sort=-updated_by` is **descending**.

Multiple field sorting could be achieved with keys seperated by comma **`,`**

### Filter

#### Posts

```bash
/posts?author={"$in":["superman"]}&active{"$nin":[1,3]}
```

There are posts filters for **author** and **active** now.
To filter the posts of which the author is superman, and posts status `active` not `[1,3]` pass query parameter as follows:

Currently operator only support `$in` and `$nin`
