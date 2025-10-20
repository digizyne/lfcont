default: fmt run

fmt:
    go fmt ./...

run:
    go run main.go

build:
    go build -o lfcont main.go

tidy:
    go mod tidy

add PACKAGE:
    go get -u {{PACKAGE}}

install:
    go install -o lf github.com/digizyne/local-first