package postgres

import (
	"context"
	"log"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
)

const (
	_defaultPoolMax          = 1
	_defaultConnAttempts     = 10
	_defaultConnTimeout      = time.Second
	_defaultMigrationsFolder = "migrations"
	_defaultSeedsFolder      = "seeds"
)

type (
	Options struct {
		Url              string
		PoolMax          *int
		ConnAttempts     *int
		ConnTimeout      *time.Duration
		MigrationsFolder *string
		SeedsFolder      *string
	}

	Postgres struct {
		Pool *pgxpool.Pool
	}
)

func New(o *Options) *Postgres {
	if o.Url == "" {
		log.Fatal("postgres - URL option not provided")
	}

	poolConfig, err := pgxpool.ParseConfig(o.Url)
	if err != nil {
		log.Fatal("postgres - parse config error: %w", err)
	}

	pMax := _defaultPoolMax
	if o.PoolMax != nil {
		pMax = *(o.PoolMax)
	}
	poolConfig.MaxConns = int32(pMax)

	cAt := _defaultConnAttempts
	if o.ConnAttempts != nil {
		cAt = *(o.ConnAttempts)
	}

	cTm := _defaultConnTimeout
	if o.ConnTimeout != nil {
		cTm = *(o.ConnTimeout)
	}

	mF := _defaultMigrationsFolder
	if o.MigrationsFolder != nil {
		mF = *(o.MigrationsFolder)
	}

	sF := _defaultSeedsFolder
	if o.SeedsFolder != nil {
		sF = *(o.SeedsFolder)
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
		log.Fatal("postgres - connection error: %w", err)
	}

	migrate(o.Url+"?sslmode=disable", mF)
	seeder(p, sF)

	return &Postgres{p}
}

func (p *Postgres) Close() {
	if p.Pool != nil {
		p.Pool.Close()
	}
}
