FROM golang:alpine AS builder

RUN apk update --no-cache && \
    apk upgrade --no-cache

WORKDIR /build

COPY . .

RUN go mod download && \
    go mod tidy && \
    go build -o site-form-checker ./cmd/main/main.go

FROM alpine

RUN apk update --no-cache && \
    apk upgrade --no-cache && \
    apk add --no-cache curl

WORKDIR /build

COPY --from=builder /build/site-form-checker /build/site-form-checker
COPY --from=builder /build/.env /build/.env

#CMD ["./site-form-checker"]