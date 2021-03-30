#!/bin/bash -e

export IMAGE=confa/migrate

docker build . -t ${IMAGE}
echo "--- RUN MIGRATIONS ---"
docker run \
  --rm \
  -w /app \
  -v `pwd`:/app \
  --network=confa_default \
  --env POSTGRES_HOST="postgres" \
  --env POSTGRES_PORT="5432" \
  --env POSTGRES_DATABASE="confa" \
  --env POSTGRES_USER="confa" \
  --env POSTGRES_PASSWORD="confa" \
  ${IMAGE} tern $@
