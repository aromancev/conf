#!/bin/bash -e

ROOT="$(dirname "$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )")"

mkdir -p $ROOT/.artifacts/docker

docker build -f $ROOT/api/cmd/sfu/Dockerfile -t confa/sfu $ROOT/api
docker save -o $ROOT/.artifacts/docker/sfu.tar confa/sfu
scp -C $ROOT/.artifacts/docker/sfu.tar $USER@$IP:~
ssh $USER@$IP "docker load -i ~/sfu.tar"
$ROOT/deploy/up.sh
