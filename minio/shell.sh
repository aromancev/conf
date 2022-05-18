#!/bin/bash -e

docker run \
	--rm \
	-ti \
	--network="confa" \
	--entrypoint=/bin/sh \
	-v `pwd`/minio/config.json:/root/.mc/config.json \
	minio/mc
