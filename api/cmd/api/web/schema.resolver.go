package web

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"errors"
	"fmt"

	"github.com/aromancev/confa/internal/auth"
	"github.com/aromancev/confa/internal/confa"
	"github.com/aromancev/confa/internal/confa/talk"
	"github.com/aromancev/confa/internal/emails"
	"github.com/aromancev/confa/internal/platform/email"
	"github.com/aromancev/confa/internal/platform/trace"
	"github.com/aromancev/confa/internal/user/ident"
	"github.com/aromancev/confa/proto/queue"
	"github.com/google/uuid"
	"github.com/prep/beanstalk"
	"github.com/rs/zerolog/log"
)

func (r *mutationResolver) Login(ctx context.Context, address string) (string, error) {
	if err := email.Validate(address); err != nil {
		return "", newError(CodeInvalidParam, "Invalid email.")
	}

	token, err := r.signer.EmailToken(address)
	if err != nil {
		log.Ctx(ctx).Err(err).Msg("Failed to create email token.")
		return "", newError(CodeInternal, "")
	}

	msg, err := emails.Login(r.baseURL, address, token)
	if err != nil {
		log.Ctx(ctx).Err(err).Msg("Failed to render login email.")
		return "", newError(CodeInternal, "")
	}

	body, err := queue.Marshal(&queue.EmailJob{
		Emails: []*queue.Email{{
			FromName:  msg.FromName,
			ToAddress: msg.ToAddress,
			Subject:   msg.Subject,
			Html:      msg.HTML,
		}},
	}, trace.ID(ctx))
	if err != nil {
		log.Ctx(ctx).Err(err).Msg("Failed to marshal email.")
		return "", newError(CodeInternal, "")
	}

	id, err := r.producer.Put(ctx, queue.TubeEmail, body, beanstalk.PutParams{})
	if err != nil {
		log.Ctx(ctx).Err(err).Msg("Failed to put email job.")
		return "", newError(CodeInternal, "")
	}
	log.Ctx(ctx).Info().Uint64("jobId", id).Msg("Email login job emitted.")
	return address, nil
}

func (r *mutationResolver) CreateSession(ctx context.Context, emailToken string) (*Token, error) {
	claims, err := r.verifier.EmailToken(emailToken)
	if err != nil {
		return nil, newError(CodeUnauthorized, "Wrong email token.")
	}

	userID, err := r.idents.GetOrCreate(ctx, ident.Ident{
		Platform: ident.PlatformEmail,
		Value:    claims.Address,
	})
	if err != nil {
		return nil, newError(CodeInternal, "")
	}

	sess, err := r.sessions.Create(ctx, userID)
	if err != nil {
		log.Ctx(ctx).Err(err).Msg("Failed to create session.")
		return nil, newError(CodeInternal, "")
	}

	access, expiresIn, err := r.signer.AccessToken(userID)
	if err != nil {
		log.Ctx(ctx).Err(err).Msg("Failed to sign access token.")
		return nil, newError(CodeInternal, "")
	}

	auth.Ctx(ctx).SetSession(sess.Key)
	return &Token{
		Token:     access,
		ExpiresIn: int(expiresIn.Seconds()),
	}, nil
}

func (r *mutationResolver) CreateConfa(ctx context.Context, handle *string) (*Confa, error) {
	var claims auth.APIClaims
	if err := r.verifier.Verify(auth.Ctx(ctx).Token(), &claims); err != nil {
		return nil, newError(CodeUnauthorized, "Invalid access token.")
	}

	var req confa.Confa
	if handle != nil {
		req.Handle = *handle
	}
	created, err := r.confas.Create(ctx, claims.UserID, req)
	switch {
	case errors.Is(err, confa.ErrValidation):
		return nil, newError(CodeInvalidParam, err.Error())
	case err != nil:
		log.Ctx(ctx).Err(err).Msg("Failed to create confa.")
		return nil, newError(CodeInternal, "")
	}

	return &Confa{
		ID:     created.ID.String(),
		Owner:  created.Owner.String(),
		Handle: created.Handle,
	}, nil
}

func (r *mutationResolver) CreateTalk(ctx context.Context, handle *string, confa *string) (*Talk, error) {
	var claims auth.APIClaims
	if err := r.verifier.Verify(auth.Ctx(ctx).Token(), &claims); err != nil {
		return nil, newError(CodeUnauthorized, "Invalid access token.")
	}

	var req talk.Talk
	if handle != nil {
		req.Handle = *handle
	}
	if confa != nil {
		var err error
		req.Confa, err = uuid.Parse(*confa)
		if err != nil {
			return nil, fmt.Errorf("failed to create talk: %w", err)
		}
	}
	created, err := r.talks.Create(ctx, claims.UserID, req)
	switch {
	case errors.Is(err, talk.ErrValidation):
		return nil, newError(CodeInvalidParam, err.Error())
	case err != nil:
		log.Ctx(ctx).Err(err).Msg("failed to create talk.")
		return nil, newError(CodeInternal, "")
	}

	return &Talk{
		ID:      created.ID.String(),
		Confa:   created.Confa.String(),
		Owner:   created.Owner.String(),
		Speaker: created.Speaker.String(),
		Handle:  created.Handle,
	}, nil
}

func (r *queryResolver) Token(ctx context.Context) (*Token, error) {
	sessKey := auth.Ctx(ctx).Session()
	if sessKey == "" {
		return nil, newError(CodeUnauthorized, "Session is not present.")
	}
	sess, err := r.sessions.Fetch(ctx, sessKey)
	if err != nil {
		return nil, newError(CodeUnauthorized, "Invalid session.")
	}

	access, expiresIn, err := r.signer.AccessToken(sess.Owner)
	if err != nil {
		log.Ctx(ctx).Err(err).Msg("Failed to sign API token.")
		return nil, newError(CodeInternal, "")
	}
	return &Token{
		Token:     access,
		ExpiresIn: int(expiresIn.Seconds()),
	}, nil
}

func (r *queryResolver) Confas(ctx context.Context, where ConfaInput, from string, limit int) (*Confas, error) {
	if limit < 0 || limit > 20 {
		limit = 20
	}

	lookup := confa.Lookup{
		Limit: uint64(limit),
	}
	var err error
	if from != "" {
		lookup.From, err = uuid.Parse(from)
		if err != nil {
			return &Confas{Limit: limit}, nil
		}
	}
	if where.ID != nil {
		lookup.ID, err = uuid.Parse(*where.ID)
		if err != nil {
			return &Confas{Limit: limit}, nil
		}
	}
	if where.Owner != nil {
		lookup.Owner, err = uuid.Parse(*where.Owner)
		if err != nil {
			return &Confas{Limit: limit}, nil
		}
	}
	if where.Handle != nil {
		lookup.Handle = *where.Handle
	}

	confas, err := r.confas.Fetch(ctx, lookup)
	if err != nil {
		log.Ctx(ctx).Err(err).Msg("Failed to fetch confa.")
		return nil, newError(CodeInternal, "")
	}

	if len(confas) == 0 {
		return &Confas{Limit: limit}, nil
	}

	res := &Confas{
		Items:    make([]*Confa, len(confas)),
		Limit:    limit,
		NextFrom: confas[len(confas)-1].ID.String(),
	}
	for i, c := range confas {
		res.Items[i] = &Confa{
			ID:     c.ID.String(),
			Owner:  c.Owner.String(),
			Handle: c.Handle,
		}
	}

	return res, nil
}

func (r *queryResolver) Talks(ctx context.Context, where TalkInput, from string, limit int) (*Talks, error) {
	if limit < 0 || limit > 20 {
		limit = 20
	}
	lookup := talk.Lookup{
		Limit: uint64(limit),
	}
	var err error
	if from != "" {
		lookup.From, err = uuid.Parse(from)
		if err != nil {
			return &Talks{Limit: limit}, nil
		}
	}
	if where.ID != nil {
		lookup.ID, err = uuid.Parse(*where.ID)
		if err != nil {
			return &Talks{Limit: limit}, nil
		}
	}
	if where.Confa != nil {
		lookup.Confa, err = uuid.Parse(*where.Confa)
		if err != nil {
			return &Talks{Limit: limit}, nil
		}
	}
	if where.Owner != nil {
		lookup.Owner, err = uuid.Parse(*where.Owner)
		if err != nil {
			return &Talks{Limit: limit}, nil
		}
	}
	if where.Speaker != nil {
		lookup.Speaker, err = uuid.Parse(*where.Speaker)
		if err != nil {
			return &Talks{Limit: limit}, nil
		}
	}
	if where.Handle != nil {
		lookup.Handle = *where.Handle
	}
	talks, err := r.talks.Fetch(ctx, lookup)
	if err != nil {
		log.Ctx(ctx).Err(err).Msg("Failed to fetch confa.")
		return nil, newError(CodeInternal, "")
	}

	if len(talks) == 0 {
		return &Talks{Limit: limit}, nil
	}

	res := &Talks{
		Items:    make([]*Talk, len(talks)),
		Limit:    limit,
		NextFrom: talks[len(talks)-1].ID.String(),
	}
	for i, t := range talks {
		res.Items[i] = &Talk{
			ID:      t.ID.String(),
			Confa:   t.Confa.String(),
			Owner:   t.Owner.String(),
			Speaker: t.Speaker.String(),
			Handle:  t.Handle,
		}
	}

	return res, nil
}

// Mutation returns MutationResolver implementation.
func (r *Resolver) Mutation() MutationResolver { return &mutationResolver{r} }

// Query returns QueryResolver implementation.
func (r *Resolver) Query() QueryResolver { return &queryResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
