package mocks

import (
	"context"

	pb "github.com/krixlion/dev_forum-user/pkg/grpc/v1"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/metadata"
)

var _ pb.UserService_GetStreamClient = (*UserStreamClient)(nil)

type UserStreamClient struct {
	*mock.Mock
}

func NewUserStreamClient() UserStreamClient {
	return UserStreamClient{
		new(mock.Mock),
	}
}

func (m UserStreamClient) Recv() (*pb.User, error) {
	returnVals := m.Called()
	return returnVals.Get(0).(*pb.User), returnVals.Error(1)
}

func (m UserStreamClient) CloseSend() error {
	returnVals := m.Called()
	return returnVals.Error(0)
}

func (m UserStreamClient) Header() (metadata.MD, error) {
	returnVals := m.Called()
	return returnVals.Get(0).(metadata.MD), returnVals.Error(1)
}

func (m UserStreamClient) Trailer() metadata.MD {
	returnVals := m.Called()
	return returnVals.Get(0).(metadata.MD)
}

func (m UserStreamClient) Context() context.Context {
	returnVals := m.Called()
	return returnVals.Get(0).(context.Context)
}

func (m UserStreamClient) SendMsg(msg any) error {
	returnVals := m.Called(msg)
	return returnVals.Error(0)
}

func (m UserStreamClient) RecvMsg(msg any) error {
	returnVals := m.Called(msg)
	return returnVals.Error(0)
}
