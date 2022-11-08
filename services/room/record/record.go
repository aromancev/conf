package record

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrValidation      = errors.New("validation error")
	ErrNotFound        = errors.New("not found")
	ErrAmbigiousLookup = errors.New("ambigious lookup")
	ErrDuplicateEntry  = errors.New("record already exists")
)

type Recording struct {
	ID        uuid.UUID `bson:"_id"`
	Room      uuid.UUID `bson:"roomId"`
	Key       string    `bson:"key"`
	Active    bool      `bson:"active,omitempty"`
	Records   Records   `bson:"records,omitempty"`
	StartedAt time.Time `bson:"startedAt"`
	StoppedAt time.Time `bson:"stoppedAt,omitempty"`
	CreatedAt time.Time `bson:"createdAt"`
}

func (r Recording) Validate() error {
	if r.ID == uuid.Nil {
		return errors.New("id should not be empty")
	}
	if r.Room == uuid.Nil {
		return errors.New("room should not be empty")
	}
	return nil
}

func (r Recording) IsReady() bool {
	if r.StoppedAt.IsZero() {
		return false
	}
	return r.Records.IsReady()
}

type Records struct {
	RecordingStarted   []uuid.UUID `bson:"recordingStarted,omitempty"`
	RecordingFinished  []uuid.UUID `bson:"recordingFinished,omitempty"`
	ProcessingStarted  []uuid.UUID `bson:"processingStarted,omitempty"`
	ProcessingFinished []uuid.UUID `bson:"processingFinished,omitempty"`
}

func (r Records) IsZero() bool {
	if len(r.RecordingStarted) != 0 {
		return false
	}
	if len(r.RecordingFinished) != 0 {
		return false
	}
	if len(r.ProcessingStarted) != 0 {
		return false
	}
	if len(r.ProcessingFinished) != 0 {
		return false
	}
	return true
}

func (r Records) IsReady() bool {
	if len(r.RecordingStarted) == 0 {
		return true
	}
	finished := make(map[uuid.UUID]struct{}, len(r.ProcessingFinished))
	for _, id := range r.ProcessingFinished {
		finished[id] = struct{}{}
	}
	for _, id := range r.RecordingStarted {
		if _, ok := finished[id]; !ok {
			return false
		}
	}
	return true
}

type Lookup struct {
	ID      uuid.UUID
	Room    uuid.UUID
	Key     string
	Limit   int64
	FromKey string
}
