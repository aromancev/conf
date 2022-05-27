package record

import (
	"context"
	"errors"
	"fmt"
	"io"
	"path"
	"sync"
	"time"

	"github.com/aromancev/confa/internal/platform/webrtc/webm"
	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	sdk "github.com/pion/ion-sdk-go"
	"github.com/pion/rtcp"
	"github.com/pion/rtp"
	"github.com/pion/webrtc/v3"
	"github.com/rs/zerolog/log"
)

type Emitter interface {
	ProcessTrack(ctx context.Context, roomID, recordID uuid.UUID, bucket, object string, kind webrtc.RTPCodecType, duration time.Duration) error
}

type Tracker struct {
	rtc     *sdk.RTC
	storage *minio.Client
	emitter Emitter
	bucket  string
	roomID  uuid.UUID

	// Using mutext to protect waitgroup from calling `Wait` before `Add`.
	mutex   sync.Mutex
	writers sync.WaitGroup
	closing bool
}

func NewTracker(ctx context.Context, storage *minio.Client, connector *sdk.Connector, emitter Emitter, bucket string, roomID uuid.UUID) (*Tracker, error) {
	tracker := &Tracker{
		rtc:     sdk.NewRTC(connector),
		storage: storage,
		emitter: emitter,
		bucket:  bucket,
		roomID:  roomID,
	}

	tracker.rtc.OnTrack = func(track *webrtc.TrackRemote, _ *webrtc.RTPReceiver) {
		tracker.mutex.Lock()
		if tracker.closing {
			tracker.mutex.Unlock()
			log.Ctx(ctx).Debug().Str("trackId", track.ID()).Msg("Received track after closing.")
			return
		}
		tracker.mutex.Unlock()

		tracker.writers.Add(1)
		go func() {
			tracker.writeTrack(ctx, track, track.Kind())
			tracker.writers.Done()
		}()
	}

	err := tracker.rtc.Join(roomID.String(), uuid.NewString())
	if err != nil {
		return nil, fmt.Errorf("failed to join room: %w", err)
	}

	return tracker, nil
}

func (t *Tracker) Close() error {
	t.mutex.Lock()
	if t.closing {
		t.mutex.Unlock()
		return nil
	}
	t.closing = true
	t.mutex.Unlock()

	t.rtc.Close()
	t.writers.Wait()
	return nil
}

func (t *Tracker) writeTrack(ctx context.Context, track *webrtc.TrackRemote, kind webrtc.RTPCodecType) {
	type RTPWriteCloser interface {
		Duration() time.Duration
		WriteRTP(packet *rtp.Packet) error
		Close() error
	}

	const pliPeriod = 3 * time.Second
	const minDuration = 1 * time.Second
	recordID := uuid.New()
	objectPath := path.Join(t.roomID.String(), recordID.String())

	log.Ctx(ctx).Info().Str("bucket", t.bucket).Str("objectPath", objectPath).Msg("Started writing track to object.")

	// If any process exits, cancel the context. We're done.
	ctx, cancelCtx := context.WithCancel(ctx)
	defer cancelCtx()
	var wg sync.WaitGroup
	done := func() {
		cancelCtx()
		wg.Done()
	}

	// Sending PLI to receive keyframes at certain intervals.
	wg.Add(1)
	go func() {
		defer done()

		for {
			err := t.rtc.GetSubTransport().GetPeerConnection().WriteRTCP(
				[]rtcp.Packet{
					&rtcp.PictureLossIndication{
						MediaSSRC: uint32(track.SSRC()),
					},
				},
			)
			if err != nil {
				log.Ctx(ctx).Err(err).Msg("Failed to write PLI packet.")
			}
			select {
			case <-ctx.Done():
				return
			case <-time.After(pliPeriod):
			}
		}
	}()

	pipedReader, pipedWriter := io.Pipe()
	defer pipedReader.Close()
	defer pipedWriter.Close()

	// Writing WebM into pipedWriter.
	var duration time.Duration
	wg.Add(1)
	go func() {
		defer done()

		var rtpWriter RTPWriteCloser
		if kind == webrtc.RTPCodecTypeVideo {
			w, err := webm.NewVideoRTPWriter(pipedWriter)
			if err != nil {
				log.Ctx(ctx).Err(err).Msg("Failed to create video writer.")
				return
			}
			rtpWriter = w
		} else {
			w, err := webm.NewAudioRTPWriter(pipedWriter)
			if err != nil {
				log.Ctx(ctx).Err(err).Msg("Failed to create audio writer.")
				return
			}
			rtpWriter = w
		}
		defer func() {
			rtpWriter.Close()
			// We have to record the duration of the track because it's not trivial to get it from a stream file without extensive processing.
			duration = rtpWriter.Duration()
		}()

		for {
			if err := ctx.Err(); err != nil {
				log.Ctx(ctx).Debug().Msg("Context cancelled when writing RTP.")
				return
			}

			packet, _, err := track.ReadRTP()
			switch {
			case errors.Is(err, io.EOF):
				log.Ctx(ctx).Debug().Msg("Track ended when writing RTP.")
				return
			case err != nil:
				log.Ctx(ctx).Err(err).Msg("Failed to read RTP.")
				continue
			}

			if err := rtpWriter.WriteRTP(packet); err != nil {
				log.Ctx(ctx).Warn().Msg("Failed to write RTP packet.")
				continue
			}
		}
	}()

	// Reading WebM into an object from pipedReader.
	wg.Add(1)
	go func() {
		defer done()

		// Even if the parent context is cancelled, object write must finish normally. Never cancel its context.
		_, err := t.storage.PutObject(context.Background(), t.bucket, objectPath, pipedReader, -1, minio.PutObjectOptions{})
		if err != nil {
			log.Ctx(ctx).Err(err).Msg("Failed to write object from track.")
		}
	}()

	wg.Wait()

	// Remve track shorter than allowed duration.
	if duration < minDuration {
		log.Ctx(ctx).Info().Msg("Track durations is less than minimum allowed WebM duration. Removing record.")
		err := t.storage.RemoveObject(context.Background(), t.bucket, objectPath, minio.RemoveObjectOptions{})
		if err != nil {
			log.Ctx(ctx).Err(err).Msg("Failed to remove object from storage.")
		}
		return
	}

	// Even if the parent context is cancelled, object write must finish normally. Never cancel its context.
	err := t.emitter.ProcessTrack(context.Background(), t.roomID, recordID, t.bucket, objectPath, kind, duration)
	if err != nil {
		log.Ctx(ctx).Err(err).Msg("Failed to emit track processing event.")
	}
	log.Ctx(ctx).Info().Str("bucket", t.bucket).Str("objectPath", objectPath).Dur("duration", duration).Msg("Finished writing track to object.")
}
