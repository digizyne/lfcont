set shell := ["/usr/bin/env", "bash", "-c"]

default: fmt up

fmt:
    go fmt ./...

up:
    docker compose --profile dev up

up-local:
    docker compose --profile local up

down:
    docker compose --profile dev down --rmi local --remove-orphans

down-local:
    docker compose --profile local down --rmi local --remove-orphans

build:
    go build -o ~/go/bin/lfcont ./cmd/main.go

build-docker:
    docker build -t lfcont:dev -f Dockerfile .

run-docker:
    docker run -it --rm --name lfcont -p 8080:8080 -v /var/run/docker.sock:/var/run/docker.sock lfcont:dev

tidy:
    go mod tidy

add PACKAGE:
    go get -u {{PACKAGE}}
