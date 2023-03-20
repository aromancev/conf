package web

import (
	"context"
	_ "embed"
	"errors"
	"strings"

	"github.com/aromancev/confa/internal/auth"
	"github.com/aromancev/confa/user"
)

type Code string

const (
	CodeBadRequest     = "BAD_REQUEST"
	CodeUnauthorized   = "UNAUTHORIZED"
	CodeDuplicateEntry = "DUPLICATE_ENTRY"
	CodeNotFound       = "NOT_FOUND"
	CodeUnknown        = "UNKNOWN_CODE"
)

type User struct {
	ID          string
	Identifiers []Identifier
	HasPassword bool
}

type Identifier struct {
	Platform string
	Value    string
}

type Service struct {
	Name    string
	Version string
	Schema  string
}

type PasswordUpdate struct {
	UserID string
}

type Resolver struct {
	publicKey *auth.PublicKey
	user      *user.Mongo
}

func NewResolver(pk *auth.PublicKey, u *user.Mongo) *Resolver {
	return &Resolver{
		publicKey: pk,
		user:      u,
	}
}

func (r *Resolver) Service(_ context.Context) Service {
	return Service{
		Name:    "iam",
		Version: "0.1.0",
		Schema:  schema,
	}
}

func (r *Resolver) User(ctx context.Context) (User, error) {
	var claims auth.APIClaims
	if err := r.publicKey.Verify(auth.Ctx(ctx).Token(), &claims); err != nil {
		return User{}, NewResolverError(CodeUnauthorized, "Invalid access token.")
	}

	usr, err := r.user.FetchOne(ctx, user.Lookup{ID: claims.UserID})
	switch {
	case errors.Is(err, user.ErrNotFound):
		return User{}, NewResolverError(CodeNotFound, "User does not exist.")
	case err != nil:
		return User{}, NewInternalError()
	}

	idents := make([]Identifier, len(usr.Idents))
	for i, ident := range usr.Idents {
		idents[i] = Identifier{
			Platform: strings.ToUpper(string(ident.Platform)),
			Value:    ident.Value,
		}
	}
	return User{
		ID:          usr.ID.String(),
		Identifiers: idents,
		HasPassword: len(usr.PasswordHash) != 0,
	}, nil
}

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

//go:embed schema.graphql
var schema string
