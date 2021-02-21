FROM golang:1.14-alpine as builder

WORKDIR /app
COPY . /app

RUN go build \
  -o bin/api \
  ./cmd/api

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/bin/* /app/

ENV EMAIL_=/bar

CMD ["/app/api"]
