#!/bin/bash -e

OUT_PATH=.

protoc rtc.proto \
    --proto_path=. \
    --go_opt=paths=source_relative \
    --go_opt=Mrtc.proto=github.com/aromancev/confa/internal/proto/rtc \
    --go_out=${OUT_PATH}/rtc \
    --twirp_out=${OUT_PATH}/rtc

protoc queue.proto \
    --proto_path=. \
    --go_opt=paths=source_relative \
    --go_opt=Mqueue.proto=github.com/aromancev/confa/internal/proto/queue \
    --go_out=${OUT_PATH}/queue \
    --twirp_out=${OUT_PATH}/queue

protoc iam.proto \
    --proto_path=. \
    --go_opt=paths=source_relative \
    --go_opt=Miam.proto=github.com/aromancev/confa/internal/proto/iam \
    --go_out=${OUT_PATH}/iam \
    --twirp_out=${OUT_PATH}/iam

protoc confa.proto \
    --proto_path=. \
    --go_opt=paths=source_relative \
    --go_opt=Mconfa.proto=github.com/aromancev/confa/internal/proto/confa \
    --go_out=${OUT_PATH}/confa \
    --twirp_out=${OUT_PATH}/confa

protoc tracker.proto \
    --proto_path=. \
    --go_opt=paths=source_relative \
    --go_opt=Mtracker.proto=github.com/aromancev/confa/internal/proto/tracker \
    --go_out=${OUT_PATH}/tracker \
    --twirp_out=${OUT_PATH}/tracker

protoc avp.proto \
    --proto_path=. \
    --go_opt=paths=source_relative \
    --go_opt=Mavp.proto=github.com/aromancev/confa/internal/proto/avp \
    --go_out=${OUT_PATH}/avp \
    --twirp_out=${OUT_PATH}/avp
