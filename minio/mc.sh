#!/bin/bash -e

docker run \
	--rm \
	-ti \
	--network="confa" \
	-v `pwd`/minio/config.json:/root/.mc/config.json \
	minio/mc $@
