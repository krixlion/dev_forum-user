package server

import (
	"context"

	"github.com/krixlion/dev_forum-proto/user_service/pb"
	"google.golang.org/grpc"
)

func (s UserServer) ValidateRequestInterceptor() grpc.UnaryServerInterceptor {

	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		switch info.FullMethod {
		case "/UserService/Create":
			return s.validateCreate(ctx, req.(pb.CreateUserRequest), handler)
		case "/UserService/Update":
			return s.validateUpdate(ctx, req.(pb.CreateUserRequest), handler)
		case "/UserService/Delete":
			return s.validateDelete(ctx, req.(pb.CreateUserRequest), handler)
		default:
			return handler(ctx, req)
		}
	}
}

func (s UserServer) validateCreate(ctx context.Context, req pb.CreateUserRequest, handler grpc.UnaryHandler) (interface{}, error) {
	return handler(ctx, req)
}

func (s UserServer) validateUpdate(ctx context.Context, req pb.CreateUserRequest, handler grpc.UnaryHandler) (interface{}, error) {
	return handler(ctx, req)
}

func (s UserServer) validateDelete(ctx context.Context, req pb.CreateUserRequest, handler grpc.UnaryHandler) (interface{}, error) {
	return handler(ctx, req)
}
