package server_test

import (
	"context"
	"errors"
	"log"
	"net"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/krixlion/dev_forum-lib/event/dispatcher"
	"github.com/krixlion/dev_forum-lib/mocks"
	"github.com/krixlion/dev_forum-lib/nulls"
	"github.com/krixlion/dev_forum-proto/user_service/pb"
	"github.com/krixlion/dev_forum-user/pkg/entity"
	"github.com/krixlion/dev_forum-user/pkg/grpc/server"
	"github.com/krixlion/dev_forum-user/pkg/storage"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

// setUpServerWithMiddleware initializes and runs in the background a gRPC
// server allowing only for local calls for testing.
// Returns a client to interact with the server.
// The server is shutdown when ctx.Done() receives.
func setUpServerWithMiddleware(ctx context.Context, db storage.Storage) pb.UserServiceClient {
	// bufconn allows the server to call itself
	// great for testing across whole infrastructure
	lis := bufconn.Listen(1024 * 1024)
	bufDialer := func(context.Context, string) (net.Conn, error) {
		return lis.Dial()
	}
	mq := mocks.Broker{Mock: new(mock.Mock)}
	server := server.NewUserServer(db, nulls.NullLogger{}, nulls.NullTracer{}, dispatcher.NewDispatcher(mq, 0))

	s := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			server.ValidateRequestInterceptor(),
		),
	)

	pb.RegisterUserServiceServer(s, server)
	go func() {
		if err := s.Serve(lis); err != nil {
			log.Fatalf("Server exited w mith error: %v", err)
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

func Test_validateCreate(t *testing.T) {
	tests := []struct {
		name    string
		storage mocks.Storage[entity.User]
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

			s := setUpServerWithMiddleware(ctx, tt.storage)

			got, err := s.Create(ctx, tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("UserServer.validateCreate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !cmp.Equal(got, tt.want, cmpopts.EquateApproxTime(time.Second), cmpopts.EquateEmpty()) {
				t.Errorf("UserServer.validateCreate():\n got = %v\n want = %v\n %v", got, tt.want, cmp.Diff(got, tt.want))
			}
		})
	}
}

func Test_validateUpdate(t *testing.T) {
	tests := []struct {
		name    string
		storage mocks.Storage[entity.User]
		req     *pb.UpdateUserRequest
		want    *pb.UpdateUserResponse
		wantErr bool
	}{
		{
			name: "Test if validation fails on invalid email",
			storage: func() mocks.Storage[entity.User] {
				m := mocks.NewStorage[entity.User]()
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

			s := setUpServerWithMiddleware(ctx, tt.storage)

			got, err := s.Update(ctx, tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("UserServer.validateUpdate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !cmp.Equal(got, tt.want, cmpopts.EquateApproxTime(time.Second), cmpopts.EquateEmpty()) {
				t.Errorf("UserServer.validateUpdate():\n got = %v\n want = %v\n %v", got, tt.want, cmp.Diff(got, tt.want))
			}
		})
	}
}

func Test_validateDelete(t *testing.T) {
	tests := []struct {
		name    string
		storage mocks.Storage[entity.User]
		req     *pb.DeleteUserRequest
		wantErr bool
	}{
		{
			name: "Test if returns OK regardless whether user exists or not",
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

			s := setUpServerWithMiddleware(ctx, tt.storage)

			_, err := s.Delete(ctx, tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("UserServer.validateDelete() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
