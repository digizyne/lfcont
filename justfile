set shell := ["/usr/bin/env", "bash", "-c"]

default: fmt up

fmt:
    go fmt ./...

up:
    docker compose --profile dev up

down:
    docker compose down --rmi local --remove-orphans

build:
    go build -o ~/go/bin/lfcont ./cmd/main.go

build-docker:
    docker build -t lfcont:dev -f Dockerfile .

tidy:
    go mod tidy

add PACKAGE:
    go get -u {{PACKAGE}}
