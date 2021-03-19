#!/bin/bash -e

ROOT="$(dirname "$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )")"

scp -C $ROOT/deploy/docker-compose.yml $USER@$IP:~
ssh $USER@$IP \
  SECRET_KEY="$SECRET_KEY" \
  PUBLIC_KEY="$PUBLIC_KEY" \
  EMAIL_SERVER="$EMAIL_SERVER" \
  EMAIL_PORT="$EMAIL_PORT" \
  EMAIL_ADDRESS="$EMAIL_ADDRESS" \
  EMAIL_PASSWORD="$EMAIL_PASSWORD" \
  POSTGRES_HOST="$POSTGRES_HOST" \
  POSTGRES_PORT="$POSTGRES_PORT" \
  POSTGRES_USER="$POSTGRES_USER" \
  POSTGRES_PASSWORD="$POSTGRES_PASSWORD" \
  POSTGRES_DATABASE="$POSTGRES_DATABASE" \
  BEANSTALKD_POOL="$BEANSTALKD_POOL" \
  "docker-compose up --no-deps -d"
ssh $USER@$IP "docker image prune -f"
