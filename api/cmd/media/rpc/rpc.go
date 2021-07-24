package rpc

import (
	"context"
	"time"

	"github.com/aromancev/confa/internal/media/sfu"
	"github.com/aromancev/confa/internal/platform/trace"
	"github.com/aromancev/confa/proto/media"
	sdk "github.com/pion/ion-sdk-go"
	"github.com/prep/beanstalk"
)

type Producer interface {
	Put(ctx context.Context, tube string, body []byte, params beanstalk.PutParams) (uint64, error)
}

type Handler struct {
	sfuAddr, mediaDir string
	engine            *sdk.Engine
	producer          Producer
}

func NewHandler(sfuAddr, mediaDir string, engine *sdk.Engine, producer Producer) *Handler {
	return &Handler{sfuAddr: sfuAddr, mediaDir: mediaDir, engine: engine, producer: producer}
}

func (s *Handler) SaveTracks(ctx context.Context, request *media.Session) (*media.Reply, error) {
	ctx = trace.New(ctx, request.TraceId)

	saver, err := sfu.NewTrackSaver(ctx, s.engine, s.producer, s.mediaDir, s.sfuAddr, request.SessionId)
	if err != nil {
		return nil, err
	}
	go func() {
		time.Sleep(1 * time.Hour)
		saver.Close()
	}()
	return &media.Reply{}, err
}
