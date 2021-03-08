FROM golang:1.16-alpine

WORKDIR /app
COPY go.mod go.sum /app/

RUN go mod download
