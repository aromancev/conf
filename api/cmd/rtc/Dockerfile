FROM golang:1.16-alpine as builder

WORKDIR /app
COPY . /app

RUN go build \
  -o bin/rtc \
  ./cmd/rtc

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/bin/* /app/

CMD ["/app/rtc"]
