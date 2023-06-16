#!/bin/bash -e

rm -rf .artifacts/hashistack/secrets
docker run \
    --rm \
    -w /etc/secrets \
	-v $PWD/.artifacts/hashistack/secrets:/etc/secrets  \
    hashicorp/consul:1.15.3 sh -c "
        consul keygen > consul_gossip_key.txt
        consul tls ca create
        consul tls cert create -server -dc dc1 -domain consul
    "
