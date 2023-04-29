package web

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/aromancev/confa/event"
	"github.com/aromancev/confa/internal/auth"
	"github.com/aromancev/confa/room/record"
	"github.com/google/uuid"
	lkauth "github.com/livekit/protocol/auth"
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
	Items []GraphEvent
	Limit int32
	Next  *EventCursor
}

type GraphEvent struct {
	ID        string
	RoomID    string
	CreatedAt string
	Payload   string
}

type EventLookup struct {
	RoomID string
}

type EventCursor struct {
	ID        *string
	CreatedAt *string
	Asc       *bool
}

type Recording struct {
	Key       string
	RoomID    string
	Status    RecordingStatus
	CreatedAt float64
	StartedAt float64
	StoppedAt *float64
}

type RecordingStatus string

const (
	RecordingStatusRecording  = "RECORDING"
	RecordingStatusProcessing = "PROCESSING"
	RecordingStatusReady      = "READY"
)

type Recordings struct {
	Items    []Recording
	Limit    int32
	NextFrom *RecordingFrom
}

type RecordingLookup struct {
	RoomID string
	Key    *string
}

type LiveKitCredentials struct {
	Key    string
	Secret string
}

type RecordingFrom struct {
	Key string
}

type EventRepo interface {
	Fetch(ctx context.Context, lookup event.Lookup) ([]event.Event, error)
}

type RecordRepo interface {
	Fetch(ctx context.Context, lookup record.Lookup) ([]record.Recording, error)
}

type SFUAccess struct {
	Token string
}

type Resolver struct {
	publicKey    *auth.PublicKey
	livekitCreds LiveKitCredentials
	events       EventRepo
	recordings   RecordRepo
}

func NewResolver(pk *auth.PublicKey, events EventRepo, recordings RecordRepo, lk LiveKitCredentials) *Resolver {
	return &Resolver{
		publicKey:    pk,
		events:       events,
		recordings:   recordings,
		livekitCreds: lk,
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
	Where  EventLookup
	Limit  int32
	Cursor *EventCursor
}) (Events, error) {
	const batchLimit = 3000

	var claims auth.APIClaims
	if err := r.publicKey.Verify(auth.Ctx(ctx).Token(), &claims); err != nil {
		return Events{}, NewResolverError(CodeUnauthorized, "Invalid access token.")
	}

	if args.Limit < 0 || args.Limit > batchLimit {
		args.Limit = batchLimit
	}

	lookup := event.Lookup{
		Limit: int64(args.Limit),
	}
	var err error
	lookup.Room, err = uuid.Parse(args.Where.RoomID)
	if err != nil {
		return Events{}, nil
	}
	if args.Cursor != nil {
		if args.Cursor.CreatedAt != nil {
			createdAt, err := strconv.ParseInt(*args.Cursor.CreatedAt, 10, 64)
			if err != nil {
				return Events{}, NewResolverError(CodeBadRequest, "Invalid filter params.")
			}
			lookup.From.CreatedAt = time.UnixMilli(createdAt)
		}
		if args.Cursor.ID != nil {
			id, err := uuid.Parse(*args.Cursor.ID)
			if err == nil {
				lookup.From.ID = id
			}
		}
		if args.Cursor.Asc != nil && *args.Cursor.Asc {
			lookup.Asc = true
		}
	}

	events, err := r.events.Fetch(ctx, lookup)
	if err != nil {
		log.Ctx(ctx).Err(err).Msg("Failed to fetch events.")
		return Events{}, NewInternalError()
	}
	res := Events{
		Items: make([]GraphEvent, len(events)),
		Limit: int32(lookup.Limit),
	}
	for i, e := range events {
		res.Items[i] = *newGraphEvent(e)
	}
	if len(res.Items) != 0 {
		last := res.Items[len(res.Items)-1]
		res.Next = &EventCursor{
			ID:        &last.ID,
			CreatedAt: &last.CreatedAt,
		}
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

func (r *Resolver) RequestSFUAccess(ctx context.Context, args struct {
	RoomID string
}) (SFUAccess, error) {
	const validFor = 2 * time.Hour

	var claims auth.APIClaims
	if err := r.publicKey.Verify(auth.Ctx(ctx).Token(), &claims); err != nil {
		return SFUAccess{}, NewResolverError(CodeUnauthorized, "Invalid access token.")
	}

	access := lkauth.NewAccessToken(r.livekitCreds.Key, r.livekitCreds.Secret)
	access.AddGrant(&lkauth.VideoGrant{
		RoomJoin: true,
		Room:     args.RoomID,
	})
	access.SetIdentity(claims.UserID.String())
	access.SetValidFor(validFor)
	token, err := access.ToJWT()
	if err != nil {
		return SFUAccess{}, NewInternalError()
	}
	return SFUAccess{
		Token: token,
	}, nil
}

func newRecording(rec record.Recording) Recording {
	api := Recording{
		Key:       rec.Key,
		RoomID:    rec.Room.String(),
		Status:    RecordingStatusRecording,
		CreatedAt: float64(rec.CreatedAt.UTC().UnixMilli()),
		StartedAt: float64(rec.StartedAt.UTC().UnixMilli()),
	}
	if !rec.StoppedAt.IsZero() {
		t := float64(rec.StoppedAt.UTC().UnixMilli())
		api.StoppedAt = &t
		if rec.IsReady() {
			api.Status = RecordingStatusReady
		} else {
			api.Status = RecordingStatusProcessing
		}
	}
	return api
}

func newGraphEvent(ev event.Event) *GraphEvent {
	roomEvent := NewRoomEvent(ev)
	pl, _ := json.Marshal(roomEvent.Payload)
	return &GraphEvent{
		ID:        roomEvent.ID,
		RoomID:    roomEvent.RoomID,
		CreatedAt: fmt.Sprint(ev.CreatedAt.UnixMilli()),
		Payload:   string(pl),
	}
}

//go:embed schema.graphql
var gqlSchema string
