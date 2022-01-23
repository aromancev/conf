FROM golang:1.16

WORKDIR /app

# Consider using
# https://github.com/wader/static-ffmpeg
RUN apt-get update \
    && apt-get install -y ffmpeg \
    && apt-get clean
