FROM golang:1.16-alpine as builder

WORKDIR /app

COPY . ./
RUN go build \
  -o bin/media \
  ./cmd/media

# Consider using
# https://github.com/wader/static-ffmpeg
FROM alpine:latest

WORKDIR /app

RUN apk update \
    && apk add --no-cache ffmpeg

COPY --from=builder /app/bin/* /app/

CMD ["/app/media"]
