FROM golang:alpine

RUN apk update --no-cache && \
    apk upgrade --no-cache

RUN mkdir build
WORKDIR /build

COPY . .

RUN go mod download && \
    go mod tidy

RUN go build cmd/main/main.go

RUN mkdir bin