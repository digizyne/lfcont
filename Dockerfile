FROM golang:1.25.3-trixie AS development
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download && go install github.com/air-verse/air@latest
COPY . .

FROM development AS build
ENV CGO_ENABLED=0 GOOS=linux
RUN go build -ldflags "-s -w" -o /app/lfcont ./cmd/main.go

FROM gcr.io/distroless/static AS production
WORKDIR /app
COPY --from=build /app/lfcont .
EXPOSE 8080
CMD ["./lfcont"]