FROM golang:1.16-alpine

WORKDIR /app

RUN apk add --no-cache protoc \
    && go install github.com/jackc/tern@v1.12.4 \
    && go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.26.0 \
    && go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.1.0

COPY go.mod go.sum /app/

RUN go mod download
