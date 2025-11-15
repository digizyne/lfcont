FROM golang:1.25.3-trixie AS development
WORKDIR /app
COPY go.mod go.sum sakey.json ./
ENV GOOGLE_APPLICATION_CREDENTIALS=/app/sakey.json
RUN go mod download && go install github.com/air-verse/air@latest && curl -fsSL https://get.pulumi.com | sh && /root/.pulumi/bin/pulumi login gs://lf-controller-pulumi-state-staging
ENV PATH="/root/.pulumi/bin:${PATH}" 
COPY . .

FROM development AS build
ENV CGO_ENABLED=0 GOOS=linux
RUN go build -ldflags "-s -w" -o /app/controller ./cmd/main.go

# TODO: change to alpine, install curl, then run pulumi install script
FROM alpine:3.22.2 AS production
WORKDIR /app
RUN apk add --no-cache curl && curl -fsSL https://get.pulumi.com | sh
ENV GOOGLE_APPLICATION_CREDENTIALS=/app/sakey.json
ENV PATH="/root/.pulumi/bin:${PATH}"
COPY --from=build /app/controller .
EXPOSE 8080
ENTRYPOINT [ "sh", "-c" ]
CMD ["pulumi login gs://lf-controller-pulumi-state-staging && ./controller" ]