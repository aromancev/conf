FROM golang:1.19-alpine as builder

WORKDIR /app
COPY . ./
RUN go build \
  -o bin/ \
  ./cmd/avp

FROM alpine:3.16

RUN apk update \
    && apk add --no-cache ffmpeg=5.0.1-r1

RUN addgroup -g 1001 -S app && mkdir -p /app && adduser -u 1001 -S -G app -h /app app
USER app
WORKDIR /app

COPY --from=builder --chown=app /app/bin/* /app/

CMD ["/app/avp"]
