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

type Kind int

const (
	KindVideo Kind = iota
	KindAudio
)

type Record struct {
	RoomID      uuid.UUID
	RecordingID uuid.UUID
	RecordID    uuid.UUID
	TrackID     string
	Kind        Kind
	Bucket      string
	Object      string
	Duration    time.Duration
}

type Emitter interface {
	RecordStarted(ctx context.Context, record Record) error
	RecordFinished(ctx context.Context, record Record) error
}

type Tracker struct {
	rtc                 *sdk.RTC
	storage             *minio.Client
	emitter             Emitter
	bucket              string
	roomID, recordingID uuid.UUID

	// Using mutext to protect waitgroup from calling `Wait` before `Add`.
	mutex   sync.Mutex
	writers sync.WaitGroup
	closed  bool
}

func NewTracker(ctx context.Context, storage *minio.Client, connector *sdk.Connector, emitter Emitter, bucket string, roomID, recordingID uuid.UUID) (*Tracker, error) {
	rtcClient, err := sdk.NewRTC(connector)
	if err != nil {
		return nil, err
	}
	tracker := &Tracker{
		rtc:         rtcClient,
		storage:     storage,
		emitter:     emitter,
		bucket:      bucket,
		roomID:      roomID,
		recordingID: recordingID,
	}

	tracker.rtc.OnTrack = func(track *webrtc.TrackRemote, _ *webrtc.RTPReceiver) {
		tracker.mutex.Lock()
		if tracker.closed {
			tracker.mutex.Unlock()
			log.Ctx(ctx).Debug().Str("trackId", track.ID()).Msg("Received track after closing.")
			return
		}
		tracker.mutex.Unlock()

		tracker.writers.Add(1)
		go func() {
			if track.Kind() == webrtc.RTPCodecTypeAudio {
				tracker.writeTrack(ctx, track, KindAudio)
			} else {
				tracker.writeTrack(ctx, track, KindVideo)
			}
			tracker.writers.Done()
		}()
	}
	// Empty handler to mute sdk error about missing handler.
	tracker.rtc.OnTrackEvent = func(event sdk.TrackEvent) {}

	err = rtcClient.Join(roomID.String(), uuid.NewString())
	if err != nil {
		return nil, fmt.Errorf("failed to join room: %w", err)
	}

	return tracker, nil
}

func (t *Tracker) Close(ctx context.Context) error {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	if t.closed {
		return nil
	}
	t.rtc.Close()
	t.writers.Wait()
	t.closed = true
	return nil
}

func (t *Tracker) writeTrack(ctx context.Context, track *webrtc.TrackRemote, kind Kind) {
	type RTPWriteCloser interface {
		Duration() time.Duration
		WriteRTP(packet *rtp.Packet) error
		Close() error
	}

	const pliPeriod = 3 * time.Second
	const minDuration = 6 * time.Second
	const rtpMaxLate = 300
	recordID := uuid.New()
	objectPath := path.Join(t.roomID.String(), recordID.String())

	log.Ctx(ctx).Info().Str("bucket", t.bucket).Str("objectPath", objectPath).Msg("Started writing track to object.")

	watchdogCtx, cancelWatchdog := context.WithCancel(ctx)
	defer cancelWatchdog()
	var wg sync.WaitGroup

	// Sending PLI to receive keyframes at certain intervals.
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer cancelWatchdog()

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
			log.Ctx(ctx).Debug().Msg("Sent PLI.")
			select {
			case <-watchdogCtx.Done():
				return
			case <-time.After(pliPeriod):
			}
		}
	}()

	pipedReader, pipedWriter := io.Pipe()
	defer pipedReader.Close()
	defer pipedWriter.Close()

	record := Record{
		RoomID:      t.roomID,
		RecordingID: t.recordingID,
		RecordID:    recordID,
		TrackID:     track.ID(),
		Kind:        kind,
		Bucket:      t.bucket,
		Object:      objectPath,
	}
	var recordStarted bool
	wg.Add(1)
	// Writing WebM into pipedWriter.
	go func() {
		defer wg.Done()
		defer cancelWatchdog()

		var rtpWriter RTPWriteCloser
		if kind == KindVideo {
			w, err := webm.NewVideoRTPWriter(pipedWriter, rtpMaxLate)
			if err != nil {
				log.Ctx(ctx).Err(err).Msg("Failed to create video writer.")
				return
			}
			rtpWriter = w
			log.Ctx(ctx).Debug().Msg("Created video writer.")
		} else {
			w, err := webm.NewAudioRTPWriter(pipedWriter)
			if err != nil {
				log.Ctx(ctx).Err(err).Msg("Failed to create audio writer.")
				return
			}
			rtpWriter = w
			log.Ctx(ctx).Debug().Msg("Created audio writer.")
		}
		defer func() {
			rtpWriter.Close()
		}()

		for {
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
			record.Duration = rtpWriter.Duration()
			// Emitting a track event only after the minimum track duration has beed recorded.
			// Not emitting immediately to avoid creating an event for invalid track.
			if !recordStarted && record.Duration >= minDuration {
				err := t.emitter.RecordStarted(ctx, record)
				if err != nil {
					log.Ctx(ctx).Err(err).Msg("Failed to emit record started event.")
				}
				recordStarted = true
			}
		}
	}()

	// Reading WebM into an object from pipedReader.
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer cancelWatchdog()

		_, err := t.storage.PutObject(ctx, t.bucket, objectPath, pipedReader, -1, minio.PutObjectOptions{})
		if err != nil {
			log.Ctx(ctx).Err(err).Msg("Failed to write object from track.")
		}
	}()

	wg.Wait()

	if !recordStarted {
		log.Ctx(ctx).Info().Str("duration", record.Duration.String()).Msg("Track durations is less than minimum allowed WebM duration. Removing record.")
		err := t.storage.RemoveObject(ctx, t.bucket, objectPath, minio.RemoveObjectOptions{})
		if err != nil {
			log.Ctx(ctx).Err(err).Msg("Failed to remove object from storage.")
		}
		return
	}

	err := t.emitter.RecordFinished(ctx, record)
	if err != nil {
		log.Ctx(ctx).Err(err).Msg("Failed to emit record finished event.")
	}
	log.Ctx(ctx).Info().Str("bucket", t.bucket).Str("objectPath", objectPath).Str("duration", record.Duration.String()).Msg("Finished writing track to object.")
}
