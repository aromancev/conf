package double

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/ory/dockertest/v3"
)

// NewDocker creates a Postgres container, runs migrations in it and returns a db connection to use.
// create callback will be called if this called starts up a new container.
// IMPORTANT: Always call returned function at the end of tests (if no error returned) to release docker resources.
// To make this work on CI, set DOCKER_HOST and DOCKER_IP env variables.
func NewDocker(key string, create func(conn *pgx.Conn)) (pgx.Tx, func()) {
	pg := pool.Use(key, create)
	tx, err := pg.Begin(context.Background())
	if err != nil {
		panic(err)
	}
	return tx, func() {
		_ = tx.Rollback(context.Background())
	}
}

var (
	pool = newDockerPool()
)

type dockerPool struct {
	sync.Mutex
	containers map[string]*container
}

type container struct {
	pool  *pgxpool.Pool
	purge func()
}

func newDockerPool() *dockerPool {
	return &dockerPool{containers: map[string]*container{}}
}

func (p *dockerPool) Use(key string, create func(conn *pgx.Conn)) *pgxpool.Pool {
	p.Lock()
	defer p.Unlock()

	ctx := context.Background()

	if c, ok := p.containers[key]; ok {
		return c.pool
	}

	pool, err := dockertest.NewPool(os.Getenv("DOCKER_HOST"))
	if err != nil {
		panic(err)
	}

	resource, err := pool.Run(
		"postgres",
		"12.6-alpine",
		[]string{
			"POSTGRES_PASSWORD=postgres",
			"POSTGRES_USER=postgres",
		},
	)
	if err != nil {
		panic(err)
	}

	defer func() {
		if err := recover(); err != nil {
			_ = pool.Purge(resource)
			panic(err)
		}
	}()

	var pg *pgxpool.Pool
	err = pool.Retry(func() error {
		pg, err = pgxpool.Connect(
			ctx,
			fmt.Sprintf(
				"host=localhost port=%s user=postgres password=postgres dbname=postgres sslmode=disable",
				resource.GetPort("5432/tcp"),
			),
		)
		return err
	})
	if err != nil {
		panic(err)
	}

	if create != nil {
		conn, err := pg.Acquire(ctx)
		if err != nil {
			panic(err)
		}
		create(conn.Conn())
		conn.Release()
	}

	c := &container{
		pool: pg,
		purge: func() {
			pg.Close()
			_ = pool.Purge(resource)
		},
	}
	p.containers[key] = c
	return pg
}

func Purge() {
	pool.Lock()
	defer pool.Unlock()

	for _, c := range pool.containers {
		c.purge()
	}
}
