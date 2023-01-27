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
	"github.com/krixlion/dev_forum-proto/user_service/pb"
	"github.com/krixlion/dev_forum-user/pkg/entity"
	"github.com/krixlion/dev_forum-user/pkg/helpers/gentest"
	"github.com/krixlion/dev_forum-user/pkg/net/grpc/server"
	"github.com/krixlion/dev_forum-user/pkg/storage"
	"github.com/stretchr/testify/mock"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

// setUpServer initializes and runs in the background a gRPC
// server allowing only for local calls for testing.
// Returns a client to interact with the server.
// The server is shutdown when ctx.Done() receives.
func setUpServer(ctx context.Context, mock storage.Storage) pb.UserServiceClient {
	// bufconn allows the server to call itself
	// great for testing across whole infrastructure
	lis := bufconn.Listen(1024 * 1024)
	bufDialer := func(context.Context, string) (net.Conn, error) {
		return lis.Dial()
	}

	s := grpc.NewServer()
	server := server.UserServer{
		Storage: mock,
	}
	pb.RegisterUserServiceServer(s, server)
	go func() {
		if err := s.Serve(lis); err != nil {
			log.Fatalf("Server exited with error: %v", err)
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

func Test_Get(t *testing.T) {
	v := gentest.RandomUser(2, 5, 5)
	user := &pb.User{
		Id:   v.Id,
		Name: v.Name,
	}

	testCases := []struct {
		desc    string
		arg     *pb.GetUserRequest
		want    *pb.GetUserResponse
		wantErr bool
		storage storage.CQRStorage
	}{
		{
			desc: "Test if response is returned properly on simple request",
			arg: &pb.GetUserRequest{
				Id: user.Id,
			},
			want: &pb.GetUserResponse{
				User: user,
			},
			storage: func() (m mockStorage) {
				m.On("Get", mock.Anything, mock.AnythingOfType("string")).Return(v, nil).Times(1)
				return
			}(),
		},
		{
			desc: "Test if error is returned properly on storage error",
			arg: &pb.GetUserRequest{
				Id: "",
			},
			want:    nil,
			wantErr: true,
			storage: func() (m mockStorage) {
				m.On("Get", mock.Anything, mock.AnythingOfType("string")).Return(entity.User{}, errors.New("test err")).Times(1)
				return
			}(),
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			ctx, shutdown := context.WithCancel(context.Background())
			defer shutdown()

			client := setUpServer(ctx, tC.storage)

			ctx, cancel := context.WithTimeout(ctx, time.Second)
			defer cancel()

			getResponse, err := client.Get(ctx, tC.arg)
			if (err != nil) != tC.wantErr {
				t.Errorf("Failed to Get User, err: %v", err)
				return
			}

			// Equals false if both are nil or they point to the same memory address
			// so be sure to use seperate structs when providing args in order to prevent SEGV.
			if getResponse != tC.want {
				if !cmp.Equal(getResponse.User, tC.want.User, cmpopts.IgnoreUnexported(pb.User{})) {
					t.Errorf("Users are not equal:\n Got = %+v\n, want = %+v\n", getResponse.User, tC.want.User)
					return
				}
			}
		})
	}
}

func Test_Create(t *testing.T) {
	v := gentest.RandomUser(2, 5, 5)
	User := &pb.User{
		Id:       v.Id,
		Name:     v.Name,
		Password: v.Password,
		Email:    v.Email,
	}

	testCases := []struct {
		desc     string
		arg      *pb.CreateUserRequest
		dontWant *pb.CreateUserResponse
		wantErr  bool
		storage  storage.CQRStorage
	}{
		{
			desc: "Test if response is returned properly on simple request",
			arg: &pb.CreateUserRequest{
				User: User,
			},
			dontWant: &pb.CreateUserResponse{
				Id: User.Id,
			},
			storage: func() (m mockStorage) {
				m.On("Create", mock.Anything, mock.AnythingOfType("entity.User")).Return(nil).Times(1)
				return
			}(),
		},
		{
			desc: "Test if error is returned properly on storage error",
			arg: &pb.CreateUserRequest{
				User: User,
			},
			dontWant: nil,
			wantErr:  true,
			storage: func() (m mockStorage) {
				m.On("Create", mock.Anything, mock.AnythingOfType("entity.User")).Return(errors.New("test err")).Times(1)
				return
			}(),
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			ctx, shutdown := context.WithCancel(context.Background())
			defer shutdown()
			client := setUpServer(ctx, tC.storage)

			createResponse, err := client.Create(ctx, tC.arg)
			if (err != nil) != tC.wantErr {
				t.Errorf("Failed to Get User, err: %v", err)
				return
			}

			// Equals false if both are nil or they point to the same memory address
			// so be sure to use seperate structs when providing args in order to prevent SEGV.
			if createResponse != tC.dontWant {
				if cmp.Equal(createResponse.Id, tC.dontWant.Id) {
					t.Errorf("User IDs was not reassigned:\n Got = %+v\n want = %+v\n", createResponse.Id, tC.dontWant.Id)
					return
				}
				if _, err := uuid.FromString(createResponse.Id); err != nil {
					t.Errorf("User ID is not correct UUID:\n ID = %+v\n err = %+v", createResponse.Id, err)
					return
				}
			}
		})
	}
}

// func Test_Update(t *testing.T) {
// 	v := gentest.RandomUser(2, 5, 5)
// 	User := &pb.User{
// 		Id:       v.Id,
// 		Name:     v.Id,
// 		Password: v.Title,
// 		Email:    v.Body,
// 	}

// 	testCases := []struct {
// 		desc    string
// 		arg     *pb.UpdateUserRequest
// 		want    *pb.UpdateUserResponse
// 		wantErr bool
// 		storage storage.CQRStorage
// 	}{
// 		{
// 			desc: "Test if response is returned properly on simple request",
// 			arg: &pb.UpdateUserRequest{
// 				User: User,
// 			},
// 			want: &pb.UpdateUserResponse{},
// 			storage: func() (m mockStorage) {
// 				m.On("Update", mock.Anything, mock.AnythingOfType("entity.User")).Return(nil).Times(1)
// 				return
// 			}(),
// 		},
// 		{
// 			desc: "Test if error is returned properly on storage error",
// 			arg: &pb.UpdateUserRequest{
// 				User: User,
// 			},
// 			want:    nil,
// 			wantErr: true,
// 			storage: func() (m mockStorage) {
// 				m.On("Update", mock.Anything, mock.AnythingOfType("entity.User")).Return(errors.New("test err")).Times(1)
// 				return
// 			}(),
// 		},
// 	}
// 	for _, tC := range testCases {
// 		t.Run(tC.desc, func(t *testing.T) {
// 			ctx, shutdown := context.WithCancel(context.Background())
// 			defer shutdown()
// 			client := setUpServer(ctx, tC.storage)

// 			got, err := client.Update(ctx, tC.arg)
// 			if (err != nil) != tC.wantErr {
// 				t.Errorf("Failed to Update User, err: %v", err)
// 				return
// 			}

// 			// Equals false if both are nil or they point to the same memory address
// 			// so be sure to use seperate structs when providing args in order to prevent SEGV.
// 			if got != tC.want {
// 				if !cmp.Equal(got, tC.want, cmpopts.IgnoreUnexported(pb.UpdateUserResponse{})) {
// 					t.Errorf("Wrong response:\n got = %+v\n want = %+v\n", got, tC.want)
// 					return
// 				}
// 			}
// 		})
// 	}
// }

func Test_Delete(t *testing.T) {
	v := gentest.RandomUser(2, 5, 5)
	User := &pb.User{
		Id:       v.Id,
		Name:     v.Name,
		Password: v.Password,
		Email:    v.Email,
	}

	testCases := []struct {
		desc    string
		arg     *pb.DeleteUserRequest
		want    *pb.DeleteUserResponse
		wantErr bool
		storage storage.CQRStorage
	}{
		{
			desc: "Test if response is returned properly on simple request",
			arg: &pb.DeleteUserRequest{
				Id: User.Id,
			},
			want: &pb.DeleteUserResponse{},
			storage: func() (m mockStorage) {
				m.On("Delete", mock.Anything, mock.AnythingOfType("string")).Return(nil).Times(1)
				return
			}(),
		},
		{
			desc: "Test if error is returned properly on storage error",
			arg: &pb.DeleteUserRequest{
				Id: User.Id,
			},
			want:    nil,
			wantErr: true,
			storage: func() (m mockStorage) {
				m.On("Delete", mock.Anything, mock.AnythingOfType("string")).Return(errors.New("test err")).Times(1)
				return
			}(),
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			ctx, shutdown := context.WithCancel(context.Background())
			defer shutdown()
			client := setUpServer(ctx, tC.storage)

			got, err := client.Delete(ctx, tC.arg)
			if (err != nil) != tC.wantErr {
				t.Errorf("Failed to Delete User, err: %v", err)
				return
			}

			// Equals false if both are nil or they point to the same memory address
			// so be sure to use seperate structs when providing args in order to prevent SEGV.
			if got != tC.want {
				if !cmp.Equal(got, tC.want, cmpopts.IgnoreUnexported(pb.DeleteUserResponse{})) {
					t.Errorf("Wrong response:\n got = %+v\n want = %+v\n", got, tC.want)
					return
				}
			}
		})
	}
}

func Test_GetStream(t *testing.T) {
	var Users []entity.User
	for i := 0; i < 5; i++ {
		User := gentest.RandomUser(2, 5, 5)
		Users = append(Users, User)
	}

	var pbUsers []*pb.User
	for _, v := range Users {
		pbUser := &pb.User{
			Id:       v.Id,
			Name:     v.Name,
			Password: v.Password,
			Email:    v.Email,
		}
		pbUsers = append(pbUsers, pbUser)
	}

	testCases := []struct {
		desc    string
		arg     *pb.GetUsersRequest
		want    []*pb.User
		wantErr bool
		storage storage.CQRStorage
	}{
		{
			desc: "Test if response is returned properly on simple request",
			arg: &pb.GetUsersRequest{
				Offset: "0",
				Limit:  "5",
			},
			want: pbUsers,
			storage: func() (m mockStorage) {
				m.On("GetMultiple", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(Users, nil).Times(1)
				return
			}(),
		},
		{
			desc:    "Test if error is returned properly on storage error",
			arg:     &pb.GetUsersRequest{},
			want:    nil,
			wantErr: true,
			storage: func() (m mockStorage) {
				m.On("GetMultiple", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return([]entity.User{}, errors.New("test err")).Times(1)
				return
			}(),
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			ctx, shutdown := context.WithCancel(context.Background())
			defer shutdown()
			client := setUpServer(ctx, tC.storage)

			stream, err := client.GetStream(ctx, tC.arg)
			if err != nil {
				t.Errorf("Failed to Get stream, err: %v", err)
				return
			}

			var got []*pb.User
			for i := 0; i < len(tC.want); i++ {
				User, err := stream.Recv()
				if (err != nil) != tC.wantErr {
					t.Errorf("Failed to receive User from stream, err: %v", err)
					return
				}
				got = append(got, User)
			}

			if !cmp.Equal(got, tC.want, cmpopts.IgnoreUnexported(pb.User{})) {
				t.Errorf("Users are not equal:\n Got = %+v\n want = %+v\n", got, tC.want)
				return
			}
		})
	}
}
