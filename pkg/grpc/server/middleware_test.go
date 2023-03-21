package server

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/krixlion/dev_forum-lib/event"
	"github.com/krixlion/dev_forum-lib/event/dispatcher"
	"github.com/krixlion/dev_forum-lib/mocks"
	"github.com/krixlion/dev_forum-lib/nulls"
	"github.com/krixlion/dev_forum-user/pkg/entity"
	pb "github.com/krixlion/dev_forum-user/pkg/grpc/v1"
	"github.com/krixlion/dev_forum-user/pkg/storage"
	"github.com/stretchr/testify/mock"
)

func setUpServerWithMiddleware(ctx context.Context, db storage.Storage, mq event.Broker) UserServer {
	s := NewUserServer(Dependencies{
		Storage:    db,
		Logger:     nulls.NullLogger{},
		Tracer:     nulls.NullTracer{},
		Dispatcher: dispatcher.NewDispatcher(mq, 0),
	})

	return s
}

func Test_validateCreate(t *testing.T) {
	tests := []struct {
		name    string
		handler mocks.UnaryHandler
		storage mocks.Storage[entity.User]
		broker  mocks.Broker
		req     *pb.CreateUserRequest
		want    *pb.CreateUserResponse
		wantErr bool
	}{
		{
			name: "Test if validation fails on invalid email",
			storage: func() mocks.Storage[entity.User] {
				m := mocks.NewStorage[entity.User]()
				return m
			}(),
			handler: func() mocks.UnaryHandler {
				m := mocks.NewUnaryHandler()
				return m
			}(),
			broker: func() mocks.Broker {
				m := mocks.NewBroker()
				m.On("ResilientPublish", mock.AnythingOfType("event.Event")).Return(nil).Once()
				return m
			}(),
			req: &pb.CreateUserRequest{
				User: &pb.User{
					Id:    "Id",
					Name:  "Name",
					Email: "invalid email",
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			s := setUpServerWithMiddleware(ctx, tt.storage, tt.broker)

			got, err := s.validateCreate(ctx, tt.req, tt.handler.GetMock())
			if (err != nil) != tt.wantErr {
				t.Errorf("UserServer.validateCreate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !cmp.Equal(got, tt.want, cmpopts.EquateApproxTime(time.Second), cmpopts.EquateEmpty()) && !tt.wantErr {
				t.Errorf("UserServer.validateCreate():\n got = %v\n want = %v\n %v", got, tt.want, cmp.Diff(got, tt.want))
			}
		})
	}
}

func Test_validateUpdate(t *testing.T) {
	tests := []struct {
		name    string
		handler mocks.UnaryHandler
		storage mocks.Storage[entity.User]
		broker  mocks.Broker
		req     *pb.UpdateUserRequest
		want    *empty.Empty
		wantErr bool
	}{
		{
			name: "Test if validation fails on invalid email",
			storage: func() mocks.Storage[entity.User] {
				m := mocks.NewStorage[entity.User]()
				return m
			}(),
			handler: func() mocks.UnaryHandler {
				m := mocks.NewUnaryHandler()
				return m
			}(),
			broker: func() mocks.Broker {
				m := mocks.NewBroker()
				m.On("ResilientPublish", mock.AnythingOfType("event.Event")).Return(nil).Once()
				return m
			}(),
			req: &pb.UpdateUserRequest{
				User: &pb.User{
					Id:    "Id",
					Name:  "Name",
					Email: "Invalid email",
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			s := setUpServerWithMiddleware(ctx, tt.storage, tt.broker)

			got, err := s.validateUpdate(ctx, tt.req, tt.handler.GetMock())
			if (err != nil) != tt.wantErr {
				t.Errorf("UserServer.validateUpdate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !cmp.Equal(got, tt.want, cmpopts.EquateApproxTime(time.Second), cmpopts.EquateEmpty()) && !tt.wantErr {
				t.Errorf("UserServer.validateUpdate():\n got = %v\n want = %v\n %v", got, tt.want, cmp.Diff(got, tt.want))
			}
		})
	}
}

func Test_validateDelete(t *testing.T) {
	tests := []struct {
		name    string
		handler mocks.UnaryHandler
		storage mocks.Storage[entity.User]
		broker  mocks.Broker
		req     *pb.DeleteUserRequest
		wantErr bool
	}{
		{
			name: "Test if returns OK regardless whether user exists or not",
			broker: func() mocks.Broker {
				m := mocks.NewBroker()
				m.On("ResilientPublish", mock.AnythingOfType("event.Event")).Return(nil).Once()
				return m
			}(),
			handler: func() mocks.UnaryHandler {
				m := mocks.NewUnaryHandler()
				return m
			}(),
			storage: func() mocks.Storage[entity.User] {
				m := mocks.NewStorage[entity.User]()
				m.On("Get", mock.Anything, mock.AnythingOfType("string")).Return(entity.User{}, errors.New("not found")).Once()
				return m
			}(),

			req: &pb.DeleteUserRequest{
				Id: "id",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			s := setUpServerWithMiddleware(ctx, tt.storage, tt.broker)

			_, err := s.validateDelete(ctx, tt.req, tt.handler.GetMock())
			if (err != nil) != tt.wantErr {
				t.Errorf("UserServer.validateDelete() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
