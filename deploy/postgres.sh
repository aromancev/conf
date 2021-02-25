#!/bin/bash -e

ROOT="$(dirname "$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )")"

scp -C $ROOT/docker-compose.yml $USER@$IP:~
ssh $USER@$IP "docker-compose up --no-deps -d"
ssh $USER@$IP "docker image prune -f"
