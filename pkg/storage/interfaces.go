package storage

import (
	"context"
	"errors"

	"github.com/krixlion/dev_forum-lib/filter"
	"github.com/krixlion/dev_forum-user/pkg/entity"
)

var ErrNotFound error = errors.New("not found")

type Storage interface {
	Getter
	Writer
}

type Getter interface {
	Get(ctx context.Context, filter filter.Filter) (entity.User, error)
	GetMultiple(ctx context.Context, offset, limit string, filter filter.Filter) ([]entity.User, error)
}

type Writer interface {
	Create(context.Context, entity.User) error
	Update(context.Context, entity.User) error
	Delete(ctx context.Context, id string) error
}
