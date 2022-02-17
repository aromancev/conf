package sfu

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
	"time"

	"github.com/google/uuid"
	sdk "github.com/pion/ion-sdk-go"
	"github.com/pion/rtcp"
	"github.com/pion/rtp"
	"github.com/pion/webrtc/v3"
	"github.com/pion/webrtc/v3/pkg/media/ivfwriter"
	"github.com/pion/webrtc/v3/pkg/media/oggwriter"
	"github.com/prep/beanstalk"
	"github.com/rs/zerolog/log"

	"github.com/aromancev/confa/internal/platform/trace"
	"github.com/aromancev/confa/internal/proto/queue"
)

type Producer interface {
	Put(ctx context.Context, tube string, body []byte, params beanstalk.PutParams) (uint64, error)
}

type TrackSaver struct {
	mediaDir string
	producer Producer
	client   *sdk.Client
	cancel   func()
}

func NewTrackSaver(ctx context.Context, engine *sdk.Engine, producer Producer, mediaDir, sfuAddr, session string) (*TrackSaver, error) {
	client, err := sdk.NewClient(engine, sfuAddr, uuid.NewString())
	if err != nil {
		return nil, err
	}

	saveCtx, cancel := context.WithCancel(log.Ctx(ctx).WithContext(context.Background()))

	s := &TrackSaver{
		mediaDir: mediaDir,
		producer: producer,
		client:   client,
		cancel:   cancel,
	}

	client.OnTrack = func(track *webrtc.TrackRemote, _ *webrtc.RTPReceiver) {
		go s.saveTrack(saveCtx, track)
	}

	err = client.Join(session)
	if err != nil {
		return nil, fmt.Errorf("failed to join session: %w", err)
	}

	return s, nil
}

func (s *TrackSaver) Close() {
	s.cancel()
	s.client.Close()
}

func (s *TrackSaver) saveTrack(ctx context.Context, track *webrtc.TrackRemote) {
	mediaID := uuid.NewString()

	l := log.Ctx(ctx).With().
		Str("trackId", track.ID()).
		Str("mediaId", mediaID).Logger()
	ctx = l.WithContext(ctx)

	var err error
	codec := trackCodec(track)
	switch codec {
	case codecVideo:
		err = s.saveVideo(ctx, track, mediaID)
	case codecAudio:
		err = s.saveAudio(ctx, track, mediaID)
	default:
		log.Ctx(ctx).Error().Str("codec", codec).Msg("Received track with unsupported codec.")
		return
	}
	if err != nil {
		log.Ctx(ctx).Err(err).Msg("Failed to save track.")
		return
	}

	body, err := queue.Marshal(&queue.VideoJob{MediaId: mediaID}, trace.ID(ctx))
	if err != nil {
		log.Ctx(ctx).Err(err).Msg("Failed to marshal video job.")
		return
	}
	_, err = s.producer.Put(context.Background(), queue.TubeVideo, body, beanstalk.PutParams{TTR: 10 * time.Minute})
	if err != nil {
		log.Ctx(ctx).Err(err).Msg("Failed to put video job.")
		return
	}

	log.Ctx(ctx).Info().Msg("Track saved.")
}

func (s *TrackSaver) writePLI(ctx context.Context, track *webrtc.TrackRemote) {
	for {
		if err := ctx.Err(); err != nil {
			log.Ctx(ctx).Info().Msg("Stopped writing PLI.")
			return
		}

		err := s.client.GetSubTransport().GetPeerConnection().WriteRTCP(
			[]rtcp.Packet{
				&rtcp.PictureLossIndication{
					MediaSSRC: uint32(track.SSRC()),
				},
			},
		)
		if err != nil {
			log.Ctx(ctx).Err(err).Msg("Failed to write PLI packet.")
		}
		time.Sleep(pliPeriod)
	}
}

func (s *TrackSaver) saveVideo(ctx context.Context, track *webrtc.TrackRemote, id string) error {
	err := os.MkdirAll(path.Join(s.mediaDir, id), 0777)
	if err != nil {
		return fmt.Errorf("failed to create dir: %w", err)
	}
	file, err := ivfwriter.New(path.Join(s.mediaDir, id, "raw.ivf"))
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	go s.writePLI(ctx, track)

	log.Ctx(ctx).Info().Msg("Started saving video track.")
	buf := make([]byte, buffSize)
	for {
		if err := ctx.Err(); err != nil {
			log.Ctx(ctx).Info().Msg("Stopped saving video.")
			return nil
		}

		n, _, err := track.Read(buf)
		switch {
		case errors.Is(err, io.EOF):
			log.Ctx(ctx).Info().Msg("Stopped saving video.")
			return nil
		case err != nil:
			log.Ctx(ctx).Info().Msg("Video saving error.")
			return nil
		}

		var packet rtp.Packet
		if err = packet.Unmarshal(buf[:n]); err != nil {
			log.Ctx(ctx).Err(err).Msg("Failed to unmarshal RTP packet.")
			return nil
		}

		if len(packet.Payload) < 4 {
			log.Ctx(ctx).Warn().Msg("Received RTP packet is too small for IVF header. Skipping.")
			continue
		}

		if err := file.WriteRTP(&packet); err != nil {
			log.Ctx(ctx).Warn().Msg("Failed to save RTP packet.")
		}
	}
}

func (s *TrackSaver) saveAudio(ctx context.Context, track *webrtc.TrackRemote, id string) error {
	err := os.MkdirAll(path.Join(s.mediaDir, id), 0777)
	if err != nil {
		return fmt.Errorf("failed to create dir: %w", err)
	}
	file, err := oggwriter.New(
		path.Join(s.mediaDir, id, "raw.ogg"),
		48000,
		2,
	)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	go s.writePLI(ctx, track)

	log.Info().Msg("Started saving audio track.")

	buf := make([]byte, buffSize)
	for {
		if err := ctx.Err(); err != nil {
			log.Ctx(ctx).Info().Msg("Stopped saving audio.")
			return nil
		}
		n, _, err := track.Read(buf)
		switch {
		case errors.Is(err, io.EOF):
			log.Ctx(ctx).Info().Msg("Stopped saving audio.")
			return nil
		case err != nil:
			log.Ctx(ctx).Info().Msg("Audio saving error.")
			return nil
		}

		var packet rtp.Packet
		if err = packet.Unmarshal(buf[:n]); err != nil {
			log.Ctx(ctx).Err(err).Msg("Failed to unmarshal RTP packet.")
			return nil
		}

		if err := file.WriteRTP(&packet); err != nil {
			log.Ctx(ctx).Warn().Msg("Failed to save RTP packet.")
		}
	}
}

func trackCodec(track *webrtc.TrackRemote) string {
	parts := strings.Split(track.Codec().RTPCodecCapability.MimeType, "/")
	if len(parts) < 2 {
		return ""
	}
	return strings.ToLower(parts[1])
}

const (
	codecVideo = "vp8"
	codecAudio = "opus"
	pliPeriod  = 3 * time.Second
	buffSize   = 4096
)
