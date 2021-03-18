#!/bin/bash -e

IMAGE=confa/migrate

docker build . -t ${IMAGE}
docker run \
  --rm \
  -w /app \
  -v `pwd`:/app \
  --network=host \
  --env POSTGRES_HOST="localhost" \
  --env POSTGRES_PORT="5432" \
  --env POSTGRES_DATABASE="confa" \
  --env POSTGRES_USER="confa" \
  --env POSTGRES_PASSWORD="confa" \
  ${IMAGE} tern $@
