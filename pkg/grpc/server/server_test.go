package server_test

import (
	"context"
	"errors"
	"log"
	"net"
	"testing"
	"time"

	"github.com/gofrs/uuid"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/krixlion/dev_forum-lib/event/dispatcher"
	"github.com/krixlion/dev_forum-lib/mocks"
	"github.com/krixlion/dev_forum-lib/nulls"
	"github.com/krixlion/dev_forum-user/internal/gentest"
	"github.com/krixlion/dev_forum-user/pkg/entity"
	"github.com/krixlion/dev_forum-user/pkg/grpc/server"
	pb "github.com/krixlion/dev_forum-user/pkg/grpc/v1"
	"github.com/krixlion/dev_forum-user/pkg/storage"
	"github.com/krixlion/dev_forum-user/pkg/storage/storagemocks"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/types/known/emptypb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

// setUpServer initializes and runs in the background a gRPC
// server allowing only for local calls for testing.
// Returns a client to interact with the server.
// The server is shutdown when ctx.Done() receives.
func setUpServer(ctx context.Context, db storage.Storage, broker mocks.Broker) pb.UserServiceClient {
	// bufconn allows the server to call itself
	// great for testing across whole infrastructure
	lis := bufconn.Listen(1024 * 1024)
	bufDialer := func(context.Context, string) (net.Conn, error) {
		return lis.Dial()
	}

	s := grpc.NewServer()
	server := server.MakeUserServer(server.Dependencies{
		Storage:    db,
		Logger:     nulls.NullLogger{},
		Tracer:     nulls.NullTracer{},
		Broker:     broker,
		Dispatcher: dispatcher.NewDispatcher(0),
	})
	pb.RegisterUserServiceServer(s, server)
	go func() {
		if err := s.Serve(lis); err != nil {
			log.Fatalf("Server exited with an error: %v", err)
		}
	}()

	go func() {
		<-ctx.Done()
		s.Stop()
	}()

	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(bufDialer), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to dial bufnet: %v", err)
	}

	client := pb.NewUserServiceClient(conn)
	return client
}

func TestUserServer_Get(t *testing.T) {
	v := gentest.RandomUser(2, 5, 5)
	user := &pb.User{
		Id:   v.Id,
		Name: v.Name,
	}

	tests := []struct {
		desc    string
		arg     *pb.GetUserRequest
		want    *pb.GetUserResponse
		wantErr bool
		storage storagemocks.Storage
		broker  mocks.Broker
	}{
		{
			desc: "Test if response is returned properly on simple request",
			arg: &pb.GetUserRequest{
				Id: user.Id,
			},
			want: &pb.GetUserResponse{
				User: user,
			},
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
		},
		{
			desc: "Test if error is returned properly on storage error",
			arg: &pb.GetUserRequest{
				Id: "",
			},
			want:    nil,
			wantErr: true,
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
		},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			ctx, shutdown := context.WithCancel(context.Background())
			defer shutdown()

			client := setUpServer(ctx, tt.storage, tt.broker)

			ctx, cancel := context.WithTimeout(ctx, time.Second)
			defer cancel()

			getResponse, err := client.Get(ctx, tt.arg)
			if (err != nil) != tt.wantErr {
				t.Errorf("Failed to Get User, err: %v", err)
				return
			}

			// Equals false if both are nil or they point to the same memory address
			// so be sure to use seperate structs when providing args in order to prevent SEGV.
			if getResponse != tt.want {
				if !cmp.Equal(getResponse.User, tt.want.User, cmpopts.IgnoreUnexported(pb.User{})) {
					t.Errorf("Users are not equal:\n Got = %+v\n, want = %+v\n", getResponse.User, tt.want.User)
					return
				}
			}
		})
	}
}

func TestUserServer_Create(t *testing.T) {
	v := gentest.RandomUser(2, 5, 5)
	User := &pb.User{
		Id:       v.Id,
		Name:     v.Name,
		Password: v.Password,
		Email:    v.Email,
	}

	tests := []struct {
		desc     string
		arg      *pb.CreateUserRequest
		dontWant *pb.CreateUserResponse
		wantErr  bool
		storage  storagemocks.Storage
		broker   mocks.Broker
	}{
		{
			desc: "Test if response is returned properly on simple request",
			arg: &pb.CreateUserRequest{
				User: User,
			},
			dontWant: &pb.CreateUserResponse{
				Id: User.Id,
			},
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
		},
		{
			desc: "Test if error is returned properly on storage error",
			arg: &pb.CreateUserRequest{
				User: User,
			},
			dontWant: nil,
			wantErr:  true,
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
		},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			ctx, shutdown := context.WithCancel(context.Background())
			defer shutdown()
			client := setUpServer(ctx, tt.storage, tt.broker)

			createResponse, err := client.Create(ctx, tt.arg)
			if err != nil {
				tt.broker.AssertNumberOfCalls(t, "ResilientPublish", 0)
				if !tt.wantErr {
					t.Errorf("Failed to Get User, err: %v", err)
					return
				}
			} else {
				tt.broker.AssertNumberOfCalls(t, "ResilientPublish", 1)
			}

			tt.storage.AssertNumberOfCalls(t, "Create", 1)

			// Equals false if both are nil or point to the same memory address
			// so be sure to use seperate variables when providing args in order to prevent SEGV.
			if createResponse != tt.dontWant {
				if _, err := uuid.FromString(createResponse.Id); err != nil {
					t.Errorf("User ID is not correct UUID:\n ID = %+v\n err = %+v", createResponse.Id, err)
					return
				}
			}
		})
	}
}

func TestUserServer_Update(t *testing.T) {
	v := gentest.RandomUser(2, 5, 5)
	User := &pb.User{
		Id:       v.Id,
		Name:     v.Id,
		Password: v.Password,
		Email:    v.Email,
	}

	tests := []struct {
		desc    string
		arg     *pb.UpdateUserRequest
		want    *emptypb.Empty
		wantErr bool
		storage storagemocks.Storage
		broker  mocks.Broker
	}{
		{
			desc: "Test if response is returned properly on simple request",
			arg: &pb.UpdateUserRequest{
				User: User,
			},
			want: &emptypb.Empty{},
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
		},
		{
			desc: "Test if error is returned properly on storage error",
			arg: &pb.UpdateUserRequest{
				User: User,
			},
			want:    nil,
			wantErr: true,
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
		},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			ctx, shutdown := context.WithCancel(context.Background())
			defer shutdown()
			client := setUpServer(ctx, tt.storage, tt.broker)

			got, err := client.Update(ctx, tt.arg)
			if err != nil {
				tt.broker.AssertNumberOfCalls(t, "ResilientPublish", 0)
				if !tt.wantErr {
					t.Errorf("Failed to Update User, err: %v", err)
					return
				}
			} else {
				tt.broker.AssertNumberOfCalls(t, "ResilientPublish", 1)
			}

			tt.storage.AssertNumberOfCalls(t, "Update", 1)
			// Equals false if both are nil or they point to the same memory address
			// so be sure to use seperate structs when providing args in order to prevent SEGV.
			if got != tt.want {
				if !cmp.Equal(got, tt.want, cmpopts.IgnoreUnexported(emptypb.Empty{})) {
					t.Errorf("Wrong response:\n got = %+v\n want = %+v\n", got, tt.want)
					return
				}
			}
		})
	}
}

func TestUserServer_Delete(t *testing.T) {
	v := gentest.RandomUser(2, 5, 5)
	User := &pb.User{
		Id:       v.Id,
		Name:     v.Name,
		Password: v.Password,
		Email:    v.Email,
	}

	tests := []struct {
		desc    string
		arg     *pb.DeleteUserRequest
		want    *emptypb.Empty
		wantErr bool
		storage storagemocks.Storage
		broker  mocks.Broker
	}{
		{
			desc: "Test if response is returned properly on simple request",
			arg: &pb.DeleteUserRequest{
				Id: User.Id,
			},
			want: &emptypb.Empty{},
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
		},
		{
			desc: "Test if error is returned properly on storage error",
			arg: &pb.DeleteUserRequest{
				Id: User.Id,
			},
			want:    nil,
			wantErr: true,
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
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			ctx, shutdown := context.WithCancel(context.Background())
			defer shutdown()
			client := setUpServer(ctx, tt.storage, tt.broker)

			got, err := client.Delete(ctx, tt.arg)
			if err != nil {
				tt.broker.AssertNumberOfCalls(t, "ResilientPublish", 0)

				if !tt.wantErr {
					t.Errorf("Failed to Delete User, err: %v", err)
					return
				}
			} else {
				tt.broker.AssertNumberOfCalls(t, "ResilientPublish", 1)
			}
			tt.storage.AssertNumberOfCalls(t, "Delete", 1)

			if !cmp.Equal(got, tt.want, cmpopts.IgnoreUnexported(emptypb.Empty{})) {
				t.Errorf("Wrong response:\n got = %+v\n want = %+v\n", got, tt.want)
				return
			}
		})
	}
}

func TestUserServer_GetStream(t *testing.T) {
	var Users []entity.User
	for i := 0; i < 5; i++ {
		User := gentest.RandomUser(2, 5, 5)
		Users = append(Users, User)
	}

	var pbUsers []*pb.User
	for _, v := range Users {
		pbUser := &pb.User{
			Id:   v.Id,
			Name: v.Name,
		}
		pbUsers = append(pbUsers, pbUser)
	}

	tests := []struct {
		desc    string
		arg     *pb.GetUsersRequest
		want    []*pb.User
		wantErr bool
		storage storagemocks.Storage
		broker  mocks.Broker
	}{
		{
			desc: "Test if response is returned properly on simple request",
			arg: &pb.GetUsersRequest{
				Offset: "0",
				Limit:  "5",
			},
			want: pbUsers,
			storage: func() storagemocks.Storage {
				m := storagemocks.NewStorage()
				m.On("GetMultiple", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("filter.Filter")).Return(Users, nil).Once()
				return m
			}(),
			broker: func() mocks.Broker {
				m := mocks.NewBroker()
				m.On("ResilientPublish", mock.AnythingOfType("event.Event")).Return(nil).Once()
				return m
			}(),
		},
		{
			desc:    "Test if error is returned properly on storage error",
			arg:     &pb.GetUsersRequest{},
			want:    nil,
			wantErr: true,
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
		},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			ctx, shutdown := context.WithCancel(context.Background())
			defer shutdown()
			client := setUpServer(ctx, tt.storage, tt.broker)

			stream, err := client.GetStream(ctx, tt.arg)
			if err != nil {
				t.Errorf("Failed to Get stream, err: %v", err)
				return
			}

			var got []*pb.User
			for i := 0; i < len(tt.want); i++ {
				User, err := stream.Recv()
				if (err != nil) != tt.wantErr {
					t.Errorf("Failed to receive User from stream, err: %v", err)
					return
				}
				got = append(got, User)
			}

			if !cmp.Equal(got, tt.want, cmpopts.IgnoreUnexported(pb.User{})) {
				t.Errorf("Users are not equal:\n Got = %+v\n want = %+v\n", got, tt.want)
				return
			}
		})
	}
}
