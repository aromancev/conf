package web

import (
	"context"
	_ "embed"
	"encoding/base64"
	"encoding/json"
	"errors"

	"github.com/aromancev/confa/auth"
	"github.com/aromancev/confa/confa"
	"github.com/aromancev/confa/confa/talk"
	"github.com/aromancev/confa/confa/talk/clap"
	"github.com/aromancev/confa/profile"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

type Code string

const (
	CodeBadRequest       = "BAD_REQUEST"
	CodeUnauthorized     = "UNAUTHORIZED"
	CodeDuplicateEntry   = "DUPLICATE_ENTRY"
	CodeNotFound         = "NOT_FOUND"
	CodeAmbiguousLookup  = "AMBIGIOUS_LOOKUP"
	CodePermissionDenied = "PERMISSION_DENIED"
	CodeUnknown          = "UNKNOWN_CODE"
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
	return NewResolverError(CodeUnknown, "internal system error")
}

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
	State       string
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

type Profiles struct {
	Items    []Profile
	Limit    int32
	NextFrom string
}

type Profile struct {
	ID              string
	OwnerID         string
	Handle          string
	DisplayName     *string
	AvatarThumbnail *Image
}

type Image struct {
	Format string
	Data   string
}

type ProfileMask struct {
	Handle      *string
	DisplayName *string
}

type ProfileLookup struct {
	OwnerIDs *[]string
	Handle   *string
}

type UploadToken struct {
	URL      string
	FormData string
}

type Resolver struct {
	publicKey      *auth.PublicKey
	confas         *confa.CRUD
	talks          *talk.UserService
	claps          *clap.CRUD
	profiles       *profile.Mongo
	profileUpdater *profile.Updater
}

func NewResolver(pk *auth.PublicKey, confas *confa.CRUD, talks *talk.UserService, claps *clap.CRUD, profiles *profile.Mongo, uploader *profile.Updater) *Resolver {
	return &Resolver{
		publicKey:      pk,
		confas:         confas,
		talks:          talks,
		claps:          claps,
		profiles:       profiles,
		profileUpdater: uploader,
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
		return Confas{}, NewResolverError(CodeUnauthorized, "Invalid access token.")
	}

	lookup, err := newConfaLookup(args.Where, args.Limit, args.From)
	if err != nil {
		return Confas{Limit: args.Limit}, nil
	}

	confas, err := r.confas.Fetch(ctx, lookup)
	if err != nil {
		log.Ctx(ctx).Err(err).Msg("Failed to fetch confa.")
		return Confas{Limit: args.Limit}, NewInternalError()
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
		return Talks{}, NewResolverError(CodeUnauthorized, "Invalid access token.")
	}

	lookup, err := newTalkLookup(args.Where, args.Limit, args.From)
	if err != nil {
		return Talks{Limit: args.Limit}, nil
	}
	talks, err := r.talks.Fetch(ctx, lookup)
	if err != nil {
		log.Ctx(ctx).Err(err).Msg("Failed to fetch talks.")
		return Talks{Limit: args.Limit}, NewInternalError()
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
		return Claps{}, NewResolverError(CodeUnauthorized, "Invalid access token.")
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
		return Claps{}, NewInternalError()
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
		return Confa{}, NewResolverError(CodeUnauthorized, "Invalid access token.")
	}
	if !claims.AllowedWrite() {
		return Confa{}, NewResolverError(CodeUnauthorized, "Writes not allowed for guest.")
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
		return Confa{}, NewResolverError(CodeBadRequest, err.Error())
	case errors.Is(err, confa.ErrDuplicateEntry):
		return Confa{}, NewResolverError(CodeDuplicateEntry, "Confa already exists.")
	case err != nil:
		log.Ctx(ctx).Err(err).Msg("Failed to create confa.")
		return Confa{}, NewInternalError()
	}

	return newConfa(created), nil
}

func (r *Resolver) UpdateConfa(ctx context.Context, args struct {
	Where   ConfaLookup
	Request ConfaMask
}) (Confa, error) {
	var claims auth.APIClaims
	if err := r.publicKey.Verify(auth.Ctx(ctx).Token(), &claims); err != nil {
		return Confa{}, NewResolverError(CodeUnauthorized, "Invalid access token.")
	}
	if !claims.AllowedWrite() {
		return Confa{}, NewResolverError(CodeUnauthorized, "Writes not allowed for guest.")
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
		return Confa{}, NewResolverError(CodeBadRequest, err.Error())
	case errors.Is(err, confa.ErrNotFound):
		return Confa{}, NewResolverError(CodeNotFound, "")
	case errors.Is(err, confa.ErrDuplicateEntry):
		return Confa{}, NewResolverError(CodeDuplicateEntry, "Confa already exists.")
	case err != nil:
		log.Ctx(ctx).Err(err).Msg("Failed to update confa.")
		return Confa{}, NewInternalError()
	}
	return newConfa(updated), nil
}

func (r *Resolver) CreateTalk(ctx context.Context, args struct {
	Where   ConfaLookup
	Request TalkMask
}) (Talk, error) {
	var claims auth.APIClaims
	if err := r.publicKey.Verify(auth.Ctx(ctx).Token(), &claims); err != nil {
		return Talk{}, NewResolverError(CodeUnauthorized, "Invalid access token.")
	}
	if !claims.AllowedWrite() {
		return Talk{}, NewResolverError(CodeUnauthorized, "Writes not allowed for guest.")
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
		return Talk{}, NewResolverError(CodeNotFound, "Confa not found.")
	}

	created, err := r.talks.Create(ctx, claims.UserID, confaLookup, req)
	switch {
	case errors.Is(err, confa.ErrNotFound):
		return Talk{}, NewResolverError(CodeNotFound, "Confa not found.")
	case errors.Is(err, confa.ErrAmbiguousLookup):
		return Talk{}, NewResolverError(CodeAmbiguousLookup, "Confa lookup should match exactly one confa.")
	case errors.Is(err, talk.ErrValidation):
		return Talk{}, NewResolverError(CodeBadRequest, err.Error())
	case errors.Is(err, talk.ErrDuplicateEntry):
		return Talk{}, NewResolverError(CodeDuplicateEntry, "Talk already exists.")
	case err != nil:
		log.Ctx(ctx).Err(err).Msg("Failed to create talk.")
		return Talk{}, NewInternalError()
	}

	return newTalk(created), nil
}

func (r *Resolver) UpdateTalk(ctx context.Context, args struct {
	Where   TalkLookup
	Request TalkMask
}) (Talk, error) {
	var claims auth.APIClaims
	if err := r.publicKey.Verify(auth.Ctx(ctx).Token(), &claims); err != nil {
		return Talk{}, NewResolverError(CodeUnauthorized, "Invalid access token.")
	}
	if !claims.AllowedWrite() {
		return Talk{}, NewResolverError(CodeUnauthorized, "Writes not allowed for guest.")
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
		return Talk{}, NewResolverError(CodeBadRequest, err.Error())
	case errors.Is(err, confa.ErrNotFound):
		return Talk{}, NewResolverError(CodeNotFound, "Confa not found.")
	case errors.Is(err, talk.ErrDuplicateEntry):
		return Talk{}, NewResolverError(CodeDuplicateEntry, "Talk already exists.")
	case err != nil:
		log.Ctx(ctx).Err(err).Msg("Failed to update talk.")
		return Talk{}, NewInternalError()
	}
	return newTalk(updated), nil
}

func (r *Resolver) StartTalkRecording(ctx context.Context, args struct {
	Where TalkLookup
}) (Talk, error) {
	var claims auth.APIClaims
	if err := r.publicKey.Verify(auth.Ctx(ctx).Token(), &claims); err != nil {
		return Talk{}, NewResolverError(CodeUnauthorized, "Invalid access token.")
	}
	if !claims.AllowedWrite() {
		return Talk{}, NewResolverError(CodeUnauthorized, "Writes not allowed for guest.")
	}

	lookup, err := newTalkLookup(args.Where, 0, nil)
	if err != nil {
		return Talk{}, nil
	}
	started, err := r.talks.StartRecording(ctx, claims.UserID, lookup)
	switch {
	case errors.Is(err, talk.ErrValidation):
		return Talk{}, NewResolverError(CodeBadRequest, err.Error())
	case errors.Is(err, talk.ErrNotFound):
		return Talk{}, NewResolverError(CodeNotFound, "Talk not found")
	case errors.Is(err, talk.ErrAmbigiousLookup):
		return Talk{}, NewResolverError(CodeAmbiguousLookup, "Lookup should match exactly one talk.")
	case errors.Is(err, talk.ErrWrongState):
		return Talk{}, NewResolverError(CodeBadRequest, "Talk must be live to start recording.")
	case err != nil:
		log.Ctx(ctx).Err(err).Msg("Failed to update talk.")
		return Talk{}, NewInternalError()
	}
	return newTalk(started), nil
}

func (r *Resolver) StopTalkRecording(ctx context.Context, args struct {
	Where TalkLookup
}) (Talk, error) {
	var claims auth.APIClaims
	if err := r.publicKey.Verify(auth.Ctx(ctx).Token(), &claims); err != nil {
		return Talk{}, NewResolverError(CodeUnauthorized, "Invalid access token.")
	}
	if !claims.AllowedWrite() {
		return Talk{}, NewResolverError(CodeUnauthorized, "Writes not allowed for guest.")
	}

	lookup, err := newTalkLookup(args.Where, 0, nil)
	if err != nil {
		return Talk{}, nil
	}
	started, err := r.talks.StopRecording(ctx, claims.UserID, lookup)
	switch {
	case errors.Is(err, talk.ErrValidation):
		return Talk{}, NewResolverError(CodeBadRequest, err.Error())
	case errors.Is(err, talk.ErrNotFound):
		return Talk{}, NewResolverError(CodeNotFound, "Talk not found")
	case errors.Is(err, talk.ErrAmbigiousLookup):
		return Talk{}, NewResolverError(CodeAmbiguousLookup, "Lookup should match exactly one talk.")
	case errors.Is(err, talk.ErrWrongState):
		return Talk{}, NewResolverError(CodeBadRequest, "Talk must be recording to stop recording.")
	case err != nil:
		log.Ctx(ctx).Err(err).Msg("Failed to update talk.")
		return Talk{}, NewInternalError()
	}
	return newTalk(started), nil
}

func (r *Resolver) UpdateClap(ctx context.Context, args struct {
	TalkID string
	Value  int32
}) (string, error) {
	var claims auth.APIClaims
	if err := r.publicKey.Verify(auth.Ctx(ctx).Token(), &claims); err != nil {
		return "", NewResolverError(CodeUnauthorized, "Invalid access token.")
	}
	if !claims.AllowedWrite() {
		return "", NewResolverError(CodeUnauthorized, "Writes not allowed for guest.")
	}

	tID, err := uuid.Parse(args.TalkID)
	if err != nil {
		return "", NewResolverError(CodeNotFound, "Talk not found.")
	}
	id, err := r.claps.CreateOrUpdate(ctx, claims.UserID, tID, uint(args.Value))
	switch {
	case errors.Is(err, clap.ErrValidation):
		return "", NewResolverError(CodeBadRequest, err.Error())
	case err != nil:
		log.Ctx(ctx).Err(err).Msg("Failed to update clap.")
		return "", NewInternalError()
	}

	return id.String(), nil
}

func (r *Resolver) UpdateProfile(ctx context.Context, args struct {
	Request ProfileMask
}) (Profile, error) {
	var claims auth.APIClaims
	if err := r.publicKey.Verify(auth.Ctx(ctx).Token(), &claims); err != nil {
		return Profile{}, NewResolverError(CodeUnauthorized, "Invalid access token.")
	}
	if !claims.AllowedWrite() {
		return Profile{}, NewResolverError(CodeUnauthorized, "Writes not allowed for guest.")
	}

	request := profile.Profile{
		ID:    uuid.New(),
		Owner: claims.UserID,
	}
	if args.Request.Handle != nil {
		request.Handle = *args.Request.Handle
	}
	if args.Request.DisplayName != nil {
		request.DisplayName = *args.Request.DisplayName
	}

	upserted, err := r.profiles.CreateOrUpdate(ctx, request)
	switch {
	case errors.Is(err, profile.ErrValidation):
		return Profile{}, NewResolverError(CodeBadRequest, err.Error())
	case err != nil:
		log.Ctx(ctx).Err(err).Msg("Failed to update profile.")
		return Profile{}, NewInternalError()
	}

	return newProfile(upserted), nil
}

func (r *Resolver) Profiles(ctx context.Context, args struct {
	Where ProfileLookup
	Limit int32
	From  *string
}) (Profiles, error) {
	var claims auth.APIClaims
	if err := r.publicKey.Verify(auth.Ctx(ctx).Token(), &claims); err != nil {
		return Profiles{}, NewResolverError(CodeUnauthorized, "Invalid access token.")
	}

	var lookup profile.Lookup
	if args.Where.OwnerIDs != nil {
		lookup.Owners = make([]uuid.UUID, 0, len(*args.Where.OwnerIDs))
		for _, id := range *args.Where.OwnerIDs {
			parsed, err := uuid.Parse(id)
			if err != nil {
				continue
			}
			lookup.Owners = append(lookup.Owners, parsed)
		}
	}
	if args.Where.Handle != nil {
		lookup.Handle = *args.Where.Handle
	}

	fetched, err := r.profiles.Fetch(ctx, lookup)
	if err != nil {
		log.Ctx(ctx).Err(err).Msg("Failed to fetch profiles.")
		return Profiles{}, NewInternalError()
	}
	res := Profiles{
		Items: make([]Profile, len(fetched)),
		Limit: int32(lookup.Limit),
	}
	if len(fetched) > 0 {
		res.NextFrom = fetched[len(fetched)-1].ID.String()
	}
	for i, p := range fetched {
		res.Items[i] = newProfile(p)
	}
	return res, nil
}

func (r *Resolver) RequestAvatarUpload(ctx context.Context) (UploadToken, error) {
	var claims auth.APIClaims
	if err := r.publicKey.Verify(auth.Ctx(ctx).Token(), &claims); err != nil {
		return UploadToken{}, NewResolverError(CodeUnauthorized, "Invalid access token.")
	}

	url, data, err := r.profileUpdater.RequestUpload(ctx, claims.UserID)
	if err != nil {
		log.Ctx(ctx).Err(err).Msg("Failed to request upload.")
		return UploadToken{}, NewInternalError()
	}
	formData, err := json.Marshal(data)
	if err != nil {
		log.Ctx(ctx).Err(err).Msg("Failed to marshal form data.")
		return UploadToken{}, NewInternalError()
	}

	return UploadToken{
		URL:      url,
		FormData: string(formData),
	}, nil
}

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
		State:       string(t.State),
	}
}

func newProfile(p profile.Profile) Profile {
	api := Profile{
		ID:      p.ID.String(),
		OwnerID: p.Owner.String(),
		Handle:  p.Handle,
	}
	if p.DisplayName != "" {
		api.DisplayName = &p.DisplayName
	}
	if !p.AvatarThumbnail.IsEmpty() {
		api.AvatarThumbnail = &Image{
			Format: p.AvatarThumbnail.Format,
			Data:   base64.StdEncoding.EncodeToString(p.AvatarThumbnail.Data),
		}
	}
	return api
}

//go:embed schema.graphql
var schema string

const (
	batchLimit = 100
)
