FROM golang:1.19-alpine

RUN apk update \
    && apk add --no-cache ffmpeg=5.0.1-r1
