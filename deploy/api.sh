#!/bin/bash -e

ROOT="$(dirname "$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )")"

mkdir -p $ROOT/.artifacts/docker

docker build -f $ROOT/api/cmd/api/Dockerfile -t confa/api $ROOT/api
docker save -o $ROOT/.artifacts/docker/api.tar confa/api
scp -C $ROOT/.artifacts/docker/api.tar $USER@$IP:~
scp -C $ROOT/docker-compose.yml $USER@$IP:~
ssh $USER@$IP "docker load -i ~/api.tar"
ssh $USER@$IP \
  EMAIL_SERVER="$EMAIL_SERVER" \
  EMAIL_PORT="$EMAIL_PORT" \
  EMAIL_ADDRESS="$EMAIL_ADDRESS" \
  EMAIL_PASSWORD="$EMAIL_PASSWORD" \
  "docker-compose up --no-deps -d"
ssh $USER@$IP "docker image prune -f"
