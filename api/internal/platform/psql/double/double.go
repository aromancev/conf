package double

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	_ "github.com/lib/pq"
	"github.com/ory/dockertest/v3"
	"github.com/pressly/goose"
)

// NewDocker creates a Postgres container, runs migrations in it and returns a db connection to use.
// IMPORTANT: Always call returned function at the end of tests (if no error returned) to release docker resources.
// To make this work on CI, set DOCKER_HOST and DOCKER_IP env variables.
func NewDocker(migrationsDir string) (pgx.Tx, func()) {
	migrationsDir, err := filepath.Abs(migrationsDir)
	if err != nil {
		panic(err)
	}
	pg := pool.Use(migrationsDir)
	tx, err := pg.Begin(context.Background())
	if err != nil {
		pool.Done(migrationsDir)
		panic(err)
	}
	return tx, func() {
		_ = tx.Rollback(context.Background())
		pool.Done(migrationsDir)
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
	using int
	purge func()
}

func newDockerPool() *dockerPool {
	return &dockerPool{containers: map[string]*container{}}
}

func (p *dockerPool) Use(migrationsDir string) *pgxpool.Pool {
	p.Lock()
	defer p.Unlock()

	if c, ok := p.containers[migrationsDir]; ok {
		c.using++
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

	conn := fmt.Sprintf(
		"host=localhost port=%s user=postgres password=postgres dbname=postgres sslmode=disable",
		resource.GetPort("5432/tcp"),
	)

	var gooseConn *sql.DB
	err = pool.Retry(func() error {
		var err error
		gooseConn, err = sql.Open("postgres", conn)
		if err != nil {
			return err
		}
		return gooseConn.Ping()
	})
	if err != nil {
		_ = pool.Purge(resource)
		panic(err)
	}
	defer gooseConn.Close()

	goose.SetVerbose(false)
	err = goose.SetDialect("postgres")
	if err != nil {
		_ = pool.Purge(resource)
		panic(err)
	}
	goose.SetTableName("goose_db_version")
	err = goose.Up(gooseConn, migrationsDir)
	if err != nil {
		_ = pool.Purge(resource)
		panic(err)
	}

	pg, err := pgxpool.Connect(context.Background(), conn)
	if err != nil {
		_ = pool.Purge(resource)
		panic(err)
	}

	c := &container{
		pool:  pg,
		using: 1,
		purge: func() {
			pg.Close()
			_ = pool.Purge(resource)
		},
	}
	p.containers[migrationsDir] = c
	return pg
}

func (p *dockerPool) Done(migrationsDir string) {
	p.Lock()
	defer p.Unlock()

	c, ok := p.containers[migrationsDir]
	if !ok {
		return
	}
	if c.using > 1 {
		c.using--
	} else {
		p.containers[migrationsDir].purge()
		delete(p.containers, migrationsDir)
	}
}
