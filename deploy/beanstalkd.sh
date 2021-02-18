#!/bin/bash -e

ROOT="$(dirname "$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )")"

mkdir -p $ROOT/.artifacts/docker

docker build -f $ROOT/beanstalkd/Dockerfile -t confa/beanstalkd $ROOT/beanstalkd
docker save -o $ROOT/.artifacts/docker/beanstalkd.tar confa/beanstalkd
scp -C $ROOT/.artifacts/docker/beanstalkd.tar $USER@$IP:~
scp -C $ROOT/docker-compose.yml $USER@$IP:~
ssh $USER@$IP "docker load -i ~/beanstalkd.tar"
ssh $USER@$IP "docker-compose up -d"
ssh $USER@$IP "docker image prune -f"
