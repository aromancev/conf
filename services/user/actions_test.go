package user

import (
	"context"
	"fmt"
	"math"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestActions(t *testing.T) {
	ctx := context.Background()

	t.Run("CreatePassword", func(t *testing.T) {
		t.Run("Happy path", func(t *testing.T) {
			t.Parallel()

			users := NewMongo(dockerMongo(t))
			actions := NewActions(users)

			_, err := users.GetOrCreate(ctx, User{
				ID: uuid.New(),
				Idents: []Ident{
					{
						Platform: PlatformGithub,
						Value:    "test",
					},
				},
			})
			require.NoError(t, err)

			ident := Ident{Platform: PlatformGithub, Value: "test"}
			_, err = actions.CreatePassword(ctx, ident, "testtest")
			require.NoError(t, err)

			_, err = actions.CheckPassword(ctx, ident, "testtest")
			require.NoError(t, err)
		})

		t.Run("Existing password returns not found", func(t *testing.T) {
			t.Parallel()

			users := NewMongo(dockerMongo(t))
			actions := NewActions(users)

			_, err := users.GetOrCreate(ctx, User{
				ID: uuid.New(),
				Idents: []Ident{
					{
						Platform: PlatformGithub,
						Value:    "test",
					},
				},
				PasswordHash: []byte{1},
			})
			require.NoError(t, err)

			_, err = actions.CreatePassword(ctx, Ident{Platform: PlatformGithub, Value: "test"}, "testtest")
			require.ErrorIs(t, err, ErrNotFound)
		})
	})

	t.Run("UpdatePassword", func(t *testing.T) {
		t.Run("Happy path", func(t *testing.T) {
			t.Parallel()

			users := NewMongo(dockerMongo(t))
			actions := NewActions(users)
			password := Password("testtest")
			hash, _ := password.Hash()

			user, err := users.GetOrCreate(ctx, User{
				ID: uuid.New(),
				Idents: []Ident{
					{
						Platform: PlatformGithub,
						Value:    "test",
					},
				},
				PasswordHash: hash,
			})
			require.NoError(t, err)

			_, err = actions.UpdatePassword(ctx, user.ID, "testtest", "testtest2")
			require.NoError(t, err)

			ident := Ident{Platform: PlatformGithub, Value: "test"}
			_, err = actions.CheckPassword(ctx, ident, "testtest2")
			require.NoError(t, err)
		})

		t.Run("Returns error if old and new are the same", func(t *testing.T) {
			t.Parallel()

			users := NewMongo(dockerMongo(t))
			actions := NewActions(users)

			_, err := actions.UpdatePassword(ctx, uuid.New(), "testtest", "testtest")
			require.ErrorIs(t, err, ErrValidation)
		})

		t.Run("Returns validation error if wrong password", func(t *testing.T) {
			t.Parallel()

			users := NewMongo(dockerMongo(t))
			actions := NewActions(users)
			password := Password("testtest")
			hash, _ := password.Hash()

			user, err := users.GetOrCreate(ctx, User{
				ID: uuid.New(),
				Idents: []Ident{
					{
						Platform: PlatformGithub,
						Value:    "test",
					},
				},
				PasswordHash: hash,
			})
			require.NoError(t, err)

			_, err = actions.UpdatePassword(ctx, user.ID, "wrongwrong", "testtest2")
			require.ErrorIs(t, err, ErrValidation)
		})

		t.Run("Happy concurrent", func(t *testing.T) {
			t.Parallel()

			users := NewMongo(dockerMongo(t))
			actions := NewActions(users)
			password := Password("testtest 0")
			hash, _ := password.Hash()

			user, err := users.GetOrCreate(ctx, User{
				ID: uuid.New(),
				Idents: []Ident{
					{
						Platform: PlatformGithub,
						Value:    "test",
					},
				},
				PasswordHash: hash,
			})
			require.NoError(t, err)

			const iterations = 3
			const threads = 2
			for i := 0; i < iterations; i++ {
				var updated atomic.Int64

				var wg sync.WaitGroup
				wg.Add(threads)
				for t := 0; t < threads; t++ {
					go func() {
						defer wg.Done()
						_, err = actions.UpdatePassword(ctx, user.ID, Password(fmt.Sprintf("testtest %d", i)), Password(fmt.Sprintf("testtest %d", i+1)))
						if err == nil {
							updated.Add(1)
						}
					}()
				}
				wg.Wait()
				require.EqualValues(t, 1, updated.Load(), "Only one thread should succeeed to update password.")
			}
		})
	})

	t.Run("Reset password", func(t *testing.T) {
		t.Run("Happy path", func(t *testing.T) {
			t.Parallel()

			users := NewMongo(dockerMongo(t))
			actions := NewActions(users)

			ident := Ident{
				Platform: PlatformGithub,
				Value:    "test",
			}
			_, err := users.GetOrCreate(ctx, User{
				ID:     uuid.New(),
				Idents: []Ident{ident},
			})
			require.NoError(t, err)

			_, err = actions.ResetPassword(
				ctx,
				Ident{
					Platform: ident.Platform,
					Value:    ident.Value,
				},
				"testtest",
			)
			require.NoError(t, err)
			_, err = actions.CheckPassword(
				ctx,
				Ident{
					Platform: ident.Platform,
					Value:    ident.Value,
				},
				"testtest",
			)
			require.NoError(t, err)

			_, err = actions.ResetPassword(
				ctx,
				Ident{
					Platform: ident.Platform,
					Value:    ident.Value,
				},
				"testtest2",
			)
			require.NoError(t, err)
			_, err = actions.CheckPassword(
				ctx, Ident{
					Platform: ident.Platform,
					Value:    ident.Value,
				},
				"testtest2",
			)
			require.NoError(t, err)
		})
		t.Run("Wrong identifier returns not found", func(t *testing.T) {
			t.Parallel()

			users := NewMongo(dockerMongo(t))
			actions := NewActions(users)

			_, err := users.GetOrCreate(ctx, User{
				ID: uuid.New(),
				Idents: []Ident{{
					Platform: PlatformGithub,
					Value:    "test",
				}},
			})
			require.NoError(t, err)

			_, err = actions.ResetPassword(
				ctx,
				Ident{
					Platform: PlatformGithub,
					Value:    "wrong",
				},
				"testtest",
			)
			require.ErrorIs(t, err, ErrNotFound)
		})
	})

	t.Run("CheckPassword", func(t *testing.T) {
		t.Run("Happy path", func(t *testing.T) {
			t.Parallel()

			users := NewMongo(dockerMongo(t))
			actions := NewActions(users)
			hash, _ := Password("testtest").Hash()

			ident := Ident{Platform: PlatformGithub, Value: "test"}
			created, err := users.GetOrCreate(ctx, User{
				ID: uuid.New(),
				Idents: []Ident{
					{
						Platform: ident.Platform,
						Value:    ident.Value,
					},
				},
				PasswordHash: hash,
			})
			require.NoError(t, err)

			checked, err := actions.CheckPassword(ctx, ident, "testtest")
			require.NoError(t, err)
			assert.Equal(t, created.ID, checked.ID)
		})

		t.Run("Wrong password or identifier returns not found", func(t *testing.T) {
			t.Parallel()

			users := NewMongo(dockerMongo(t))
			actions := NewActions(users)
			hash, _ := Password("testtest").Hash()

			ident := Ident{Platform: PlatformGithub, Value: "test"}
			_, err := users.GetOrCreate(ctx, User{
				ID: uuid.New(),
				Idents: []Ident{
					{
						Platform: ident.Platform,
						Value:    ident.Value,
					},
				},
				PasswordHash: hash,
			})
			require.NoError(t, err)

			_, err = actions.CheckPassword(ctx, ident, "wrongwrong")
			require.ErrorIs(t, err, ErrNotFound)
			_, err = actions.CheckPassword(ctx, Ident{Platform: ident.Platform, Value: "wrong"}, "testtest")
			require.ErrorIs(t, err, ErrNotFound)
		})

		t.Run("Wrong password or nonexistent user behave the same way", func(t *testing.T) {
			t.Parallel()

			users := NewMongo(dockerMongo(t))
			actions := NewActions(users)
			hash, _ := Password("testtest").Hash()

			ident := Ident{Platform: PlatformGithub, Value: "test"}
			_, err := users.GetOrCreate(ctx, User{
				ID: uuid.New(),
				Idents: []Ident{
					{
						Platform: ident.Platform,
						Value:    ident.Value,
					},
				},
				PasswordHash: hash,
			})
			require.NoError(t, err)

			start := time.Now()
			_, err = actions.CheckPassword(ctx, ident, "wrongwrong")
			require.ErrorIs(t, err, ErrNotFound)
			wrongPass := time.Since(start)

			start = time.Now()
			_, err = actions.CheckPassword(ctx, Ident{Platform: ident.Platform, Value: "wrong"}, "testtest")
			require.ErrorIs(t, err, ErrNotFound)
			notFound := time.Since(start)

			// Takes about the same time.
			diffMs := math.Abs(float64(wrongPass.Milliseconds() - notFound.Milliseconds()))
			assert.Less(t, diffMs, float64(wrongPass.Milliseconds()+notFound.Milliseconds())*0.1)
		})
	})
}
