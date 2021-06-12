FROM golang:1.16-alpine as builder

WORKDIR /app
COPY . /app

RUN go build \
  -o bin/media \
  ./cmd/media

FROM alpine:latest

WORKDIR /app

RUN apk update \
    && apk add --no-cache ffmpeg

COPY --from=builder /app/bin/* /app/

CMD ["/app/media"]
