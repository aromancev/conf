FROM golang:1.16-alpine as builder

WORKDIR /app
COPY . ./
RUN go build \
  -o bin/api \
  ./cmd/api

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/bin/* /app/

CMD ["/app/api"]
