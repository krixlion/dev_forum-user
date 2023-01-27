package server

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/krixlion/dev_forum-proto/user_service/pb"
	"github.com/krixlion/dev_forum-user/pkg/entity"
	"github.com/krixlion/dev_forum-user/pkg/logging"
	"github.com/krixlion/dev_forum-user/pkg/storage"

	"github.com/gofrs/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UserServer struct {
	pb.UnimplementedUserServiceServer
	Storage storage.Storage
	Logger  logging.Logger
}

func (s UserServer) Close() error {
	var errMsg string

	err := s.Storage.Close()
	if err != nil {
		errMsg = fmt.Sprintf("%s, failed to close storage: %s", errMsg, err)
	}

	if errMsg != "" {
		return errors.New(errMsg)
	}

	return nil
}

func (s UserServer) Create(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	user := entity.UserFromPB(req.GetUser())
	id, err := uuid.NewV4()
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	// Assign new UUID to new user.
	user.Id = id.String()

	if err = s.Storage.Create(ctx, user); err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	return &pb.CreateUserResponse{
		Id: id.String(),
	}, nil
}

func (s UserServer) Delete(ctx context.Context, req *pb.DeleteUserRequest) (*pb.DeleteUserResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	if err := s.Storage.Delete(ctx, req.GetId()); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	return &pb.DeleteUserResponse{}, nil
}

func (s UserServer) Update(ctx context.Context, req *pb.UpdateUserRequest) (*pb.UpdateUserResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	user := entity.UserFromPB(req.GetUser())

	if err := s.Storage.Update(ctx, user); err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	return &pb.UpdateUserResponse{}, nil
}

func (s UserServer) Get(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	user, err := s.Storage.Get(ctx, req.GetId())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to get user: %v", err)
	}

	return &pb.GetUserResponse{
		User: &pb.User{
			Id:   user.Id,
			Name: user.Name,
		},
	}, nil
}

func (s UserServer) GetSecret(ctx context.Context, req *pb.GetUserSecretRequest) (*pb.GetUserSecretResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	user, err := s.Storage.Get(ctx, req.GetId())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to get user: %v", err)
	}

	return &pb.GetUserSecretResponse{
		User: &pb.User{
			Id:       user.Id,
			Name:     user.Name,
			Password: user.Password,
			Email:    user.Email,
		},
	}, nil

}

func (s UserServer) GetStream(req *pb.GetUsersRequest, stream pb.UserService_GetStreamServer) error {
	ctx, cancel := context.WithTimeout(stream.Context(), time.Second*10)
	defer cancel()

	Users, err := s.Storage.GetMultiple(ctx, req.GetOffset(), req.GetLimit())
	if err != nil {
		return err
	}

	for _, v := range Users {
		select {
		case <-ctx.Done():
			return nil
		default:
			User := pb.User{
				Id:       v.Id,
				Name:     v.Name,
				Password: v.Password,
				Email:    v.Email,
			}

			err := stream.Send(&User)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// func (s UserServer) Interceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
// 	if info.FullMethod != "/proto.UserService/Get" {
// 		return handler(ctx, req)
// 	}
// 	// metadata.FromIncomingContext(ctx)
// 	return nil, nil
// }
