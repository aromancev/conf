package web

import (
	"context"
	_ "embed"
	"encoding/json"
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

type Event struct {
	ID        string
	OwnerID   string
	RoomID    string
	CreatedAt string
	Payload   EventPayload
}

type EventPayload struct {
	Type    string
	Payload string
}

type Events struct {
	Items    []Event
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
		Schema:  schema,
	}
}

func (r *Resolver) Events(ctx context.Context, args struct {
	Where Event
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
		created, err := time.Parse(time.RFC3339, args.From.CreatedAt)
		if err != nil {
			return Events{}, NewResolverError(CodeBadRequest, "Invalid from CreatedAt")
		}
		lookup.From = event.From{
			ID:        id,
			CreatedAt: created,
		}
	}

	events, err := r.events.Fetch(ctx, lookup)
	if err != nil {
		log.Ctx(ctx).Err(err).Msg("Failed to fetch events.")
		return Events{}, NewInternalError()
	}
	res := Events{
		Items: make([]Event, len(events)),
		Limit: int32(lookup.Limit),
	}
	if len(events) != 0 {
		lastEvent := events[len(events)-1]
		res.NextFrom = &EventFrom{
			ID:        lastEvent.ID.String(),
			CreatedAt: lastEvent.CreatedAt.Format(time.RFC3339),
		}
	}
	for i, e := range events {
		payload, err := json.Marshal(e.Payload.Payload)
		if err != nil {
			log.Ctx(ctx).Err(err).Msg("Failed to marshal event.")
			return Events{}, NewInternalError()
		}
		res.Items[i] = Event{
			ID:        e.ID.String(),
			OwnerID:   e.Owner.String(),
			RoomID:    e.Room.String(),
			CreatedAt: e.CreatedAt.String(),
			Payload: EventPayload{
				Type:    string(e.Payload.Type),
				Payload: string(payload),
			},
		}
	}
	return res, nil
}

//go:embed schema.graphql
var schema string

const (
	batchLimit = 100
)
