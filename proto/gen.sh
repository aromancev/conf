#!/bin/bash -e

GO_PATH=../service-go/internal/proto

# Clearing old files.
find ${GO_PATH} -name "*.pb.go" -type f -delete
find ${GO_PATH} -name "*.twirp.go" -type f -delete

# Generating new files.
protoc rtc.proto --go_opt=Mrtc.proto=github.com/aromancev/confa/internal/proto/rtc --go_out=${GO_PATH}/rtc --twirp_out=${GO_PATH}/rtc --proto_path=. --go_opt=paths=source_relative
protoc queue.proto --go_opt=Mqueue.proto=github.com/aromancev/confa/internal/proto/queue --go_out=${GO_PATH}/queue --twirp_out=${GO_PATH}/queue --proto_path=. --go_opt=paths=source_relative
 