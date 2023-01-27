package main

import (
	"log"
	"os"

	"github.com/krixlion/dev_forum-user/migrations"
	"github.com/krixlion/dev_forum-user/pkg/env"
	"github.com/krixlion/dev_forum-user/pkg/storage/db"
	"github.com/pressly/goose/v3"
)

func main() {
	env.Load("app")

	goose.SetBaseFS(&migrations.EmbedPath)
	db_port := os.Getenv("DB_PORT")
	db_host := os.Getenv("DB_HOST")
	db_user := os.Getenv("DB_USER")
	db_pass := os.Getenv("DB_PASS")
	db_name := os.Getenv("DB_NAME")
	storage, err := db.Make(db_host, db_port, db_user, db_pass, db_name)
	if err != nil {
		log.Fatalf("Failed to migrate: %v", err)
	}

	if err := goose.SetDialect(db.Driver); err != nil {
		log.Fatalf("Failed to migrate: %v", err)
	}

	if err := goose.Up(storage.Conn(), "."); err != nil {
		log.Fatalf("Failed to migrate: %v", err)
	}
}
