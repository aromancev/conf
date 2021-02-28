package confa

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/aromancev/confa/internal/platform/psql/double"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestSQL(t *testing.T) {
	t.Parallel()

	db, done := double.NewDocker(t, "")
	defer done()

	_, err := db.QueryContext(context.Background(), "select 1")
	require.NoError(t, err)
}

func TestPostAndGetSQL(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	var db *sql.DB
	//db, done := double.NewDocker(t, "")
	conn := fmt.Sprintf(
		"host=0.0.0.0 port=5432 user=confa password=confa dbname=confa sslmode=disable",
		)
	var err error
	db, err = sql.Open("postgres", conn)
	if err != nil {
		require.NoError(t, err)
	}
	db.Ping()
	println("JOAPS")
	confaSQL := NewSQL()

	id := uuid.New()
	c := Confa{}
	c.ID=uuid.New()
	c.Owner=uuid.New()
	c.Name="JOPA"

	created, err := confaSQL.Create(ctx, db, c)
	fetched, err := confaSQL.FetchOne(ctx, db, Lookup{ID: id})
	println(created[0].toString())
	println(fetched.toString())
}