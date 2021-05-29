FROM golang:1.16

WORKDIR /app

RUN apt-get update \
    && apt-get install -y mkvtoolnix ffmpeg \
    && apt-get clean

COPY go.mod go.sum /app/

RUN go mod download
