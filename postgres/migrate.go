package postgres

import (
	"errors"
	"log"
	"time"

	gMigrate "github.com/golang-migrate/migrate/v4"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

const (
	_defaultAttempts = 10
	_defaultTimeout  = time.Second
)

func migrate(url string, f string) {
	var (
		attempts = _defaultAttempts
		err      error
		m        *gMigrate.Migrate
	)

	for attempts > 0 {
		m, err = gMigrate.New("file://"+f, url)
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
	if err != nil && !errors.Is(err, gMigrate.ErrNoChange) {
		log.Printf("Migrate: up error: %s", err)
		return
	}

	if errors.Is(err, gMigrate.ErrNoChange) {
		log.Printf("Migrate: no change")
		return
	}

	log.Printf("Migrate: up success")
}
