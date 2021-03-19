#!/bin/bash -e

ROOT="$(dirname "$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )")"

mkdir -p $ROOT/.artifacts/docker

docker run \
		--rm \
		-w /app \
		-v $ROOT/web:/app \
		node:15.7.0 /bin/sh -c "npm install -g pnpm; pnpm install; pnpm run build"
docker build -f $ROOT/web/Dockerfile -t confa/web $ROOT/web
docker save -o $ROOT/.artifacts/docker/web.tar confa/web
scp -C $ROOT/.artifacts/docker/web.tar $USER@$IP:~
ssh $USER@$IP "docker load -i ~/web.tar"
$ROOT/deploy/up.sh
