package dbmocks

import (
	"context"

	"github.com/krixlion/dev_forum-user/pkg/entity"
	"github.com/stretchr/testify/mock"
)

type Storage struct {
	*mock.Mock
}

func NewStorage() Storage {
	return Storage{
		Mock: new(mock.Mock),
	}
}

func (m Storage) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m Storage) Get(ctx context.Context, filter string) (entity.User, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).(entity.User), args.Error(1)
}

func (m Storage) GetMultiple(ctx context.Context, offset, limit, filter string) ([]entity.User, error) {
	args := m.Called(ctx, offset, limit)
	return args.Get(0).([]entity.User), args.Error(1)
}

func (m Storage) Create(ctx context.Context, a entity.User) error {
	args := m.Called(ctx, a)
	return args.Error(0)
}

func (m Storage) Update(ctx context.Context, a entity.User) error {
	args := m.Called(ctx, a)
	return args.Error(0)
}

func (m Storage) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
