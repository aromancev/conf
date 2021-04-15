#!/bin/bash -e

export IMAGE=confa/api
docker build . -t ${IMAGE}
docker run \
  --rm \
  -w /app \
  -v `pwd`:/app \
  ${IMAGE} go generate ./...
