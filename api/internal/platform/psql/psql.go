package psql

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

type Queryer interface {
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
}

type Execer interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
}

type QueryerExecer interface {
	Queryer
	Execer
}

type Transactioner interface {
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
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
func Transaction(ctx context.Context, trans Transactioner, f func(context.Context, *sql.Tx) error) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	tx, err := trans.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}

	err = f(ctx, tx)
	if err != nil {
		_ = tx.Rollback() // We can't do anything if rollback returns an error.
		return err
	}

	err = tx.Commit()
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
	//b.sql.WriteString("(?" + strings.Repeat(",?", len(args)-1) + ")")
	b.sql.WriteString("(?" + strings.Repeat(",?", len(args)-1) + ")")
	b.args = append(b.args, args...)
}

func (b *ValuesBuilder) GetRow(args ...interface{}) {
	if len(b.args) != 0 {
		b.sql.WriteString(",")
	}
	b.sql.WriteString("(" + strings.Repeat(",", len(args)-1) + ")")
	b.args = append(b.args, args...)
}

func (b *ValuesBuilder) Query() (string, []interface{}) {
	return b.sql.String(), b.args
}

//// ValuesGetter can be used to build queries for multiple get by id (fields).
//type ValuesGetter struct {
//	sql  strings.Builder
//	args []interface{}
//}

//func NewValuesGetter() *ValuesGetter {
//	return &ValuesGetter{}
//}
//
//func (b *ValuesGetter) GetRow(args ...interface{}) {
//	if len(b.args) != 0 {
//		b.sql.WriteString(",")
//	}
//	b.sql.WriteString("(?" + strings.Repeat(",?", len(args)-1) + ")")
//	b.args = append(b.args, args...)
//}
//
//func (b *ValuesBuilder) Query() (string, []interface{}) {
//	return b.sql.String(), b.args
//}
