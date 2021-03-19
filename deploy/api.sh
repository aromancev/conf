#!/bin/bash -e

ROOT="$(dirname "$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )")"

mkdir -p $ROOT/.artifacts/docker

docker build -f $ROOT/api/cmd/api/Dockerfile -t confa/api $ROOT/api
docker save -o $ROOT/.artifacts/docker/api.tar confa/api
scp -C $ROOT/.artifacts/docker/api.tar $USER@$IP:~
ssh $USER@$IP "docker load -i ~/api.tar"
$ROOT/deploy/up.sh
