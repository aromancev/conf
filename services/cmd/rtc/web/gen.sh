#!/bin/bash -e

quicktype \
    --src-lang schema \
    --out schema.go \
    --package web \
    --top-level Message \
    room.schema.json
