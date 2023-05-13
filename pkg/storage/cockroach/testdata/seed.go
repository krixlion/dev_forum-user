package testdata

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"

	_ "github.com/lib/pq"

	"github.com/krixlion/dev_forum-lib/env"
)

func init() {
	if err := initTestData(); err != nil {
		panic(err)
	}

}

func Seed() error {
	env.Load("app")

	port := os.Getenv("DB_PORT")
	host := os.Getenv("DB_HOST")
	user := os.Getenv("DB_USER")
	pass := os.Getenv("DB_PASS")
	dbName := os.Getenv("DB_NAME")

	db, err := sql.Open("postgres", fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=disable", user, pass, host, port, dbName))
	if err != nil {
		return err
	}

	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	if _, err := db.ExecContext(ctx, `TRUNCATE "users";`); err != nil {
		return err
	}

	stmt, err := db.PrepareContext(ctx, `INSERT INTO users (id, name, email, password)	VALUES ($1, $2, $3, $4);`)
	if err != nil {
		return err
	}

	for _, user := range Users {
		if _, err := stmt.ExecContext(ctx, user.Id, user.Name, user.Email, user.Password); err != nil {
			return err
		}
	}

	return nil
}
