package server_test

import (
	"context"
	"errors"
	"log"
	"net"
	"testing"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/krixlion/dev_forum-lib/mocks"
	"github.com/krixlion/dev_forum-lib/nulls"
	"github.com/krixlion/dev_forum-user/internal/gentest"
	"github.com/krixlion/dev_forum-user/pkg/entity"
	"github.com/krixlion/dev_forum-user/pkg/grpc/server"
	pb "github.com/krixlion/dev_forum-user/pkg/grpc/v1"
	"github.com/krixlion/dev_forum-user/pkg/storage"
	"github.com/krixlion/dev_forum-user/pkg/storage/storagemocks"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

// setUpServer initializes and runs in the background a gRPC
// server allowing only for local calls for testing.
// Returns a client to interact with the server.
// The server is cancel when ctx.Done() receives.
func setUpServer(ctx context.Context, db storage.Storage, broker mocks.Broker) pb.UserServiceClient {
	// bufconn allows the server to call itself
	// great for testing across whole infrastructure
	lis := bufconn.Listen(1024 * 1024)
	bufDialer := func(context.Context, string) (net.Conn, error) {
		return lis.Dial()
	}

	s := grpc.NewServer()
	server := server.MakeUserServer(server.Dependencies{
		Storage: db,
		Logger:  nulls.NullLogger{},
		Tracer:  nulls.NullTracer{},
		Broker:  broker,
	})
	pb.RegisterUserServiceServer(s, server)
	go func() {
		if err := s.Serve(lis); err != nil {
			log.Fatalf("Server exited with an error: %v", err)
		}
	}()

	conn, err := grpc.NewClient("passthrough://bufnet", grpc.WithContextDialer(bufDialer), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to dial bufnet: %v", err)
	}

	go func() {
		<-ctx.Done()
		s.Stop()
		if err := conn.Close(); err != nil {
			log.Fatalf("Failed to close client conn, err: %v", err)
		}
	}()

	return pb.NewUserServiceClient(conn)
}

func TestUserServer_Get(t *testing.T) {
	v := gentest.RandomUser(2, 5, 5)
	user := &pb.User{
		Id:   v.Id,
		Name: v.Name,
	}

	tests := []struct {
		name    string
		arg     *pb.GetUserRequest
		storage storagemocks.Storage
		broker  mocks.Broker
		want    *pb.GetUserResponse
		wantErr bool
	}{
		{
			name: "Test if response is returned properly on simple request",
			arg:  &pb.GetUserRequest{Id: user.Id},
			storage: func() storagemocks.Storage {
				m := storagemocks.NewStorage()
				m.On("Get", mock.Anything, mock.AnythingOfType("filter.Filter")).Return(v, nil).Once()
				return m
			}(),
			broker: func() mocks.Broker {
				m := mocks.NewBroker()
				m.On("ResilientPublish", mock.AnythingOfType("event.Event")).Return(nil).Once()
				return m
			}(),
			want:    &pb.GetUserResponse{User: user},
			wantErr: false,
		},
		{
			name: "Test if error is returned properly on storage error",
			arg:  &pb.GetUserRequest{Id: ""},
			storage: func() storagemocks.Storage {
				m := storagemocks.NewStorage()
				m.On("Get", mock.Anything, mock.AnythingOfType("filter.Filter")).Return(entity.User{}, errors.New("test err")).Once()
				return m
			}(),
			broker: func() mocks.Broker {
				m := mocks.NewBroker()
				m.On("ResilientPublish", mock.AnythingOfType("event.Event")).Return(nil).Once()
				return m
			}(),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			client := setUpServer(ctx, tt.storage, tt.broker)

			got, err := client.Get(ctx, tt.arg)
			if (err != nil) != tt.wantErr {
				t.Errorf("UserServer.Get():\n error = %v\n wantErr = %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			if !cmp.Equal(got.User, tt.want.User, cmpopts.IgnoreUnexported(pb.User{})) {
				t.Errorf("UserServer.Get():\n got = %v\n want = %v", got, tt.want)
			}
		})
	}
}

func TestUserServer_Create(t *testing.T) {
	v := gentest.RandomUser(2, 5, 5)
	user := &pb.User{
		Id:       v.Id,
		Name:     v.Name,
		Password: v.Password,
		Email:    v.Email,
	}

	tests := []struct {
		name    string
		arg     *pb.CreateUserRequest
		storage storagemocks.Storage
		broker  mocks.Broker
		want    *pb.CreateUserResponse
		wantErr bool
	}{
		{
			name: "Test if response is returned properly on simple request",
			arg:  &pb.CreateUserRequest{User: user},
			storage: func() storagemocks.Storage {
				m := storagemocks.NewStorage()
				m.On("Create", mock.Anything, mock.AnythingOfType("entity.User")).Return(nil).Once()
				return m
			}(),
			broker: func() mocks.Broker {
				m := mocks.NewBroker()
				m.On("ResilientPublish", mock.AnythingOfType("event.Event")).Return(nil).Once()
				return m
			}(),
			want: &pb.CreateUserResponse{Id: user.Id},
		},
		{
			name: "Test if error is returned properly on storage error",
			arg:  &pb.CreateUserRequest{User: user},
			storage: func() storagemocks.Storage {
				m := storagemocks.NewStorage()
				m.On("Create", mock.Anything, mock.AnythingOfType("entity.User")).Return(errors.New("test err")).Once()
				return m
			}(),
			broker: func() mocks.Broker {
				m := mocks.NewBroker()
				m.On("ResilientPublish", mock.AnythingOfType("event.Event")).Return(nil).Once()
				return m
			}(),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			client := setUpServer(ctx, tt.storage, tt.broker)

			got, err := client.Create(ctx, tt.arg)
			if (err != nil) != tt.wantErr {
				t.Errorf("UserServer.Create():\n error = %v\n wantErr = %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			if _, err := uuid.FromString(got.Id); err != nil {
				t.Errorf("UserServer.Create(): failed to parse user id as a uuid:\n id = %v\n err = %v", got.Id, err)
			}
		})
	}
}

func TestUserServer_Update(t *testing.T) {
	v := gentest.RandomUser(2, 5, 5)
	user := &pb.User{
		Id:       v.Id,
		Name:     v.Id,
		Password: v.Password,
		Email:    v.Email,
	}

	tests := []struct {
		name    string
		arg     *pb.UpdateUserRequest
		storage storagemocks.Storage
		broker  mocks.Broker
		wantErr bool
	}{
		{
			name: "Test if response is returned properly on simple request",
			arg:  &pb.UpdateUserRequest{User: user},
			storage: func() storagemocks.Storage {
				m := storagemocks.NewStorage()
				m.On("Update", mock.Anything, mock.AnythingOfType("entity.User")).Return(nil).Once()
				return m
			}(),
			broker: func() mocks.Broker {
				m := mocks.NewBroker()
				m.On("ResilientPublish", mock.AnythingOfType("event.Event")).Return(nil).Once()
				return m
			}(),
			wantErr: false,
		},
		{
			name: "Test if error is returned properly on storage error",
			arg:  &pb.UpdateUserRequest{User: user},
			storage: func() storagemocks.Storage {
				m := storagemocks.NewStorage()
				m.On("Update", mock.Anything, mock.AnythingOfType("entity.User")).Return(errors.New("test err")).Once()
				return m
			}(),
			broker: func() mocks.Broker {
				m := mocks.NewBroker()
				m.On("ResilientPublish", mock.AnythingOfType("event.Event")).Return(nil).Once()
				return m
			}(),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			client := setUpServer(ctx, tt.storage, tt.broker)

			if _, err := client.Update(ctx, tt.arg); (err != nil) != tt.wantErr {
				t.Errorf("UserServer.Update():\n error = %v\n wantErr = %v", err, tt.wantErr)
			}
		})
	}
}

func TestUserServer_Delete(t *testing.T) {
	v := gentest.RandomUser(2, 5, 5)
	user := &pb.User{
		Id: v.Id,
	}

	tests := []struct {
		name    string
		arg     *pb.DeleteUserRequest
		storage storagemocks.Storage
		broker  mocks.Broker
		wantErr bool
	}{
		{
			name: "Test if response is returned properly on simple request",
			arg:  &pb.DeleteUserRequest{Id: user.Id},
			storage: func() storagemocks.Storage {
				m := storagemocks.NewStorage()
				m.On("Delete", mock.Anything, mock.AnythingOfType("string")).Return(nil).Once()
				return m
			}(),
			broker: func() mocks.Broker {
				m := mocks.NewBroker()
				m.On("ResilientPublish", mock.AnythingOfType("event.Event")).Return(nil).Once()
				return m
			}(),
			wantErr: false,
		},
		{
			name: "Test if error is returned properly on storage error",
			arg:  &pb.DeleteUserRequest{Id: user.Id},
			storage: func() storagemocks.Storage {
				m := storagemocks.NewStorage()
				m.On("Delete", mock.Anything, mock.AnythingOfType("string")).Return(errors.New("test err")).Once()
				return m
			}(),
			broker: func() mocks.Broker {
				m := mocks.NewBroker()
				m.On("ResilientPublish", mock.AnythingOfType("event.Event")).Return(nil).Once()
				return m
			}(),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			client := setUpServer(ctx, tt.storage, tt.broker)

			if _, err := client.Delete(ctx, tt.arg); (err != nil) != tt.wantErr {
				t.Errorf("UserServer.Delete():\n error = %v\n wantErr = %v", err, tt.wantErr)
			}
		})
	}
}

func TestUserServer_GetStream(t *testing.T) {
	var users []entity.User
	for i := 0; i < 5; i++ {
		user := gentest.RandomUser(2, 5, 5)
		users = append(users, user)
	}

	var pbUsers []*pb.User
	for _, v := range users {
		pbUser := &pb.User{
			Id:   v.Id,
			Name: v.Name,
		}
		pbUsers = append(pbUsers, pbUser)
	}

	tests := []struct {
		name    string
		arg     *pb.GetUsersRequest
		storage storagemocks.Storage
		broker  mocks.Broker
		want    []*pb.User
		wantErr bool
	}{
		{
			name: "Test if response is returned properly on simple request",
			arg:  &pb.GetUsersRequest{Offset: "0", Limit: "5"},
			storage: func() storagemocks.Storage {
				m := storagemocks.NewStorage()
				m.On("GetMultiple", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("filter.Filter")).Return(users, nil).Once()
				return m
			}(),
			broker: func() mocks.Broker {
				m := mocks.NewBroker()
				m.On("ResilientPublish", mock.AnythingOfType("event.Event")).Return(nil).Once()
				return m
			}(),
			want:    pbUsers,
			wantErr: false,
		},
		{
			name: "Test if error is returned properly on storage error",
			arg:  &pb.GetUsersRequest{Offset: "n/a", Limit: "n/a", Filter: "n/a"},
			storage: func() storagemocks.Storage {
				m := storagemocks.NewStorage()
				m.On("GetMultiple", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("filter.Filter")).Return([]entity.User{}, errors.New("test err")).Once()
				return m
			}(),
			broker: func() mocks.Broker {
				m := mocks.NewBroker()
				m.On("ResilientPublish", mock.AnythingOfType("event.Event")).Return(nil).Once()
				return m
			}(),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			client := setUpServer(ctx, tt.storage, tt.broker)

			stream, err := client.GetStream(ctx, tt.arg)
			if err != nil {
				t.Errorf("UserServer.GetStream(): failed to init stream:\n error = %v\n", err)
				return
			}

			var got []*pb.User
			for i := 0; i < len(tt.want); i++ {
				user, err := stream.Recv()
				if (err != nil) != tt.wantErr {
					t.Errorf("UserServer.GetStream():\n error = %v\n wantErr = %v", err, tt.wantErr)
					return
				}
				got = append(got, user)
			}

			if tt.wantErr {
				return
			}

			if !cmp.Equal(got, tt.want, cmpopts.IgnoreUnexported(pb.User{})) {
				t.Errorf("UserServer.GetStream():\n got = %v\n want = %v", got, tt.want)
			}
		})
	}
}
