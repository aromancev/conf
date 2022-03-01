package web

import (
	"context"
	_ "embed"
	"time"

	"github.com/aromancev/confa/auth"
	"github.com/aromancev/confa/event"
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
	CreatedAt float64
}

type EventLimit struct {
	Count   int32
	Seconds int32
}

type EventRepo interface {
	Fetch(ctx context.Context, lookup event.Lookup) ([]event.Event, error)
}

type Resolver struct {
	publicKey *auth.PublicKey
	events    EventRepo
}

func NewResolver(pk *auth.PublicKey, events EventRepo) *Resolver {
	return &Resolver{
		publicKey: pk,
		events:    events,
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
	if args.Limit.Seconds != 0 {
		delta := time.Second * time.Duration(args.Limit.Seconds)
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
		id, err := uuid.Parse(args.From.ID)
		if err != nil {
			return Events{}, NewResolverError(CodeBadRequest, "Invalid from ID")
		}
		lookup.From = event.From{
			ID:        id,
			CreatedAt: time.UnixMilli(int64(args.From.CreatedAt)),
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
			CreatedAt: float64(lastEvent.CreatedAt.UnixMilli()),
		}
	}
	for i, e := range events {
		res.Items[i] = *NewRoomEvent(e)
	}
	return res, nil
}

//go:embed schema.graphql
var gqlSchema string

const (
	batchLimit = 100
)
