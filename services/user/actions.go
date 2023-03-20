package user

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

type Repo interface {
	WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error
	GetOrCreate(ctx context.Context, request User) (User, error)
	UpdateOne(ctx context.Context, lookup Lookup, request Update) (User, error)
	FetchOne(ctx context.Context, lookup Lookup) (User, error)
}

type Actions struct {
	repo Repo
}

func NewActions(repo Repo) *Actions {
	return &Actions{
		repo: repo,
	}
}

func (a *Actions) GetOrCreate(ctx context.Context, request User) (User, error) {
	request.ID = uuid.New()
	return a.repo.GetOrCreate(ctx, request)
}

func (a *Actions) CreatePassword(ctx context.Context, ident Ident, password Password) (User, error) {
	if err := ident.Validate(); err != nil {
		return User{}, fmt.Errorf("%w: %s", ErrValidation, err)
	}
	if err := password.Validate(); err != nil {
		return User{}, fmt.Errorf("%w: %s", ErrValidation, err)
	}
	passwordHash, err := password.Hash()
	if err != nil {
		return User{}, fmt.Errorf("failed to hash password: %w", err)
	}

	user, err := a.repo.UpdateOne(
		ctx,
		Lookup{
			Idents: []Ident{
				ident,
			},
			WithoutPassword: true,
		},
		Update{
			PasswordHash: passwordHash,
		},
	)
	if err != nil {
		return User{}, fmt.Errorf("failed to update user: %w", err)
	}
	return user, nil
}

func (a *Actions) UpdatePassword(ctx context.Context, userID uuid.UUID, oldPassword, newPassword Password) (User, error) {
	if oldPassword == newPassword {
		return User{}, fmt.Errorf("%w: old and new password cannot be the same", ErrValidation)
	}
	if err := newPassword.Validate(); err != nil {
		return User{}, fmt.Errorf("%w: %s", ErrValidation, err)
	}
	newPasswordHash, err := newPassword.Hash()
	if err != nil {
		return User{}, fmt.Errorf("failed to hash password: %w", err)
	}
	var user User
	err = a.repo.WithTransaction(ctx, func(ctx context.Context) error {
		user, err := a.repo.FetchOne(ctx, Lookup{ID: userID})
		if err != nil {
			return fmt.Errorf("failed to fetch user:%w", err)
		}
		ok, err := oldPassword.Check(user.PasswordHash)
		if err != nil {
			return fmt.Errorf("failed to check password: %w", err)
		}
		if !ok {
			return fmt.Errorf("%w: invalid password", ErrValidation)
		}
		user, err = a.repo.UpdateOne(ctx, Lookup{ID: userID}, Update{PasswordHash: newPasswordHash})
		if err != nil {
			return fmt.Errorf("failed to update user: %w", err)
		}
		return nil
	})
	if err != nil {
		return User{}, fmt.Errorf("failed to commit transaction: %w", err)
	}
	return user, nil
}

func (a *Actions) ResetPassword(ctx context.Context, ident Ident, password Password) (User, error) {
	if err := ident.Validate(); err != nil {
		return User{}, fmt.Errorf("%w: %s", ErrValidation, err)
	}
	if err := password.Validate(); err != nil {
		return User{}, fmt.Errorf("%w: %s", ErrValidation, err)
	}
	passwordHash, err := password.Hash()
	if err != nil {
		return User{}, fmt.Errorf("failed to hash password: %w", err)
	}

	user, err := a.repo.UpdateOne(
		ctx,
		Lookup{
			Idents: []Ident{
				ident,
			},
		},
		Update{
			PasswordHash: passwordHash,
		},
	)
	if err != nil {
		return User{}, fmt.Errorf("failed to update user: %w", err)
	}
	return user, nil
}

func (a *Actions) CheckPassword(ctx context.Context, ident Ident, password Password) (User, error) {
	user, err := a.repo.FetchOne(
		ctx,
		Lookup{
			Idents: []Ident{
				ident,
			},
		},
	)
	if err != nil {
		// Spend aproximately same time to simulate a password check
		// so it's harder to find out if user exists with timing attacks.
		SimulatePasswordCheck()
		return User{}, fmt.Errorf("failed to fetch user:%w", err)
	}
	ok, err := password.Check(user.PasswordHash)
	if err != nil {
		return User{}, fmt.Errorf("failed to check password: %w", err)
	}
	if !ok {
		return User{}, fmt.Errorf("%w: invalid password", ErrNotFound)
	}
	return user, nil
}
