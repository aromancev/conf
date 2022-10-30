#!/bin/bash -e

protoc rtc.proto \
    --proto_path=. \
    --go_opt=paths=source_relative \
    --go_opt=Mrtc.proto=github.com/aromancev/confa/rtc \
    --go_out=rtc \
    --twirp_out=rtc

protoc queue.proto \
    --proto_path=. \
    --go_opt=paths=source_relative \
    --go_opt=Mqueue.proto=github.com/aromancev/confa/queue \
    --go_out=queue \
    --twirp_out=queue

protoc sender.proto \
    --proto_path=. \
    --go_opt=paths=source_relative \
    --go_opt=Msender.proto=github.com/aromancev/confa/sender \
    --go_out=sender \
    --twirp_out=sender

protoc confa.proto \
    --proto_path=. \
    --go_opt=paths=source_relative \
    --go_opt=Mconfa.proto=github.com/aromancev/confa/confa \
    --go_out=confa \
    --twirp_out=confa

protoc tracker.proto \
    --proto_path=. \
    --go_opt=paths=source_relative \
    --go_opt=Mtracker.proto=github.com/aromancev/confa/tracker \
    --go_out=tracker \
    --twirp_out=tracker

protoc avp.proto \
    --proto_path=. \
    --go_opt=paths=source_relative \
    --go_opt=Mavp.proto=github.com/aromancev/confa/avp \
    --go_out=avp \
    --twirp_out=avp
