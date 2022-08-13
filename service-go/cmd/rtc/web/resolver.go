package web

import (
	"context"
	_ "embed"
	"strconv"
	"time"

	"github.com/aromancev/confa/auth"
	"github.com/aromancev/confa/event"
	"github.com/aromancev/confa/room/record"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

type Code string

const (
	CodeBadRequest       = "BAD_REQUEST"
	CodeUnauthorized     = "UNAUTHORIZED"
	CodeDuplicateEntry   = "DUPLICATE_ENTRY"
	CodeNotFound         = "NOT_FOUND"
	CodePermissionDenied = "PERMISSION_DENIED"
	CodeUnknown          = "UNKNOWN_CODE"
)

const (
	OrderAsc  string = "ASC"
	OrderDesc string = "DESC"
)

type ResolverError struct {
	message    string
	extensions map[string]interface{}
}

func (e ResolverError) Error() string {
	return e.message
}

func (e ResolverError) Extensions() map[string]interface{} {
	return e.extensions
}

func NewResolverError(code Code, message string) ResolverError {
	return ResolverError{
		message: message,
		extensions: map[string]interface{}{
			"code": code,
		},
	}
}

func NewInternalError() ResolverError {
	return NewResolverError("internal system error", CodeUnknown)
}

type Service struct {
	Name    string
	Version string
	Schema  string
}

type Events struct {
	Items    []RoomEvent
	Limit    int32
	NextFrom *EventFrom
}

type EventLookup struct {
	RoomID string
}

type EventFrom struct {
	ID        string
	CreatedAt string
}

type EventLimit struct {
	Count   int32
	Seconds *int32
}

type Recording struct {
	Key       string
	RoomID    string
	CreatedAt float64
	StartedAt float64
	StoppedAt *float64
}

type Recordings struct {
	Items    []Recording
	Limit    int32
	NextFrom *RecordingFrom
}

type RecordingLookup struct {
	RoomID string
	Key    *string
}

type RecordingFrom struct {
	Key string
}

type EventRepo interface {
	Fetch(ctx context.Context, lookup event.Lookup) ([]event.Event, error)
}

type RecordRepo interface {
	Fetch(ctx context.Context, lookup record.Lookup) ([]record.Record, error)
}

type Resolver struct {
	publicKey  *auth.PublicKey
	events     EventRepo
	recordings RecordRepo
}

func NewResolver(pk *auth.PublicKey, events EventRepo, recordings RecordRepo) *Resolver {
	return &Resolver{
		publicKey:  pk,
		events:     events,
		recordings: recordings,
	}
}

func (r *Resolver) Service(_ context.Context) Service {
	return Service{
		Name:    "rtc",
		Version: "0.1.0",
		Schema:  gqlSchema,
	}
}

func (r *Resolver) Events(ctx context.Context, args struct {
	Where EventLookup
	Limit EventLimit
	From  *EventFrom
	Order *string
}) (Events, error) {
	const batchLimit = 3000

	var claims auth.APIClaims
	if err := r.publicKey.Verify(auth.Ctx(ctx).Token(), &claims); err != nil {
		return Events{}, NewResolverError(CodeUnauthorized, "Invalid access token.")
	}

	if args.Limit.Count < 0 || args.Limit.Count > batchLimit {
		args.Limit.Count = batchLimit
	}

	lookup := event.Lookup{
		Limit: int64(args.Limit.Count),
		Asc:   args.Order != nil && *args.Order == OrderAsc,
	}
	if args.Limit.Seconds != nil {
		delta := time.Second * time.Duration(*args.Limit.Seconds)
		if lookup.Asc {
			lookup.From.CreatedAt = time.Now().UTC().Add(delta)
		} else {
			lookup.From.CreatedAt = time.Now().UTC().Add(-delta)
		}
	}
	var err error
	lookup.Room, err = uuid.Parse(args.Where.RoomID)
	if err != nil {
		return Events{}, nil
	}
	if args.From != nil {
		createdAt, err := strconv.ParseInt(args.From.CreatedAt, 10, 64)
		if err != nil {
			return Events{}, NewResolverError(CodeBadRequest, "Invalid from.createdAt")
		}
		lookup.From = event.From{
			CreatedAt: time.UnixMilli(createdAt),
		}
		if args.From.ID != "" {
			id, err := uuid.Parse(args.From.ID)
			if err == nil {
				lookup.From.ID = id
			}
		}
	}

	events, err := r.events.Fetch(ctx, lookup)
	if err != nil {
		log.Ctx(ctx).Err(err).Msg("Failed to fetch events.")
		return Events{}, NewInternalError()
	}
	res := Events{
		Items: make([]RoomEvent, len(events)),
		Limit: int32(lookup.Limit),
	}
	if len(events) != 0 {
		lastEvent := events[len(events)-1]
		res.NextFrom = &EventFrom{
			ID:        lastEvent.ID.String(),
			CreatedAt: strconv.FormatInt(lastEvent.CreatedAt.UnixMilli(), 10),
		}
	}
	for i, e := range events {
		res.Items[i] = *NewRoomEvent(e)
	}
	return res, nil
}

func (r *Resolver) Recordings(ctx context.Context, args struct {
	Where RecordingLookup
	Limit int32
	From  *RecordingFrom
}) (Recordings, error) {
	const batchLimit = 100

	var claims auth.APIClaims
	if err := r.publicKey.Verify(auth.Ctx(ctx).Token(), &claims); err != nil {
		return Recordings{}, NewResolverError(CodeUnauthorized, "Invalid access token.")
	}

	if args.Limit < 0 || args.Limit > batchLimit {
		args.Limit = batchLimit
	}

	lookup := record.Lookup{
		Limit: int64(args.Limit),
	}
	var err error
	lookup.Room, err = uuid.Parse(args.Where.RoomID)
	if err != nil {
		return Recordings{Limit: args.Limit}, nil
	}
	if args.Where.Key != nil {
		lookup.Key = *args.Where.Key
	}
	if args.From != nil {
		lookup.FromKey = args.From.Key
	}

	recordings, err := r.recordings.Fetch(ctx, lookup)
	if err != nil {
		log.Ctx(ctx).Err(err).Msg("Failed to fetch events.")
		return Recordings{}, NewInternalError()
	}
	res := Recordings{
		Items: make([]Recording, len(recordings)),
		Limit: int32(lookup.Limit),
	}
	if len(recordings) != 0 {
		last := recordings[len(recordings)-1]
		res.NextFrom = &RecordingFrom{
			Key: last.Key,
		}
	}
	for i, r := range recordings {
		res.Items[i] = newRecording(r)
	}
	return res, nil
}

func newRecording(rec record.Record) Recording {
	api := Recording{
		Key:       rec.Key,
		RoomID:    rec.Room.String(),
		CreatedAt: float64(rec.CreatedAt.UTC().UnixMilli()),
		StartedAt: float64(rec.StartedAt.UTC().UnixMilli()),
	}
	if !rec.StoppedAt.IsZero() {
		t := float64(rec.StoppedAt.UTC().UnixMilli())
		api.StoppedAt = &t
	}
	return api
}

//go:embed schema.graphql
var gqlSchema string
