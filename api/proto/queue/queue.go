package queue

import (
	"google.golang.org/protobuf/proto"
)

//go:generate protoc  queue.proto --go_opt=Mqueue.proto=github.com/aromancev/confa/proto/queue --proto_path=. --go_opt=paths=source_relative --go_out=. --twirp_out=.

const (
	TubeEmail = "email"
	TubeVideo = "video"
	TubeImage = "image"
	TubeEvent = "event"
)

func Marshal(payload proto.Message, trace string) ([]byte, error) {
	pl, err := proto.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return proto.Marshal(&Job{
		Payload: pl,
		TraceId: trace,
	})
}

func Unmarshal(body []byte) (*Job, error) {
	var job Job
	err := proto.Unmarshal(body, &job)
	if err != nil {
		return nil, err
	}
	return &job, nil
}
