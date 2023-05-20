package record

import (
	"context"
	"errors"
	"io"
	"path"
	"sync"
	"time"

	"github.com/aromancev/confa/internal/platform/webrtc/webm"
	"github.com/aromancev/confa/internal/proto/rtc"
	"github.com/google/uuid"
	"github.com/livekit/protocol/livekit"
	lksdk "github.com/livekit/server-sdk-go"
	"github.com/minio/minio-go/v7"
	"github.com/pion/rtp"
	"github.com/pion/webrtc/v3"
	"github.com/rs/zerolog/log"
)

type LivekitCredentials struct {
	URL    string
	Key    string
	Secret string
}

type Record struct {
	RoomID      uuid.UUID
	RecordingID uuid.UUID
	RecordID    uuid.UUID
	Kind        rtc.TrackKind
	Source      rtc.TrackSource
	Bucket      string
	Object      string
	Duration    time.Duration
	CreatedAt   time.Time
}

type Emitter interface {
	RecordStarted(ctx context.Context, record Record) error
	RecordFinished(ctx context.Context, record Record) error
}

type Tracker struct {
	room                *lksdk.Room
	storage             *minio.Client
	emitter             Emitter
	bucket              string
	roomID, recordingID uuid.UUID

	// Using mutext to protect waitgroup from calling `Wait` before `Add`.
	mutex   sync.Mutex
	writers sync.WaitGroup
	closed  bool
}

func NewTracker(ctx context.Context, storage *minio.Client, emitter Emitter, creds LivekitCredentials, bucket string, roomID, recordingID uuid.UUID) (*Tracker, error) {
	tracker := &Tracker{
		storage:     storage,
		emitter:     emitter,
		bucket:      bucket,
		roomID:      roomID,
		recordingID: recordingID,
	}

	room, err := lksdk.ConnectToRoom(
		creds.URL,
		lksdk.ConnectInfo{
			APIKey:              creds.Key,
			APISecret:           creds.Secret,
			RoomName:            roomID.String(),
			ParticipantIdentity: uuid.NewString(),
		},
		&lksdk.RoomCallback{
			ParticipantCallback: lksdk.ParticipantCallback{
				OnTrackSubscribed: func(track *webrtc.TrackRemote, pub *lksdk.RemoteTrackPublication, rp *lksdk.RemoteParticipant) {
					startAt := time.Now()

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
							tracker.writeTrack(ctx, track, rp.WritePLI, rtc.TrackKind_AUDIO, newSource(pub.Source()), startAt)
						} else {
							tracker.writeTrack(ctx, track, rp.WritePLI, rtc.TrackKind_VIDEO, newSource(pub.Source()), startAt)
						}
						tracker.writers.Done()
					}()
				},
			},
		},
	)
	if err != nil {
		return &Tracker{}, err
	}
	tracker.room = room
	return tracker, nil
}

func (t *Tracker) Close(ctx context.Context) error {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	if t.closed {
		return nil
	}
	t.room.Disconnect()
	t.writers.Wait()
	t.closed = true
	return nil
}

func (t *Tracker) writeTrack(ctx context.Context, track *webrtc.TrackRemote, pli lksdk.PLIWriter, kind rtc.TrackKind, source rtc.TrackSource, startAt time.Time) {
	type RTPWriteCloser interface {
		Duration() time.Duration
		WriteRTP(packet *rtp.Packet) error
		Close() error
	}

	const pliPeriod = 3 * time.Second
	const minDuration = 6 * time.Second
	const rtpMaxLate = 300 // should be 1000 for 2s of fHD video and 200 for 4s audio.
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
			pli(track.SSRC())
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
		Kind:        kind,
		Source:      source,
		Bucket:      t.bucket,
		Object:      objectPath,
		CreatedAt:   startAt,
	}
	var recordStarted bool
	wg.Add(1)
	// Writing WebM into pipedWriter.
	go func() {
		defer wg.Done()
		defer cancelWatchdog()

		var rtpWriter RTPWriteCloser
		if kind == rtc.TrackKind_VIDEO {
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

func newSource(lk livekit.TrackSource) rtc.TrackSource {
	switch lk {
	case livekit.TrackSource_CAMERA:
		return rtc.TrackSource_CAMERA
	case livekit.TrackSource_MICROPHONE:
		return rtc.TrackSource_MICROPHONE
	case livekit.TrackSource_SCREEN_SHARE:
		return rtc.TrackSource_SCREEN
	case livekit.TrackSource_SCREEN_SHARE_AUDIO:
		return rtc.TrackSource_SCREEN_AUDIO
	default:
		return rtc.TrackSource_UNKNOWN
	}
}
