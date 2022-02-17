package clap

import (
	"context"

	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMongo(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("CreateOrUpdate", func(t *testing.T) {
		t.Parallel()

		claps := NewMongo(dockerMongo(t))
		request := Clap{
			ID:      uuid.New(),
			Owner:   uuid.New(),
			Speaker: uuid.New(),
			Confa:   uuid.New(),
			Talk:    uuid.New(),
			Value:   5,
		}
		id, err := claps.CreateOrUpdate(ctx, request)
		require.NoError(t, err)
		assert.Equal(t, request.ID, id)

		// More than 50 is not allowed.
		request.Value = 51
		_, err = claps.CreateOrUpdate(ctx, request)
		assert.ErrorIs(t, err, ErrValidation)
	})

	t.Run("Aggregate", func(t *testing.T) {
		tt := []struct {
			name    string
			inserts []Clap
			lookup  Lookup
			expect  uint64
			err     error
		}{
			{
				name: "Happy path",
				inserts: []Clap{
					{
						Owner:   uuid.UUID{1},
						Speaker: uuid.UUID{1},
						Confa:   uuid.UUID{1},
						Talk:    uuid.UUID{1},
						Value:   1,
					},
				},
				lookup: Lookup{
					Confa: uuid.UUID{1},
				},
				expect: 1,
			},
			{
				name:    "Empty lookup returns zero",
				inserts: nil,
				lookup: Lookup{
					Confa: uuid.UUID{1},
				},
				expect: 0,
			},
			{
				name: "Updates value",
				inserts: []Clap{
					{
						Owner:   uuid.UUID{1},
						Speaker: uuid.UUID{1},
						Confa:   uuid.UUID{1},
						Talk:    uuid.UUID{1},
						Value:   1,
					},
					{
						Owner:   uuid.UUID{1},
						Speaker: uuid.UUID{1},
						Confa:   uuid.UUID{1},
						Talk:    uuid.UUID{1},
						Value:   2,
					},
				},
				lookup: Lookup{
					Confa: uuid.UUID{1},
				},
				expect: 2,
			},
			{
				name: "Sums multiple claps",
				inserts: []Clap{
					{
						Owner:   uuid.UUID{1},
						Speaker: uuid.UUID{1},
						Confa:   uuid.UUID{1},
						Talk:    uuid.UUID{1},
						Value:   1,
					},
					{
						Owner:   uuid.UUID{2},
						Speaker: uuid.UUID{1},
						Confa:   uuid.UUID{1},
						Talk:    uuid.UUID{1},
						Value:   1,
					},
				},
				lookup: Lookup{
					Confa: uuid.UUID{1},
				},
				expect: 2,
			},
			{
				name: "Finds by speaker",
				inserts: []Clap{
					{
						Owner:   uuid.UUID{1},
						Speaker: uuid.UUID{1},
						Confa:   uuid.UUID{1},
						Talk:    uuid.UUID{1},
						Value:   1,
					},
					{
						Owner:   uuid.UUID{2},
						Speaker: uuid.UUID{1},
						Confa:   uuid.UUID{2},
						Talk:    uuid.UUID{2},
						Value:   1,
					},
				},
				lookup: Lookup{
					Speaker: uuid.UUID{1},
				},
				expect: 2,
			},
			{
				name: "Finds by talk",
				inserts: []Clap{
					{
						Owner:   uuid.UUID{1},
						Speaker: uuid.UUID{1},
						Confa:   uuid.UUID{1},
						Talk:    uuid.UUID{1},
						Value:   1,
					},
					{
						Owner:   uuid.UUID{1},
						Speaker: uuid.UUID{2},
						Confa:   uuid.UUID{1},
						Talk:    uuid.UUID{1},
						Value:   1,
					},
				},
				lookup: Lookup{
					Talk: uuid.UUID{1},
				},
				expect: 2,
			},
			{
				name: "Finds by owner",
				inserts: []Clap{
					{
						Owner:   uuid.UUID{1},
						Speaker: uuid.UUID{1},
						Confa:   uuid.UUID{1},
						Talk:    uuid.UUID{1},
						Value:   1,
					},
					{
						Owner:   uuid.UUID{2},
						Speaker: uuid.UUID{1},
						Confa:   uuid.UUID{1},
						Talk:    uuid.UUID{1},
						Value:   1,
					},
				},
				lookup: Lookup{
					Owner: uuid.UUID{1},
				},
				expect: 1,
			},
		}

		for _, c := range tt {
			c := c // Parallel execution protection.
			t.Run(c.name, func(t *testing.T) {
				t.Parallel()

				claps := NewMongo(dockerMongo(t))
				for _, ins := range c.inserts {
					ins.ID = uuid.New()
					id, err := claps.CreateOrUpdate(ctx, ins)
					require.NoError(t, err)
					require.Equal(t, ins.ID, id)
				}
				sum, err := claps.Aggregate(ctx, c.lookup)
				require.ErrorIs(t, err, c.err)
				assert.Equal(t, int(c.expect), int(sum))
			})
		}
	})
}
