package storage

import (
	"context"
	"io"

	"github.com/krixlion/dev_forum-user/pkg/entity"
	"github.com/krixlion/dev_forum-user/pkg/event"
)

type CQRStorage interface {
	Getter
	Writer
	CatchUp(event.Event)
}

type Storage interface {
	Getter
	Writer
}

type Getter interface {
	io.Closer
	Get(ctx context.Context, id string) (entity.User, error)
	GetMultiple(ctx context.Context, offset, limit string) ([]entity.User, error)
}

type Writer interface {
	io.Closer
	Create(context.Context, entity.User) error
	Update(context.Context, entity.User) error
	Delete(ctx context.Context, id string) error
}

type Eventstore interface {
	event.Consumer
	Writer
}
