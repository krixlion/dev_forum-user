package main

import (
	"context"
	"log"
	"os"

	"github.com/krixlion/dev_forum-lib/env"
	"github.com/krixlion/dev_forum-user/migrations"
	"github.com/krixlion/dev_forum-user/pkg/storage/cockroach"
	"github.com/pressly/goose/v3"
	"go.opentelemetry.io/otel"
)

func main() {
	env.Load("app")
	tracer := otel.Tracer("user-service")
	_, span := tracer.Start(context.Background(), "Migrate")
	defer span.End()

	goose.SetBaseFS(&migrations.EmbedPath)
	db_port := os.Getenv("DB_PORT")
	db_host := os.Getenv("DB_HOST")
	db_user := os.Getenv("DB_USER")
	db_pass := os.Getenv("DB_PASS")
	db_name := os.Getenv("DB_NAME")
	storage, err := cockroach.Make(db_host, db_port, db_user, db_pass, db_name, tracer)
	if err != nil {
		log.Fatalf("Failed to make DB: %v", err)
	}

	if err := goose.SetDialect(cockroach.Driver); err != nil {
		log.Fatalf("Failed to set dialect: %v", err)
	}

	if err := goose.Up(storage.Conn(), "."); err != nil {
		log.Fatalf("Failed to migrate: %v", err)
	}
}
