FROM golang:1.14-alpine

WORKDIR /app
COPY go.mod go.sum /app/

RUN go mod download
