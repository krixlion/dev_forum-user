package server

import (
	"context"
	"time"

	"github.com/krixlion/dev_forum-lib/event"
	"github.com/krixlion/dev_forum-lib/event/dispatcher"
	"github.com/krixlion/dev_forum-lib/logging"
	"github.com/krixlion/dev_forum-proto/user_service/pb"
	"github.com/krixlion/dev_forum-user/pkg/storage"
	"golang.org/x/crypto/bcrypt"

	"github.com/gofrs/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type UserServer struct {
	pb.UnimplementedUserServiceServer
	storage    storage.Storage
	logger     logging.Logger
	dispatcher *dispatcher.Dispatcher
}

func NewUserServer(storage storage.Storage, logger logging.Logger, dispatcher *dispatcher.Dispatcher) UserServer {
	return UserServer{
		storage:    storage,
		logger:     logger,
		dispatcher: dispatcher,
	}
}

func (s UserServer) Close() error {
	return s.storage.Close()
}

func (s UserServer) Create(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	user := userFromPB(req.GetUser())
	id, err := uuid.NewV4()
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.MinCost)
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	// Assign new UUID to new user.
	user.Id = id.String()
	user.Password = string(hash)
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Time{}

	if err := s.storage.Create(ctx, user); err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	s.dispatcher.Publish(event.MakeEvent(event.UserAggregate, event.UserCreated, user))

	return &pb.CreateUserResponse{
		Id: id.String(),
	}, nil
}

func (s UserServer) Delete(ctx context.Context, req *pb.DeleteUserRequest) (*pb.DeleteUserResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()
	id := req.GetId()

	if err := s.storage.Delete(ctx, id); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	s.dispatcher.Publish(event.MakeEvent(event.UserAggregate, event.UserDeleted, id))

	return &pb.DeleteUserResponse{}, nil
}

func (s UserServer) Update(ctx context.Context, req *pb.UpdateUserRequest) (*pb.UpdateUserResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	user := userFromPB(req.GetUser())
	user.UpdatedAt = time.Now()

	if err := s.storage.Update(ctx, user); err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	s.dispatcher.Publish(event.MakeEvent(event.UserAggregate, event.UserUpdated, user))

	return &pb.UpdateUserResponse{}, nil
}

func (s UserServer) Get(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	user, err := s.storage.Get(ctx, req.GetId())
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

	user, err := s.storage.Get(ctx, req.GetId())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to get user: %v", err)
	}

	return &pb.GetUserSecretResponse{
		User: &pb.User{
			Id:        user.Id,
			Name:      user.Name,
			Password:  user.Password,
			Email:     user.Email,
			CreatedAt: timestamppb.New(user.CreatedAt),
			UpdatedAt: timestamppb.New(user.UpdatedAt),
		},
	}, nil
}

func (s UserServer) GetStream(req *pb.GetUsersRequest, stream pb.UserService_GetStreamServer) error {
	ctx, cancel := context.WithTimeout(stream.Context(), time.Second*10)
	defer cancel()

	users, err := s.storage.GetMultiple(ctx, req.GetOffset(), req.GetLimit())
	if err != nil {
		return err
	}

	for _, v := range users {
		select {
		case <-ctx.Done():
			return nil
		default:
			user := pb.User{
				Id:   v.Id,
				Name: v.Name,
			}

			if err := stream.Send(&user); err != nil {
				return err
			}
		}
	}
	return nil
}
