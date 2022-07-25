package postgres

import (
	"errors"
	"log"
	"time"

	"github.com/golang-migrate/migrate/v4"
	core "github.com/tossaro/go-api-core"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

const (
	_defaultAttempts = 10
	_defaultTimeout  = time.Second
)

func init() {
	cfg, e := core.NewConfig()
	if e != nil {
		log.Printf("Config error: %s", e)
	}

	cfg.Postgre.Url += "?sslmode=disable"

	var (
		attempts = _defaultAttempts
		err      error
		m        *migrate.Migrate
	)

	for attempts > 0 {
		m, err = migrate.New("file://migrations", cfg.Postgre.Url)
		if err == nil {
			break
		}

		log.Printf("Migrate: trying to connect / read file, attempts left: %d", attempts)
		time.Sleep(_defaultTimeout)
		attempts--
	}

	if err != nil {
		log.Printf("Migrate: error: %s", err)
		return
	}

	err = m.Up()
	defer m.Close()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		log.Printf("Migrate: up error: %s", err)
		return
	}

	if errors.Is(err, migrate.ErrNoChange) {
		log.Printf("Migrate: no change")
		return
	}

	log.Printf("Migrate: up success")
}
