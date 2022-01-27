package web

import (
	"context"
	_ "embed"
	"errors"

	"github.com/aromancev/confa/auth"
	"github.com/aromancev/confa/internal/confa"
	"github.com/aromancev/confa/internal/confa/talk"
	"github.com/aromancev/confa/internal/confa/talk/clap"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

type Service struct {
	Name    string
	Version string
	Schema  string
}

type Confas struct {
	Items    []Confa
	Limit    int32
	NextFrom string
}

type Confa struct {
	ID          string
	OwnerID     string
	Handle      string
	Title       string
	Description string
}

type ConfaLookup struct {
	ID      *string
	OwnerID *string
	Handle  *string
}

type ConfaMask struct {
	Handle      *string
	Title       *string
	Description *string
}

type Talk struct {
	ID          string
	OwnerID     string
	SpeakerID   string
	ConfaID     string
	RoomID      string
	Handle      string
	Title       string
	Description string
}

type TalkLookup struct {
	ID        *string
	OwnerID   *string
	SpeakerID *string
	ConfaID   *string
	Handle    *string
}

type Talks struct {
	Items    []Talk
	Limit    int32
	NextFrom string
}

type TalkMask struct {
	Handle      *string
	Title       *string
	Description *string
}

type Claps struct {
	Value     int32
	UserValue int32
}

type ClapLookup struct {
	SpeakerID *string
	ConfaID   *string
	TalkID    *string
}

type Resolver struct {
	publicKey *auth.PublicKey
	confas    *confa.CRUD
	talks     *talk.CRUD
	claps     *clap.CRUD
}

func NewResolver(pk *auth.PublicKey, confas *confa.CRUD, talks *talk.CRUD, claps *clap.CRUD) *Resolver {
	return &Resolver{
		publicKey: pk,
		confas:    confas,
		talks:     talks,
		claps:     claps,
	}
}

func (r *Resolver) Service(_ context.Context) Service {
	return Service{
		Name:    "confa",
		Version: "0.1.0",
		Schema:  schema,
	}
}

func (r *Resolver) Confas(ctx context.Context, args struct {
	Where ConfaLookup
	Limit int32
	From  *string
}) (Confas, error) {
	var claims auth.APIClaims
	if err := r.publicKey.Verify(auth.Ctx(ctx).Token(), &claims); err != nil {
		return Confas{}, newError(CodeUnauthorized, "Invalid access token.")
	}

	lookup, err := newConfaLookup(args.Where, args.Limit, args.From)
	if err != nil {
		return Confas{Limit: args.Limit}, nil
	}

	confas, err := r.confas.Fetch(ctx, lookup)
	if err != nil {
		log.Ctx(ctx).Err(err).Msg("Failed to fetch confa.")
		return Confas{Limit: args.Limit}, newInternalError()
	}

	if len(confas) == 0 {
		return Confas{Limit: args.Limit}, nil
	}

	res := Confas{
		Items: make([]Confa, len(confas)),
		Limit: int32(lookup.Limit),
	}
	if len(confas) > 0 {
		res.NextFrom = confas[len(confas)-1].ID.String()
	}
	for i, c := range confas {
		res.Items[i] = newConfa(c)
	}
	return res, nil
}

func (r *Resolver) Talks(ctx context.Context, args struct {
	Where TalkLookup
	Limit int32
	From  *string
}) (Talks, error) {
	var claims auth.APIClaims
	if err := r.publicKey.Verify(auth.Ctx(ctx).Token(), &claims); err != nil {
		return Talks{}, newError(CodeUnauthorized, "Invalid access token.")
	}

	lookup, err := newTalkLookup(args.Where, args.Limit, args.From)
	if err != nil {
		return Talks{Limit: args.Limit}, nil
	}
	talks, err := r.talks.Fetch(ctx, lookup)
	if err != nil {
		log.Ctx(ctx).Err(err).Msg("failed to fetch talk.")
		return Talks{Limit: args.Limit}, newInternalError()
	}

	if len(talks) == 0 {
		return Talks{Limit: args.Limit}, nil
	}

	res := Talks{
		Items: make([]Talk, len(talks)),
		Limit: int32(lookup.Limit),
	}
	if len(talks) > 0 {
		res.NextFrom = talks[len(talks)-1].ID.String()
	}
	for i, t := range talks {
		res.Items[i] = newTalk(t)
	}
	return res, nil
}

func (r *Resolver) AggregateClaps(ctx context.Context, args struct {
	Where ClapLookup
}) (Claps, error) {
	var claims auth.APIClaims
	if err := r.publicKey.Verify(auth.Ctx(ctx).Token(), &claims); err != nil {
		return Claps{}, newError(CodeUnauthorized, "Invalid access token.")
	}

	var lookup clap.Lookup
	var err error
	if args.Where.ConfaID != nil {
		lookup.Confa, err = uuid.Parse(*args.Where.ConfaID)
		if err != nil {
			return Claps{}, nil
		}
	}
	if args.Where.SpeakerID != nil {
		lookup.Speaker, err = uuid.Parse(*args.Where.SpeakerID)
		if err != nil {
			return Claps{}, nil
		}
	}
	if args.Where.TalkID != nil {
		lookup.Talk, err = uuid.Parse(*args.Where.TalkID)
		if err != nil {
			return Claps{}, nil
		}
	}
	res, err := r.claps.Aggregate(ctx, lookup, claims.UserID)
	if err != nil {
		log.Ctx(ctx).Err(err).Msg("Failed to aggregate claps.")
		return Claps{}, newInternalError()
	}
	claps := Claps{
		Value:     int32(res.Value),
		UserValue: int32(res.UserValue),
	}
	return claps, nil
}

func (r *Resolver) CreateConfa(ctx context.Context, args struct {
	Request ConfaMask
}) (Confa, error) {
	var claims auth.APIClaims
	if err := r.publicKey.Verify(auth.Ctx(ctx).Token(), &claims); err != nil {
		return Confa{}, newError(CodeUnauthorized, "Invalid access token.")
	}
	if !claims.AllowedWrite() {
		return Confa{}, newError(CodeUnauthorized, "Writes not allowed for guest.")
	}

	var req confa.Confa
	if args.Request.Handle != nil {
		req.Handle = *args.Request.Handle
	}
	if args.Request.Title != nil {
		req.Title = *args.Request.Title
	}
	if args.Request.Description != nil {
		req.Description = *args.Request.Description
	}
	created, err := r.confas.Create(ctx, claims.UserID, req)
	switch {
	case errors.Is(err, confa.ErrValidation):
		return Confa{}, newError(CodeBadRequest, err.Error())
	case errors.Is(err, confa.ErrDuplicateEntry):
		return Confa{}, newError(CodeDuplicateEntry, err.Error())
	case err != nil:
		log.Ctx(ctx).Err(err).Msg("Failed to create confa.")
		return Confa{}, newInternalError()
	}

	return newConfa(created), nil
}

func (r *Resolver) UpdateConfa(ctx context.Context, args struct {
	Where   ConfaLookup
	Request ConfaMask
}) (Confa, error) {
	var claims auth.APIClaims
	if err := r.publicKey.Verify(auth.Ctx(ctx).Token(), &claims); err != nil {
		return Confa{}, newError(CodeUnauthorized, "Invalid access token.")
	}
	if !claims.AllowedWrite() {
		return Confa{}, newError(CodeUnauthorized, "Writes not allowed for guest.")
	}

	lookup, err := newConfaLookup(args.Where, 0, nil)
	if err != nil {
		return Confa{}, nil
	}

	mask := confa.Mask{
		Handle:      args.Request.Handle,
		Title:       args.Request.Title,
		Description: args.Request.Description,
	}
	updated, err := r.confas.Update(ctx, claims.UserID, lookup, mask)
	switch {
	case errors.Is(err, confa.ErrValidation):
		return Confa{}, newError(CodeBadRequest, err.Error())
	case errors.Is(err, confa.ErrNotFound):
		return Confa{}, newError(CodeNotFound, err.Error())
	case errors.Is(err, confa.ErrDuplicateEntry):
		return Confa{}, newError(CodeDuplicateEntry, err.Error())
	case err != nil:
		log.Ctx(ctx).Err(err).Msg("Failed to update confa.")
		return Confa{}, newInternalError()
	}
	return newConfa(updated), nil
}

func (r *Resolver) CreateTalk(ctx context.Context, args struct {
	Where   ConfaLookup
	Request TalkMask
}) (Talk, error) {
	var claims auth.APIClaims
	if err := r.publicKey.Verify(auth.Ctx(ctx).Token(), &claims); err != nil {
		return Talk{}, newError(CodeUnauthorized, "Invalid access token.")
	}
	if !claims.AllowedWrite() {
		return Talk{}, newError(CodeUnauthorized, "Writes not allowed for guest.")
	}

	var req talk.Talk
	if args.Request.Handle != nil {
		req.Handle = *args.Request.Handle
	}
	if args.Request.Title != nil {
		req.Handle = *args.Request.Title
	}
	if args.Request.Description != nil {
		req.Handle = *args.Request.Description
	}
	confaLookup, err := newConfaLookup(args.Where, 1, nil)
	if err != nil {
		return Talk{}, newError(CodeNotFound, "Confa not found.")
	}

	created, err := r.talks.Create(ctx, claims.UserID, confaLookup, req)
	switch {
	case errors.Is(err, confa.ErrNotFound):
		return Talk{}, newError(CodeNotFound, err.Error())
	case errors.Is(err, confa.ErrUnexpectedResult):
		return Talk{}, newError(CodeBadRequest, err.Error())
	case errors.Is(err, talk.ErrValidation):
		return Talk{}, newError(CodeBadRequest, err.Error())
	case errors.Is(err, talk.ErrDuplicateEntry):
		return Talk{}, newError(CodeDuplicateEntry, err.Error())
	case err != nil:
		log.Ctx(ctx).Err(err).Msg("Failed to create talk.")
		return Talk{}, newInternalError()
	}

	return newTalk(created), nil
}

func (r *Resolver) UpdateTalk(ctx context.Context, args struct {
	Where   TalkLookup
	Request TalkMask
}) (Talk, error) {
	var claims auth.APIClaims
	if err := r.publicKey.Verify(auth.Ctx(ctx).Token(), &claims); err != nil {
		return Talk{}, newError(CodeUnauthorized, "Invalid access token.")
	}
	if !claims.AllowedWrite() {
		return Talk{}, newError(CodeUnauthorized, "Writes not allowed for guest.")
	}

	lookup, err := newTalkLookup(args.Where, 0, nil)
	if err != nil {
		return Talk{}, nil
	}
	mask := talk.Mask{
		Handle:      args.Request.Handle,
		Title:       args.Request.Title,
		Description: args.Request.Description,
	}
	updated, err := r.talks.Update(ctx, claims.UserID, lookup, mask)
	switch {
	case errors.Is(err, confa.ErrValidation):
		return Talk{}, newError(CodeBadRequest, err.Error())
	case errors.Is(err, confa.ErrNotFound):
		return Talk{}, newError(CodeNotFound, err.Error())
	case errors.Is(err, talk.ErrDuplicateEntry):
		return Talk{}, newError(CodeDuplicateEntry, err.Error())
	case err != nil:
		log.Ctx(ctx).Err(err).Msg("Failed to update talk.")
		return Talk{}, newInternalError()
	}
	return newTalk(updated), nil
}

func (r *Resolver) UpdateClap(ctx context.Context, args struct {
	TalkID string
	Value  int32
}) (string, error) {
	var claims auth.APIClaims
	if err := r.publicKey.Verify(auth.Ctx(ctx).Token(), &claims); err != nil {
		return "", newError(CodeUnauthorized, "Invalid access token.")
	}
	if !claims.AllowedWrite() {
		return "", newError(CodeUnauthorized, "Writes not allowed for guest.")
	}

	tID, err := uuid.Parse(args.TalkID)
	if err != nil {
		return "", newError(CodeNotFound, "Talk not found.")
	}
	id, err := r.claps.CreateOrUpdate(ctx, claims.UserID, tID, uint(args.Value))
	switch {
	case errors.Is(err, clap.ErrValidation):
		return "", newError(CodeBadRequest, err.Error())
	case err != nil:
		log.Ctx(ctx).Err(err).Msg("Failed to update clap.")
		return "", newInternalError()
	}

	return id.String(), nil
}

//go:embed schema.graphql
var schema string

func newConfaLookup(input ConfaLookup, limit int32, from *string) (confa.Lookup, error) {
	if limit <= 0 || limit > batchLimit {
		limit = batchLimit
	}

	lookup := confa.Lookup{
		Limit: int64(limit),
	}
	var err error
	if from != nil {
		lookup.From, err = uuid.Parse(*from)
		if err != nil {
			return confa.Lookup{}, err
		}
	}
	if input.ID != nil {
		lookup.ID, err = uuid.Parse(*input.ID)
		if err != nil {
			return confa.Lookup{}, err
		}
	}
	if input.OwnerID != nil {
		lookup.Owner, err = uuid.Parse(*input.OwnerID)
		if err != nil {
			return confa.Lookup{}, err
		}
	}
	if input.Handle != nil {
		lookup.Handle = *input.Handle
	}
	return lookup, nil
}

func newConfa(c confa.Confa) Confa {
	return Confa{
		ID:          c.ID.String(),
		OwnerID:     c.Owner.String(),
		Handle:      c.Handle,
		Title:       c.Title,
		Description: c.Description,
	}
}

func newTalkLookup(input TalkLookup, limit int32, from *string) (talk.Lookup, error) {
	if limit < 0 || limit > batchLimit {
		limit = batchLimit
	}
	lookup := talk.Lookup{
		Limit: int64(limit),
	}
	var err error
	if from != nil {
		lookup.From, err = uuid.Parse(*from)
		if err != nil {
			return talk.Lookup{}, err
		}
	}
	if input.ID != nil {
		lookup.ID, err = uuid.Parse(*input.ID)
		if err != nil {
			return talk.Lookup{}, err
		}
	}
	if input.ConfaID != nil {
		lookup.Confa, err = uuid.Parse(*input.ConfaID)
		if err != nil {
			return talk.Lookup{}, err
		}
	}
	if input.OwnerID != nil {
		lookup.Owner, err = uuid.Parse(*input.OwnerID)
		if err != nil {
			return talk.Lookup{}, err
		}
	}
	if input.SpeakerID != nil {
		lookup.Speaker, err = uuid.Parse(*input.SpeakerID)
		if err != nil {
			return talk.Lookup{}, err
		}
	}
	if input.Handle != nil {
		lookup.Handle = *input.Handle
	}
	return lookup, nil
}

func newTalk(t talk.Talk) Talk {
	return Talk{
		ID:          t.ID.String(),
		ConfaID:     t.Confa.String(),
		OwnerID:     t.Owner.String(),
		SpeakerID:   t.Speaker.String(),
		RoomID:      t.Room.String(),
		Handle:      t.Handle,
		Title:       t.Title,
		Description: t.Description,
	}
}
