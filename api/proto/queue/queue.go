package queue

import (
	"google.golang.org/protobuf/proto"
)

//go:generate protoc --proto_path=. --go_opt=Mqueue.proto=github.com/aromancev/confa/proto/queue --go-grpc_opt=Mqueue.proto=github.com/aromancev/confa/proto/queue queue.proto --go_opt=paths=source_relative --go-grpc_opt=paths=source_relative --go-grpc_out=. --go_out=.

const (
	TubeEmail = "email"
	TubeVideo = "video"
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
