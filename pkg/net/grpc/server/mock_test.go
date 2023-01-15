package server_test

import (
	"context"

	"github.com/stretchr/testify/mock"
)

type mockStorage struct {
	mock.Mock
}

func (m mockStorage) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m mockStorage) Get(ctx context.Context, id string) (entity.Entity, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(entity.Entity), args.Error(1)
}

func (m mockStorage) GetMultiple(ctx context.Context, offset string, limit string) ([]entity.Entity, error) {
	args := m.Called(ctx, offset, limit)
	return args.Get(0).([]entity.Entity), args.Error(1)
}

func (m mockStorage) Create(ctx context.Context, a entity.Entity) error {
	args := m.Called(ctx, a)
	return args.Error(0)
}

func (m mockStorage) Update(ctx context.Context, a entity.Entity) error {
	args := m.Called(ctx, a)
	return args.Error(0)
}

func (m mockStorage) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m mockStorage) CatchUp(e event.Event) {
	m.Called(e)
}
