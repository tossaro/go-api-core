package postgres

import (
	"context"
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"

	"github.com/jackc/pgx/v4/pgxpool"
)

func seeder(p *pgxpool.Pool, f string) {
	files, err := ioutil.ReadDir(f)
	if err != nil {
		return
	}

	rows, err := p.Query(context.Background(), "CREATE TABLE IF NOT EXISTS seeds (id bigserial PRIMARY KEY,name varchar NOT NULL,created_at timestamptz NOT NULL DEFAULT (now()))")
	if err != nil {
		log.Print("Seeder: table initialize error: " + err.Error())
		return
	}
	rows.Close()

	var count int
	for _, file := range files {
		ctx := context.TODO()
		var name = strings.TrimSuffix(file.Name(), filepath.Ext(file.Name()))
		var id *int
		p.QueryRow(ctx, "SELECT id FROM seeds where name = $1", name).Scan(&id)
		if id != nil {
			continue
		}
		log.Print("Seeder: -- start seeding " + name + " --")

		tx, err := p.Begin(ctx)
		if err != nil {
			log.Print("Seeder: begin " + name + " error: " + err.Error())
			return
		}

		readFile, err := ioutil.ReadFile(f + "/" + file.Name())
		if err != nil {
			log.Print("Seeder: read file " + file.Name() + " error: " + err.Error())
			return
		}
		log.Print("Seeder: reading " + file.Name() + " succeed.")

		sqlStmtSlice := strings.Split(string(readFile), ";\r")

		defer func() {
			if err != nil {
				tx.Rollback(ctx)
				log.Print("Seeder: rolling back " + name)
			}
		}()

		for _, q := range sqlStmtSlice {
			_, err := tx.Exec(ctx, q)

			if err != nil {
				log.Print("Seeder: executing " + name + " error: " + err.Error())
				return
			}
		}

		err = tx.Commit(ctx)
		if err != nil {
			log.Print("Seeder: commiting " + name + " error: " + err.Error())
			return
		}

		rows, err := p.Query(ctx, "INSERT INTO seeds(name) values($1)", name)
		if err != nil {
			log.Print("Seeder: marking seed " + name + " error: " + err.Error())
			return
		}
		rows.Close()

		log.Print("Seeder: seeding " + name + " finished.")
		count++
	}
	if count == 0 {
		log.Print("Seeder: no change")
	}
}
