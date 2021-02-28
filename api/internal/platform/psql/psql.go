package psql

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
)

type Queryer interface {
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
}

type Execer interface {
	Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error)
}

type Conn interface {
	Queryer
	Execer
	Begin(ctx context.Context) (pgx.Tx, error)
}

type Config struct {
	Host, Port, User, Password, Database string
}

func New(c Config) (*sql.DB, error) {
	conn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s",
		c.Host, c.Port, c.User, c.Password, c.Database,
	)
	db, err := sql.Open("postgres", conn)
	if err != nil {
		return nil, err
	}

	return db, nil
}

// Transaction can be used to execute multiple operations atomically.
//
// It does not protect you from data races or any
// concurrency issues. It only means that either all operations will be committed or none. Use locks, constraints,
// retries, and other well-known patterns to avoid concurrency side-effects.
//
// Make sure to always use parameter provided to the callback to perform db operations.
// DO this:
// err = mysql.Transaction(ctx, db, func(ctx context.Context, tx *sql.Tx) error {
//	 _, err = tx.Execute(...)
//   . . .
//	 return nil
// })
// DO NOT do this:
// err = mysql.Transaction(ctx, db, func(ctx context.Context, tx *sql.Tx) error {
//	 _, err = db.Execute(...)
//   . . .
//	 return nil
// })
//
// If any error is returned from the callback, transaction will be aborted.
func Transaction(ctx context.Context, conn Conn, f func(context.Context, Conn) error) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	tx, err := conn.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}

	err = f(ctx, tx)
	if err != nil {
		_ = tx.Rollback(ctx) // We can't do anything if rollback returns an error.
		return err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil
}

// ValuesBuilder can be used to build queries for multiple insert with placeholders.
type ValuesBuilder struct {
	sql  strings.Builder
	args []interface{}
}

func NewValuesBuilder() *ValuesBuilder {
	return &ValuesBuilder{}
}

func (b *ValuesBuilder) WriteRow(args ...interface{}) {
	if len(b.args) != 0 {
		b.sql.WriteString(",")
	}
	b.sql.WriteString("(?" + strings.Repeat(",?", len(args)-1) + ")")
	b.args = append(b.args, args...)
}

func (b *ValuesBuilder) Query() (string, []interface{}) {
	return b.sql.String(), b.args
}
