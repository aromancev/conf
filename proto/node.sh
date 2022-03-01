#!/bin/bash -e

GO_PATH=../service-go

quicktype \
    --src-lang schema \
    --out ${GO_PATH}/cmd/rtc/web/schema.go \
    --package web \
    --top-level Message \
    ${GO_PATH}/cmd/rtc/web/room.schema.json
