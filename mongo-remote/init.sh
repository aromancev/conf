#!/bin/bash -e

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
IMAGE=mongo:4.4

docker run \
  --rm \
  --network="host" \
  -v ${DIR}/init-rs.js:/app/init-rs.js \
  ${IMAGE} mongo --quiet "mongodb://mongo:mongo@localhost:27017" /app/init-rs.js

docker run \
  --rm \
  --network="host" \
  -v ${DIR}/create-users.js:/app/create-users.js \
  ${IMAGE} mongo --quiet --eval="const iamPwd = 'iam'; const rtcPwd = 'rtc'; const confaPwd = 'confa'; " "mongodb://mongo:mongo@localhost:27017/?replicaSet=rs" /app/create-users.js
