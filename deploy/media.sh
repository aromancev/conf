#!/bin/bash -e

ROOT="$(dirname "$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )")"

mkdir -p $ROOT/.artifacts/docker

docker build -f $ROOT/api/cmd/media/Dockerfile -t confa/media $ROOT/media
docker save -o $ROOT/.artifacts/docker/media.tar confa/media
scp -C $ROOT/.artifacts/docker/media.tar $USER@$IP:~
ssh $USER@$IP "docker load -i ~/media.tar"
$ROOT/deploy/up.sh
