package storage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/krixlion/dev_forum-lib/event"
	"github.com/krixlion/dev_forum-lib/logging"
	"github.com/krixlion/dev_forum-user/pkg/entity"
	"go.opentelemetry.io/otel/trace"
)

var _ Storage = (*DB)(nil)

// DB is a wrapper for the read model and write model to use with Storage interface.
type DB struct {
	cmd    Eventstore
	query  Storage
	logger logging.Logger
	tracer trace.Tracer
}

func NewCQRStorage(eventstore Eventstore, query Storage, logger logging.Logger, tracer trace.Tracer) CQRStorage {
	return &DB{
		cmd:    eventstore,
		query:  query,
		logger: logger,
		tracer: tracer,
	}
}

func (storage DB) Close() error {
	var errMsg string

	if err := storage.cmd.Close(); err != nil {
		errMsg = fmt.Sprintf("%s, failed to close eventStore: %s", errMsg, err)
	}

	if err := storage.query.Close(); err != nil {
		errMsg = fmt.Sprintf("failed to close readStorage: %s", err)
	}

	if errMsg != "" {
		return errors.New(errMsg)
	}

	return nil
}

func (storage DB) Get(ctx context.Context, id string) (entity.User, error) {
	return storage.query.Get(ctx, id)
}

func (storage DB) GetMultiple(ctx context.Context, offset, limit string) ([]entity.User, error) {
	return storage.query.GetMultiple(ctx, offset, limit)
}

func (storage DB) Update(ctx context.Context, user entity.User) error {
	return storage.cmd.Update(ctx, user)
}

func (storage DB) Create(ctx context.Context, user entity.User) error {
	return storage.cmd.Create(ctx, user)
}

func (storage DB) Delete(ctx context.Context, id string) error {
	return storage.cmd.Delete(ctx, id)
}

// CatchUp handles events required to keep the read model consistent.
func (db DB) CatchUp(e event.Event) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	ctx, span := db.tracer.Start(ctx, "redis.CatchUp")
	defer span.End()

	switch e.Type {
	case event.UserCreated:
		var user entity.User
		if err := json.Unmarshal(e.Body, &user); err != nil {
			// // tracing.SetSpanErr(span, err)
			db.logger.Log(ctx, "Failed to parse event",
				"err", err,
				"event", e,
			)
			return
		}

		if err := db.query.Create(ctx, user); err != nil {
			// // tracing.SetSpanErr(span, err)
			db.logger.Log(ctx, "Failed to create user",
				"err", err,
				"event", e,
			)
		}
		return

	case event.UserDeleted:
		var id string
		if err := json.Unmarshal(e.Body, &id); err != nil {
			// // tracing.SetSpanErr(span, err)
			db.logger.Log(ctx, "Failed to parse event",
				"err", err,
				"event", e,
			)
		}

		if err := db.query.Delete(ctx, id); err != nil {
			// // tracing.SetSpanErr(span, err)
			db.logger.Log(ctx, "Failed to delete user",
				"err", err,
				"event", e,
			)
		}
		return

	case event.UserUpdated:
		var user entity.User
		if err := json.Unmarshal(e.Body, &user); err != nil {
			// // tracing.SetSpanErr(span, err)
			db.logger.Log(ctx, "Failed to parse event",
				"err", err,
				"event", e,
			)
			return
		}

		if err := db.query.Update(ctx, user); err != nil {
			// // tracing.SetSpanErr(span, err)

			db.logger.Log(ctx, "Failed to update user",
				"err", err,
				"event", e,
			)
		}
		return
	}
}
