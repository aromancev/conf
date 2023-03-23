package web

import (
	"context"
	_ "embed"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/aromancev/confa/confa"
	"github.com/aromancev/confa/confa/talk"
	"github.com/aromancev/confa/internal/auth"
	"github.com/aromancev/confa/internal/routes"
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

type Confa struct {
	ID          string
	OwnerID     string
	Handle      string
	Title       string
	Description string
	CreatedAt   string
}

type Confas struct {
	Items []Confa
	Limit int32
	Next  *ConfaCursor
}

type ConfaCursor struct {
	ID        *string
	CreatedAt *string
	Asc       *bool
}

type ConfaLookup struct {
	ID      *string
	OwnerID *string
	Handle  *string
}

type ConfaUpdate struct {
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
	CreatedAt   string
}

type TalkLookup struct {
	ID        *string
	OwnerID   *string
	SpeakerID *string
	ConfaID   *string
	Handle    *string
}

type Talks struct {
	Items []Talk
	Limit int32
	Next  *TalkCursor
}

type TalkCursor struct {
	ID        *string
	CreatedAt *string
	Asc       *bool
}

type TalkUpdate struct {
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
	Items []Profile
	Limit int32
	Next  *ProfileCursor
}

type Profile struct {
	ID              string
	OwnerID         string
	Handle          string
	GivenName       *string
	FamilyName      *string
	AvatarThumbnail *Image
	AvatarURL       *string
}

type ProfileCursor struct {
	ID *string
}

type Image struct {
	Format string
	Data   string
}

type ProfileUpdate struct {
	Handle     *string
	GivenName  *string
	FamilyName *string
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
	confas         *confa.User
	talks          *talk.User
	profiles       *profile.Mongo
	profileUpdater *profile.Updater
	storageRoutes  *routes.Storage
}

func NewResolver(pk *auth.PublicKey, confas *confa.User, talks *talk.User, profiles *profile.Mongo, uploader *profile.Updater, storageRoutes *routes.Storage) *Resolver {
	return &Resolver{
		publicKey:      pk,
		confas:         confas,
		talks:          talks,
		profiles:       profiles,
		profileUpdater: uploader,
		storageRoutes:  storageRoutes,
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
	Where  ConfaLookup
	Limit  int32
	Cursor *ConfaCursor
}) (Confas, error) {
	var claims auth.APIClaims
	if err := r.publicKey.Verify(auth.Ctx(ctx).Token(), &claims); err != nil {
		return Confas{}, NewResolverError(CodeUnauthorized, "Invalid access token.")
	}

	lookup, err := newConfaLookup(args.Where, args.Limit, args.Cursor)
	if err != nil {
		return Confas{Limit: args.Limit}, NewResolverError(CodeBadRequest, "Fiter params are not valid.")
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
	for i, c := range confas {
		res.Items[i] = newConfa(c)
	}
	if len(res.Items) == int(lookup.Limit) {
		last := res.Items[len(res.Items)-1]
		res.Next = &ConfaCursor{
			ID:        &last.ID,
			CreatedAt: &last.CreatedAt,
		}
	}
	return res, nil
}

func (r *Resolver) Talks(ctx context.Context, args struct {
	Where  TalkLookup
	Limit  int32
	Cursor *TalkCursor
}) (Talks, error) {
	var claims auth.APIClaims
	if err := r.publicKey.Verify(auth.Ctx(ctx).Token(), &claims); err != nil {
		return Talks{}, NewResolverError(CodeUnauthorized, "Invalid access token.")
	}

	lookup, err := newTalkLookup(args.Where, args.Limit, args.Cursor)
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
	for i, t := range talks {
		res.Items[i] = newTalk(t)
	}
	if len(res.Items) == int(lookup.Limit) {
		last := res.Items[len(res.Items)-1]
		res.Next = &TalkCursor{
			ID:        &last.ID,
			CreatedAt: &last.CreatedAt,
		}
	}
	return res, nil
}

func (r *Resolver) CreateConfa(ctx context.Context, args struct {
	Request ConfaUpdate
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
	Request ConfaUpdate
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
		return Confa{}, NewResolverError(CodeBadRequest, "Fiter params are not valid.")
	}

	updated, err := r.confas.Update(ctx, claims.UserID, lookup, confa.Update{
		Handle:      args.Request.Handle,
		Title:       args.Request.Title,
		Description: args.Request.Description,
	})
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
	Request TalkUpdate
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
		return Talk{}, NewResolverError(CodeBadRequest, "Fiter params are not valid.")
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
	Request TalkUpdate
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
	updated, err := r.talks.Update(ctx, claims.UserID, lookup, talk.Update{
		Handle:      args.Request.Handle,
		Title:       args.Request.Title,
		Description: args.Request.Description,
	})
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

func (r *Resolver) UpdateProfile(ctx context.Context, args struct {
	Request ProfileUpdate
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
	if args.Request.GivenName != nil {
		request.GivenName = *args.Request.GivenName
	}
	if args.Request.FamilyName != nil {
		request.FamilyName = *args.Request.FamilyName
	}

	upserted, err := r.profiles.CreateOrUpdate(ctx, request)
	switch {
	case errors.Is(err, profile.ErrValidation):
		return Profile{}, NewResolverError(CodeBadRequest, err.Error())
	case err != nil:
		log.Ctx(ctx).Err(err).Msg("Failed to update profile.")
		return Profile{}, NewInternalError()
	}

	return newProfile(r.storageRoutes, upserted), nil
}

func (r *Resolver) Profiles(ctx context.Context, args struct {
	Where  ProfileLookup
	Limit  int32
	Cursor *ProfileCursor
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
	if args.Cursor != nil && args.Cursor.ID != nil {
		id, err := uuid.Parse(*args.Cursor.ID)
		if err == nil {
			lookup.From.ID = id
		}
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
	for i, p := range fetched {
		res.Items[i] = newProfile(r.storageRoutes, p)
	}
	if len(res.Items) > 0 {
		last := res.Items[len(res.Items)-1]
		res.Next = &ProfileCursor{
			ID: &last.ID,
		}
	}
	return res, nil
}

func (r *Resolver) RequestAvatarUpload(ctx context.Context) (UploadToken, error) {
	var claims auth.APIClaims
	if err := r.publicKey.Verify(auth.Ctx(ctx).Token(), &claims); err != nil {
		return UploadToken{}, NewResolverError(CodeUnauthorized, "Invalid access token.")
	}

	url, data, err := r.profileUpdater.UpdateAndRequestUpload(ctx, claims.UserID)
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

func newConfaLookup(input ConfaLookup, limit int32, cursor *ConfaCursor) (confa.Lookup, error) {
	if limit <= 0 || limit > batchLimit {
		limit = batchLimit
	}

	lookup := confa.Lookup{
		Limit: int64(limit),
	}
	var err error
	if cursor != nil {
		if cursor.CreatedAt != nil {
			createdAt, err := strconv.ParseInt(*cursor.CreatedAt, 10, 64)
			if err != nil {
				return confa.Lookup{}, NewResolverError(CodeBadRequest, "Invalid cursor.createdAt")
			}
			lookup.From.CreatedAt = time.UnixMilli(createdAt)
		}
		if cursor.ID != nil {
			id, err := uuid.Parse(*cursor.ID)
			if err == nil {
				lookup.From.ID = id
			}
		}
		if cursor.Asc != nil && *cursor.Asc {
			lookup.Asc = true
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
		CreatedAt:   fmt.Sprint(c.CreatedAt.UnixMilli()),
	}
}

func newTalkLookup(input TalkLookup, limit int32, cursor *TalkCursor) (talk.Lookup, error) {
	if limit < 0 || limit > batchLimit {
		limit = batchLimit
	}
	lookup := talk.Lookup{
		Limit: int64(limit),
	}
	var err error
	if cursor != nil {
		if cursor.CreatedAt != nil {
			createdAt, err := strconv.ParseInt(*cursor.CreatedAt, 10, 64)
			if err != nil {
				return talk.Lookup{}, NewResolverError(CodeBadRequest, "Invalid cursor.createdAt")
			}
			lookup.From.CreatedAt = time.UnixMilli(createdAt)
		}
		if cursor.ID != nil {
			id, err := uuid.Parse(*cursor.ID)
			if err == nil {
				lookup.From.ID = id
			}
		}
		if cursor.Asc != nil && *cursor.Asc {
			lookup.Asc = true
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
		CreatedAt:   fmt.Sprint(t.CreatedAt.UnixMilli()),
	}
}

func newProfile(storage *routes.Storage, p profile.Profile) Profile {
	api := Profile{
		ID:      p.ID.String(),
		OwnerID: p.Owner.String(),
		Handle:  p.Handle,
	}
	if p.GivenName != "" {
		api.GivenName = &p.GivenName
	}
	if p.FamilyName != "" {
		api.FamilyName = &p.FamilyName
	}
	if !p.AvatarThumbnail.IsEmpty() {
		api.AvatarThumbnail = &Image{
			Format: p.AvatarThumbnail.Format,
			Data:   base64.StdEncoding.EncodeToString(p.AvatarThumbnail.Data),
		}
	}
	if p.AvatarID != uuid.Nil {
		url := storage.ProfileAvatar(p.Owner, p.AvatarID)
		api.AvatarURL = &url
	}
	return api
}

//go:embed schema.graphql
var schema string

const (
	batchLimit = 100
)
