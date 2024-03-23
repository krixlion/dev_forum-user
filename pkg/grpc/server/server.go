package server

import (
	"context"
	"time"

	"github.com/krixlion/dev_forum-lib/cert"
	"github.com/krixlion/dev_forum-lib/event"
	"github.com/krixlion/dev_forum-lib/event/dispatcher"
	"github.com/krixlion/dev_forum-lib/filter"
	"github.com/krixlion/dev_forum-lib/logging"
	pb "github.com/krixlion/dev_forum-user/pkg/grpc/v1"
	"github.com/krixlion/dev_forum-user/pkg/storage"

	fmask "github.com/mennanov/fieldmask-utils"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type UserServer struct {
	pb.UnimplementedUserServiceServer
	storage    storage.Storage
	dispatcher *dispatcher.Dispatcher
	broker     event.Broker
	logger     logging.Logger
	tracer     trace.Tracer
	config     Config
}

type Config struct {
	VerifyClientCert bool
}

type Dependencies struct {
	Storage    storage.Storage
	Broker     event.Broker
	Dispatcher *dispatcher.Dispatcher
	Logger     logging.Logger
	Tracer     trace.Tracer
	Config     Config
}

func MakeUserServer(d Dependencies) UserServer {
	return UserServer{
		storage:    d.Storage,
		broker:     d.Broker,
		dispatcher: d.Dispatcher,
		tracer:     d.Tracer,
		logger:     d.Logger,
		config:     d.Config,
	}
}

func (s UserServer) Close() error {
	return s.storage.Close()
}

func (s UserServer) Create(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	user := userFromPB(req.GetUser())

	if err := s.storage.Create(ctx, user); err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	event, err := event.MakeEvent(event.UserAggregate, event.UserCreated, user)
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	if err := s.broker.ResilientPublish(event); err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	return &pb.CreateUserResponse{
		Id: user.Id,
	}, nil
}

func (s UserServer) Delete(ctx context.Context, req *pb.DeleteUserRequest) (*emptypb.Empty, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	id := req.GetId()

	if err := s.storage.Delete(ctx, id); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	event, err := event.MakeEvent(event.UserAggregate, event.UserDeleted, id)
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	if err := s.broker.ResilientPublish(event); err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	return &emptypb.Empty{}, nil
}

func (s UserServer) Update(ctx context.Context, req *pb.UpdateUserRequest) (*emptypb.Empty, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	mask, err := fmask.MaskFromPaths(req.GetFieldMask().GetPaths(), mapUserFields)
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	if err := fmask.StructToStruct(mask, req.GetUser(), req.User); err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	user := userFromPB(req.GetUser())
	user.UpdatedAt = time.Now()

	if err := s.storage.Update(ctx, user); err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	event, err := event.MakeEvent(event.UserAggregate, event.UserUpdated, user)
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	if err := s.broker.ResilientPublish(event); err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	return &emptypb.Empty{}, nil
}

func (s UserServer) Get(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	query := filter.Filter{{
		Attribute: "id",
		Operator:  filter.Equal,
		Value:     req.GetId(),
	}}

	user, err := s.storage.Get(ctx, query)
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

	if s.config.VerifyClientCert {
		if err := cert.VerifyClientTLS(ctx, "auth-service"); err != nil {
			return nil, err
		}
	}

	query := filter.Filter{}

	switch req.GetQuery().(type) {
	case *pb.GetUserSecretRequest_Email:
		query = append(query, filter.Parameter{
			Attribute: "email",
			Operator:  filter.Equal,
			Value:     req.GetEmail(),
		})
	case *pb.GetUserSecretRequest_Id:
		query = append(query, filter.Parameter{
			Attribute: "id",
			Operator:  filter.Equal,
			Value:     req.GetId(),
		})
	}

	user, err := s.storage.Get(ctx, query)
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

	query, err := filter.Parse(req.GetFilter())
	if err != nil {
		return err
	}

	users, err := s.storage.GetMultiple(ctx, req.GetOffset(), req.GetLimit(), query)
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
