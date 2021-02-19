#!/bin/bash -e

ROOT="$(dirname "$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )")"

mkdir -p $ROOT/.artifacts/docker

docker build -f $ROOT/api/cmd/rtc/Dockerfile -t confa/rtc $ROOT/api
docker save -o $ROOT/.artifacts/docker/rtc.tar confa/rtc
scp -C $ROOT/.artifacts/docker/rtc.tar $USER@$IP:~
scp -C $ROOT/docker-compose.yml $USER@$IP:~
ssh $USER@$IP "docker load -i ~/rtc.tar"
ssh $USER@$IP "docker-compose up --no-deps -d"
ssh $USER@$IP "docker image prune -f"
