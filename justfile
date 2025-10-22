set shell := ["/usr/bin/env", "bash", "-c"]

default: fmt run

fmt:
    go fmt ./...

run:
    air

build:
    go build -o lfcont main.go

tidy:
    go mod tidy

add PACKAGE:
    go get -u {{PACKAGE}}

install:
    go install -o lf github.com/digizyne/local-first