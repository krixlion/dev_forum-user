package mocks

import (
	"context"

	pb "github.com/krixlion/dev_forum-user/pkg/grpc/v1"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

var _ pb.UserServiceClient = (*UserClient)(nil)

type UserClient struct {
	*mock.Mock
}

func NewUserClient() UserClient {
	return UserClient{
		Mock: new(mock.Mock),
	}
}

func (m UserClient) Create(ctx context.Context, in *pb.CreateUserRequest, opts ...grpc.CallOption) (*pb.CreateUserResponse, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*pb.CreateUserResponse), args.Error(1)
}

func (m UserClient) Update(ctx context.Context, in *pb.UpdateUserRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*emptypb.Empty), args.Error(1)
}

func (m UserClient) Delete(ctx context.Context, in *pb.DeleteUserRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*emptypb.Empty), args.Error(1)
}

func (m UserClient) Get(ctx context.Context, in *pb.GetUserRequest, opts ...grpc.CallOption) (*pb.GetUserResponse, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*pb.GetUserResponse), args.Error(1)
}

func (m UserClient) GetSecret(ctx context.Context, in *pb.GetUserSecretRequest, opts ...grpc.CallOption) (*pb.GetUserSecretResponse, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*pb.GetUserSecretResponse), args.Error(1)
}

func (m UserClient) GetStream(ctx context.Context, in *pb.GetUsersRequest, opts ...grpc.CallOption) (pb.UserService_GetStreamClient, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(pb.UserService_GetStreamClient), args.Error(1)
}
