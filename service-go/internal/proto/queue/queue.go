package queue

import (
	"google.golang.org/protobuf/proto"
)

const (
	TubeEmail = "email"
	TubeVideo = "video"
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
