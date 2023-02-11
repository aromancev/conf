package talk

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aromancev/confa/confa"
	"github.com/aromancev/confa/internal/proto/rtc"

	"github.com/google/uuid"
)

type UserRepo interface {
	Create(ctx context.Context, requests ...Talk) ([]Talk, error)
	UpdateOne(ctx context.Context, lookup Lookup, request Update) (Talk, error)
	Fetch(ctx context.Context, lookup Lookup) ([]Talk, error)
	FetchOne(ctx context.Context, lookup Lookup) (Talk, error)
}

type Emitter interface {
	StartRecording(ctx context.Context, talkID, roomID uuid.UUID) error
	StopRecording(ctx context.Context, talkID, roomID uuid.UUID, after time.Duration) error
}

type ConfaRepo interface {
	FetchOne(ctx context.Context, lookup confa.Lookup) (confa.Confa, error)
}

type User struct {
	repo    UserRepo
	emitter Emitter
	confas  ConfaRepo
	rtc     rtc.RTC
}

func NewUser(repo UserRepo, confas ConfaRepo, emitter Emitter, r rtc.RTC) *User {
	return &User{
		repo:    repo,
		emitter: emitter,
		confas:  confas,
		rtc:     r,
	}
}

func (s *User) Create(ctx context.Context, userID uuid.UUID, confaLookup confa.Lookup, request Talk) (Talk, error) {
	request.ID = uuid.New()
	request.Owner = userID
	request.Speaker = userID
	request.State = StateLive
	confaLookup.Owner = userID
	if request.Handle == "" {
		request.Handle = strings.Split(request.ID.String(), "-")[4]
	}
	if err := request.Validate(); err != nil {
		return Talk{}, fmt.Errorf("%w: %s", ErrValidation, err)
	}
	conf, err := s.confas.FetchOne(ctx, confaLookup)
	if err != nil {
		return Talk{}, fmt.Errorf("failed to fetch confa: %w", err)
	}
	request.Confa = conf.ID

	ownerID, _ := userID.MarshalBinary()
	room, err := s.rtc.CreateRoom(ctx, &rtc.Room{
		OwnerId: ownerID,
	})
	if err != nil {
		return Talk{}, fmt.Errorf("failed to create room: %w", err)
	}

	var roomID uuid.UUID
	err = roomID.UnmarshalBinary(room.Id)
	if err != nil {
		return Talk{}, fmt.Errorf("failed to parse room id: %w", err)
	}
	request.Room = roomID

	created, err := s.repo.Create(ctx, request)
	if err != nil {
		return Talk{}, fmt.Errorf("failed to create talk: %w", err)
	}
	return created[0], nil
}

func (s *User) Update(ctx context.Context, userID uuid.UUID, lookup Lookup, request Update) (Talk, error) {
	lookup.Owner = userID
	return s.repo.UpdateOne(ctx, lookup, request)
}

func (s *User) Fetch(ctx context.Context, lookup Lookup) ([]Talk, error) {
	return s.repo.Fetch(ctx, lookup)
}

func (s *User) StartRecording(ctx context.Context, userID uuid.UUID, lookup Lookup) (Talk, error) {
	lookup.Owner = userID
	talk, err := s.repo.FetchOne(ctx, lookup)
	if err != nil {
		return Talk{}, fmt.Errorf("failed to fetch talk: %w", err)
	}
	if talk.State != StateLive {
		return Talk{}, ErrWrongState
	}

	err = s.emitter.StartRecording(ctx, talk.ID, talk.Room)
	if err != nil {
		return Talk{}, fmt.Errorf("failed to emit start recording: %w", err)
	}
	return talk, nil
}

func (s *User) StopRecording(ctx context.Context, userID uuid.UUID, lookup Lookup) (Talk, error) {
	lookup.Owner = userID
	talk, err := s.repo.FetchOne(ctx, lookup)
	if err != nil {
		return Talk{}, fmt.Errorf("failed to fetch talk: %w", err)
	}
	if talk.State != StateRecording {
		return Talk{}, ErrWrongState
	}

	err = s.emitter.StopRecording(ctx, talk.ID, talk.Room, 0)
	if err != nil {
		return Talk{}, fmt.Errorf("failed to emit stop recording: %w", err)
	}
	return talk, nil
}
