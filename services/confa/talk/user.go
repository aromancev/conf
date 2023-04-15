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
	Delete(ctx context.Context, lookup Lookup) (UpdateResult, error)
}

type Emitter interface {
	StartRecording(ctx context.Context, talkID, roomID uuid.UUID) error
	StopRecording(ctx context.Context, talkID, roomID uuid.UUID, after time.Duration) error
}

type ConfaRepo interface {
	FetchOne(ctx context.Context, lookup confa.Lookup) (confa.Confa, error)
	Delete(ctx context.Context, lookup confa.Lookup) (confa.UpdateResult, error)
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

func (u *User) Create(ctx context.Context, userID uuid.UUID, confaLookup confa.Lookup, request Talk) (Talk, error) {
	request.ID = uuid.New()
	request.Owner = userID
	request.Speaker = userID
	request.State = StateLive
	confaLookup.Owner = userID
	if request.Handle == "" {
		request.Handle = strings.Split(request.ID.String(), "-")[4]
	}
	if err := request.Validate(); err != nil {
		return Talk{}, fmt.Errorf("%w: %u", ErrValidation, err)
	}
	conf, err := u.confas.FetchOne(ctx, confaLookup)
	if err != nil {
		return Talk{}, fmt.Errorf("failed to fetch confa: %w", err)
	}
	request.Confa = conf.ID

	ownerID, _ := userID.MarshalBinary()
	room, err := u.rtc.CreateRoom(ctx, &rtc.Room{
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

	created, err := u.repo.Create(ctx, request)
	if err != nil {
		return Talk{}, fmt.Errorf("failed to create talk: %w", err)
	}
	return created[0], nil
}

func (u *User) Update(ctx context.Context, userID uuid.UUID, lookup Lookup, request Update) (Talk, error) {
	lookup.Owner = userID
	return u.repo.UpdateOne(ctx, lookup, request)
}

func (u *User) Fetch(ctx context.Context, lookup Lookup) ([]Talk, error) {
	return u.repo.Fetch(ctx, lookup)
}

func (u *User) Delete(ctx context.Context, userID uuid.UUID, lookup Lookup) (UpdateResult, error) {
	lookup.Owner = userID

	res, err := u.repo.Delete(ctx, lookup)
	if err != nil {
		return UpdateResult{}, err
	}
	return res, nil
}

func (u *User) DeleteConfa(ctx context.Context, userID uuid.UUID, lookup confa.Lookup) (UpdateResult, error) {
	lookup.Owner = userID

	cnf, err := u.confas.FetchOne(ctx, lookup)
	if err != nil {
		return UpdateResult{}, err
	}
	res, err := u.repo.Delete(ctx, Lookup{
		Confa: cnf.ID,
	})
	if err != nil {
		return UpdateResult{}, err
	}
	_, err = u.confas.Delete(ctx, lookup)
	if err != nil {
		return UpdateResult{}, err
	}
	return res, nil
}

func (u *User) StartRecording(ctx context.Context, userID uuid.UUID, lookup Lookup) (Talk, error) {
	lookup.Owner = userID
	talk, err := u.repo.FetchOne(ctx, lookup)
	if err != nil {
		return Talk{}, fmt.Errorf("failed to fetch talk: %w", err)
	}
	if talk.State != StateLive {
		return Talk{}, ErrWrongState
	}

	err = u.emitter.StartRecording(ctx, talk.ID, talk.Room)
	if err != nil {
		return Talk{}, fmt.Errorf("failed to emit start recording: %w", err)
	}
	return talk, nil
}

func (u *User) StopRecording(ctx context.Context, userID uuid.UUID, lookup Lookup) (Talk, error) {
	lookup.Owner = userID
	talk, err := u.repo.FetchOne(ctx, lookup)
	if err != nil {
		return Talk{}, fmt.Errorf("failed to fetch talk: %w", err)
	}
	if talk.State != StateRecording {
		return Talk{}, ErrWrongState
	}

	err = u.emitter.StopRecording(ctx, talk.ID, talk.Room, 0)
	if err != nil {
		return Talk{}, fmt.Errorf("failed to emit stop recording: %w", err)
	}
	return talk, nil
}
