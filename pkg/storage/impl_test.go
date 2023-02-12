package storage_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/krixlion/dev_forum-lib/mocks"
	"github.com/krixlion/dev_forum-lib/nulls"
	"github.com/krixlion/dev_forum-user/pkg/entity"
	"github.com/krixlion/dev_forum-user/pkg/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_Get(t *testing.T) {
	type args struct {
		ctx context.Context
		id  string
	}

	tests := []struct {
		desc    string
		query   mocks.Storage[entity.User]
		args    args
		want    entity.User
		wantErr bool
	}{
		{
			desc: "Test if method is invoked",
			args: args{
				ctx: context.Background(),
				id:  "",
			},
			want: entity.User{},
			query: func() mocks.Storage[entity.User] {
				m := mocks.Storage[entity.User]{Mock: new(mock.Mock)}
				m.On("Get", mock.Anything, mock.AnythingOfType("string")).Return(entity.User{}, nil).Once()
				return m
			}(),
		},
		{
			desc: "Test if method forwards an error",
			args: args{
				ctx: context.Background(),
				id:  "",
			},
			want:    entity.User{},
			wantErr: true,
			query: func() mocks.Storage[entity.User] {
				m := mocks.Storage[entity.User]{Mock: new(mock.Mock)}
				m.On("Get", mock.Anything, mock.AnythingOfType("string")).Return(entity.User{}, errors.New("test err")).Once()
				return m
			}(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			db := storage.NewCQRStorage(nil, tt.query, nulls.NullLogger{}, nulls.NullTracer{})
			got, err := db.Get(tt.args.ctx, tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("storage.Get():\n error = %+v\n wantErr = %+v\n", err, tt.wantErr)
				return
			}

			if !cmp.Equal(got, tt.want) {
				t.Errorf("storage.Get():\n got = %+v\n want = %+v\n", got, tt.want)
				return
			}
			assert.True(t, tt.query.AssertCalled(t, "Get", mock.Anything, tt.args.id))
		})
	}
}
func Test_GetMultiple(t *testing.T) {
	type args struct {
		ctx    context.Context
		offset string
		limit  string
	}

	tests := []struct {
		desc    string
		query   mocks.Storage[entity.User]
		args    args
		want    []entity.User
		wantErr bool
	}{
		{
			desc: "Test if method is invoked",
			args: args{
				ctx:    context.Background(),
				limit:  "",
				offset: "",
			},
			want: []entity.User{},
			query: func() mocks.Storage[entity.User] {
				m := mocks.Storage[entity.User]{Mock: new(mock.Mock)}
				m.On("GetMultiple", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return([]entity.User{}, nil).Once()
				return m
			}(),
		},
		{
			desc: "Test if method forwards an error",
			args: args{
				ctx:    context.Background(),
				limit:  "",
				offset: "",
			},
			want:    []entity.User{},
			wantErr: true,
			query: func() mocks.Storage[entity.User] {
				m := mocks.Storage[entity.User]{Mock: new(mock.Mock)}
				m.On("GetMultiple", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return([]entity.User{}, errors.New("test err")).Once()
				return m
			}(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			db := storage.NewCQRStorage(mocks.Eventstore[entity.User]{}, tt.query, nulls.NullLogger{}, nulls.NullTracer{})
			got, err := db.GetMultiple(tt.args.ctx, tt.args.offset, tt.args.limit)
			if (err != nil) != tt.wantErr {
				t.Errorf("storage.GetMultiple():\n error = %+v\n wantErr = %+v\n", err, tt.wantErr)
				return
			}

			if !cmp.Equal(got, tt.want, cmpopts.EquateEmpty()) {
				t.Errorf("storage.GetMultiple():\n got = %+v\n want = %+v\n", got, tt.want)
				return
			}

			assert.True(t, tt.query.AssertCalled(t, "GetMultiple", mock.Anything, tt.args.offset, tt.args.limit))
		})
	}
}
func Test_Create(t *testing.T) {
	type args struct {
		ctx     context.Context
		article entity.User
	}

	tests := []struct {
		desc    string
		cmd     mocks.Eventstore[entity.User]
		args    args
		wantErr bool
	}{
		{
			desc: "Test if method is invoked",
			args: args{
				ctx:     context.Background(),
				article: entity.User{},
			},
			cmd: func() mocks.Eventstore[entity.User] {
				m := mocks.Eventstore[entity.User]{Mock: new(mock.Mock)}
				m.On("Create", mock.Anything, mock.AnythingOfType("entity.User")).Return(nil).Once()
				return m
			}(),
		},
		{
			desc: "Test if an error is forwarded",
			args: args{
				ctx:     context.Background(),
				article: entity.User{},
			},
			wantErr: true,
			cmd: func() mocks.Eventstore[entity.User] {
				m := mocks.Eventstore[entity.User]{Mock: new(mock.Mock)}
				m.On("Create", mock.Anything, mock.AnythingOfType("entity.User")).Return(errors.New("test err")).Once()
				return m
			}(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			db := storage.NewCQRStorage(tt.cmd, mocks.Storage[entity.User]{}, nulls.NullLogger{}, nulls.NullTracer{})
			err := db.Create(tt.args.ctx, tt.args.article)
			if (err != nil) != tt.wantErr {
				t.Errorf("storage.Create():\n error = %+v\n wantErr = %+v\n", err, tt.wantErr)
				return
			}
			assert.True(t, tt.cmd.AssertCalled(t, "Create", mock.Anything, tt.args.article))
		})
	}
}
func Test_Update(t *testing.T) {
	type args struct {
		ctx     context.Context
		article entity.User
	}

	tests := []struct {
		desc    string
		cmd     mocks.Eventstore[entity.User]
		args    args
		wantErr bool
	}{
		{
			desc: "Test if method is invoked",
			args: args{
				ctx:     context.Background(),
				article: entity.User{},
			},

			cmd: func() mocks.Eventstore[entity.User] {
				m := mocks.Eventstore[entity.User]{Mock: new(mock.Mock)}
				m.On("Update", mock.Anything, mock.AnythingOfType("entity.User")).Return(nil).Once()
				return m
			}(),
		},
		{
			desc: "Test if error is forwarded",
			args: args{
				ctx:     context.Background(),
				article: entity.User{},
			},
			wantErr: true,
			cmd: func() mocks.Eventstore[entity.User] {
				m := mocks.Eventstore[entity.User]{Mock: new(mock.Mock)}
				m.On("Update", mock.Anything, mock.AnythingOfType("entity.User")).Return(errors.New("test err")).Once()
				return m
			}(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			db := storage.NewCQRStorage(tt.cmd, mocks.Storage[entity.User]{}, nulls.NullLogger{}, nulls.NullTracer{})
			err := db.Update(tt.args.ctx, tt.args.article)
			if (err != nil) != tt.wantErr {
				t.Errorf("storage.Update():\n error = %+v\n wantErr = %+v\n", err, tt.wantErr)
				return
			}
			assert.True(t, tt.cmd.AssertCalled(t, "Update", mock.Anything, tt.args.article))
		})
	}
}
func Test_Delete(t *testing.T) {
	type args struct {
		ctx context.Context
		id  string
	}

	tests := []struct {
		desc    string
		cmd     mocks.Eventstore[entity.User]
		args    args
		wantErr bool
	}{
		{
			desc: "Test if method is invoked",
			args: args{
				ctx: context.Background(),
				id:  "",
			},

			cmd: func() mocks.Eventstore[entity.User] {
				m := mocks.Eventstore[entity.User]{Mock: new(mock.Mock)}
				m.On("Delete", mock.Anything, mock.AnythingOfType("string")).Return(nil).Once()
				return m
			}(),
		},
		{
			desc: "Test if error is forwarded",
			args: args{
				ctx: context.Background(),
				id:  "",
			},
			wantErr: true,
			cmd: func() mocks.Eventstore[entity.User] {
				m := mocks.Eventstore[entity.User]{Mock: new(mock.Mock)}
				m.On("Delete", mock.Anything, mock.AnythingOfType("string")).Return(errors.New("test err")).Once()
				return m
			}(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			db := storage.NewCQRStorage(tt.cmd, mocks.Storage[entity.User]{}, nulls.NullLogger{}, nulls.NullTracer{})
			err := db.Delete(tt.args.ctx, tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("storage.Delete():\n error = %+v\n wantErr = %+v\n", err, tt.wantErr)
				return
			}
			assert.True(t, tt.cmd.AssertCalled(t, "Delete", mock.Anything, tt.args.id))
			assert.True(t, tt.cmd.AssertExpectations(t))
		})
	}
}
