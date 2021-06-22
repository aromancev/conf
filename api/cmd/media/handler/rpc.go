package handler

import (
	"context"
	"time"

	sdk "github.com/pion/ion-sdk-go"

	"github.com/aromancev/confa/internal/media/sfu"
	"github.com/aromancev/confa/internal/platform/trace"
	pmedia "github.com/aromancev/confa/proto/media"
)

type MediaServer struct {
	sfuAddr, mediaDir string
	engine            *sdk.Engine
	producer          Producer
}

func NewMediaServer(sfuAddr, mediaDir string, engine *sdk.Engine, producer Producer) *MediaServer {
	return &MediaServer{sfuAddr: sfuAddr, mediaDir: mediaDir, engine: engine, producer: producer}
}

func (s *MediaServer) SaveTracks(ctx context.Context, request *pmedia.Session) (*pmedia.Reply, error) {
	ctx = trace.New(ctx, request.TraceId)

	saver, err := sfu.NewTrackSaver(ctx, s.engine, s.producer, s.mediaDir, s.sfuAddr, request.SessionId)
	if err != nil {
		return nil, err
	}
	go func() {
		time.Sleep(1 * time.Hour)
		saver.Close()
	}()
	return &pmedia.Reply{}, err
}
