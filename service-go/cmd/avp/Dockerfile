FROM golang:1.17-alpine as builder

WORKDIR /app
COPY . ./
RUN go build \
  -o bin/ \
  ./cmd/avp

FROM alpine:latest

WORKDIR /app

RUN apk update \
    && apk add --no-cache ffmpeg=4.4.1-r2

COPY --from=builder /app/bin/* /app/

CMD ["/app/avp"]
