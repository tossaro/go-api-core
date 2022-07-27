package postgres

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
)

const (
	_defaultPoolMax      = 1
	_defaultConnAttempts = 10
	_defaultConnTimeout  = time.Second
)

type (
	Options struct {
		Url          string
		PoolMax      *int
		ConnAttempts *int
		ConnTimeout  *time.Duration
	}

	Postgres struct {
		Pool *pgxpool.Pool
	}
)

func New(opts *Options) (*Postgres, error) {
	poolConfig, err := pgxpool.ParseConfig(opts.Url)
	if err != nil {
		return nil, fmt.Errorf("postgres - pgxpool.ParseConfig: %w", err)
	}

	pMax := _defaultPoolMax
	if opts.PoolMax != nil {
		pMax = *(opts.PoolMax)
	}
	poolConfig.MaxConns = int32(pMax)

	cAt := _defaultConnAttempts
	if opts.ConnAttempts != nil {
		cAt = *(opts.ConnAttempts)
	}

	cTm := _defaultConnTimeout
	if opts.ConnTimeout != nil {
		cTm = *(opts.ConnTimeout)
	}

	var p *pgxpool.Pool
	for cAt > 0 {
		p, err = pgxpool.ConnectConfig(context.Background(), poolConfig)
		if err == nil {
			break
		}
		log.Printf("Postgres is trying to connect, attempts left: %d", cAt)
		time.Sleep(cTm)
		cAt--
	}

	if err != nil {
		return nil, fmt.Errorf("postgres - connection failed: %w", err)
	}

	migrate(opts.Url + "?sslmode=disable")

	return &Postgres{p}, nil
}

func (p *Postgres) Close() {
	if p.Pool != nil {
		p.Pool.Close()
	}
}
