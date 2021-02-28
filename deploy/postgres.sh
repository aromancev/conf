#!/bin/bash -e

ROOT="$(dirname "$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )")"

scp -C $ROOT/docker-compose.yml $USER@$IP:~
ssh $USER@$IP \
  POSTGRES_HOST="$POSTGRES_HOST" \
  POSTGRES_PORT="$POSTGRES_PORT" \
  POSTGRES_USER="$POSTGRES_USER" \
  POSTGRES_PASSWORD="$POSTGRES_PASSWORD" \
  POSTGRES_DATABASE="$POSTGRES_DATABASE" \
  "docker-compose up --no-deps -d"

ssh $USER@$IP "docker image prune -f"
