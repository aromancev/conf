FROM golang:1.16

WORKDIR /app

ENV PROTOC_VERSION=3.17.2

# intel: x86_64
# arm/m1: aarch_64
ENV ARC=aarch_64

RUN apt-get update \
    && apt-get install -y unzip \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/* \
    && wget https://github.com/protocolbuffers/protobuf/releases/download/v${PROTOC_VERSION}/protoc-${PROTOC_VERSION}-linux-${ARC}.zip \
    && unzip protoc-${PROTOC_VERSION}-linux-${ARC}.zip \
    && mv -v bin/* /usr/bin/  \
    && mv -v include/* /usr/include/  \
    && rm -r bin include protoc-${PROTOC_VERSION}-linux-${ARC}.zip \
    && go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.26.0 \
    && go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.1.0 \
    && go install github.com/twitchtv/twirp/protoc-gen-twirp@v8.1.0
