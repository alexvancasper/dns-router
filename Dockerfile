#---Build stage---
FROM golang:1.21.3 AS builder
COPY . /go/src/
WORKDIR /go/src/src/
RUN go mod tidy
RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags='-w -s' -o /go/bin/nameserver

#---Final stage---
FROM alpine:latest
COPY --from=builder /go/bin/nameserver /nameserver
CMD ["/nameserver", "-c", "/config.yaml"] 