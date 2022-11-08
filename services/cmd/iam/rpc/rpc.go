package rpc

import (
	"context"
	"errors"
	"fmt"

	"github.com/aromancev/confa/internal/proto/iam"
	"github.com/aromancev/confa/user"
	"github.com/google/uuid"
	"github.com/twitchtv/twirp"
)

type Handler struct {
	users *user.Mongo
}

func NewHandler(users *user.Mongo) *Handler {
	return &Handler{
		users: users,
	}
}

func (h *Handler) GetUser(ctx context.Context, lookup *iam.UserLookup) (*iam.User, error) {
	var userID uuid.UUID
	_ = userID.UnmarshalBinary(lookup.UserId)
	usr, err := h.users.FetchOne(ctx, user.Lookup{
		ID: userID,
	})
	switch {
	case errors.Is(err, user.ErrNotFound):
		return nil, twirp.NewError(twirp.NotFound, "User not found.")
	case err != nil:
		return nil, fmt.Errorf("failed to fetch user: %w", err)
	}
	idents := make([]*iam.User_Ident, len(usr.Idents))
	for i, ident := range usr.Idents {
		idents[i] = &iam.User_Ident{
			Value:    ident.Value,
			Platform: iam.Platform_UNKNOWN,
		}
		switch ident.Platform {
		case user.PlatformEmail:
			idents[i].Platform = iam.Platform_EMAIL
		case user.PlatformTwitter:
			idents[i].Platform = iam.Platform_TWITTER
		case user.PlatformGithub:
			idents[i].Platform = iam.Platform_GITHUB
		}
	}
	return &iam.User{
		Id:     lookup.UserId,
		Idents: idents,
	}, nil
}
