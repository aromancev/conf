#!/bin/bash -e

GO_PATH=../service-go

protoc rtc.proto \
    --proto_path=. \
    --go_opt=paths=source_relative \
    --go_opt=Mrtc.proto=github.com/aromancev/confa/internal/proto/rtc \
    --go_out=${GO_PATH}/internal/proto/rtc \
    --twirp_out=${GO_PATH}/internal/proto/rtc

protoc queue.proto \
    --proto_path=. \
    --go_opt=paths=source_relative \
    --go_opt=Mqueue.proto=github.com/aromancev/confa/internal/proto/queue \
    --go_out=${GO_PATH}/internal/proto/queue \
    --twirp_out=${GO_PATH}/internal/proto/queue

protoc iam.proto \
    --proto_path=. \
    --go_opt=paths=source_relative \
    --go_opt=Miam.proto=github.com/aromancev/confa/internal/proto/iam \
    --go_out=${GO_PATH}/internal/proto/iam \
    --twirp_out=${GO_PATH}/internal/proto/iam

protoc confa.proto \
    --proto_path=. \
    --go_opt=paths=source_relative \
    --go_opt=Mconfa.proto=github.com/aromancev/confa/internal/proto/confa \
    --go_out=${GO_PATH}/internal/proto/confa \
    --twirp_out=${GO_PATH}/internal/proto/confa
