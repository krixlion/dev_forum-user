package server

import (
	"context"
	"html"
	"net/mail"
	"time"

	"github.com/gofrs/uuid"
	"github.com/krixlion/dev_forum-lib/tracing"
	pb "github.com/krixlion/dev_forum-user/pkg/grpc/v1"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s UserServer) ValidateRequestInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		switch info.FullMethod {
		case "/user.UserService/Create":
			return s.validateCreate(ctx, req.(*pb.CreateUserRequest), handler)
		case "/user.UserService/Update":
			return s.validateUpdate(ctx, req.(*pb.UpdateUserRequest), handler)
		case "/user.UserService/Delete":
			return s.validateDelete(ctx, req.(*pb.DeleteUserRequest), handler)
		default:
			return handler(ctx, req)
		}
	}
}

func (s UserServer) validateCreate(ctx context.Context, req *pb.CreateUserRequest, handler grpc.UnaryHandler) (interface{}, error) {
	ctx, span := s.tracer.Start(ctx, "server.validateCreate")
	defer span.End()

	user := req.GetUser()
	if user == nil {
		err := status.Error(codes.InvalidArgument, "User not provided")
		tracing.SetSpanErr(span, err)
		return nil, err
	}

	// Sanitize user input.
	// Assign a new ID: do not let users assign custom IDs.
	id, err := uuid.NewV4()
	if err != nil {
		err := status.Error(codes.Internal, err.Error())
		tracing.SetSpanErr(span, err)
		return nil, err
	}
	user.Id = id.String()
	user.Name = html.EscapeString(user.GetName())

	// Validate email.
	if _, err := mail.ParseAddress(user.Email); err != nil {
		tracing.SetSpanErr(span, err)
		return nil, err
	}

	// Password has to be at least 8 characters long.
	if len(user.GetPassword()) < 8 {
		err := status.Error(codes.FailedPrecondition, "Provided password is too short")
		tracing.SetSpanErr(span, err)
		return nil, err
	}

	// Hash password before saving.
	hash, err := bcrypt.GenerateFromPassword([]byte(user.GetPassword()), bcrypt.MinCost)
	if err != nil {
		err := status.Errorf(codes.Internal, err.Error())
		return nil, err
	}

	user.Password = string(hash)
	user.CreatedAt = timestamppb.New(time.Now())
	user.UpdatedAt = timestamppb.New(time.Time{})

	return handler(ctx, req)
}

func (s UserServer) validateUpdate(ctx context.Context, req *pb.UpdateUserRequest, handler grpc.UnaryHandler) (interface{}, error) {
	ctx, span := s.tracer.Start(ctx, "server.validateUpdate")
	defer span.End()

	user := req.GetUser()

	if user == nil {
		err := status.Error(codes.FailedPrecondition, "User not provided")
		tracing.SetSpanErr(span, err)
		return nil, err
	}

	// Sanitize user input.
	user.Id = ""
	user.Name = html.EscapeString(user.GetName())
	user.CreatedAt = timestamppb.New(time.Time{})
	user.UpdatedAt = timestamppb.New(time.Now())

	// Validate email.
	if _, err := mail.ParseAddress(user.GetEmail()); err != nil {
		tracing.SetSpanErr(span, err)
		return nil, err
	}

	// Password has to be at least 8 characters long.
	if len(user.Password) < 8 {
		err := status.Error(codes.FailedPrecondition, "Provided password is too short")
		tracing.SetSpanErr(span, err)
		return nil, err
	}

	return handler(ctx, req)
}

func (s UserServer) validateDelete(ctx context.Context, req *pb.DeleteUserRequest, handler grpc.UnaryHandler) (interface{}, error) {
	ctx, span := s.tracer.Start(ctx, "server.validateDelete")
	defer span.End()

	id := req.GetId()

	if id == "" {
		err := status.Error(codes.FailedPrecondition, "User id not provided")
		tracing.SetSpanErr(span, err)
		return nil, err
	}

	if _, err := s.storage.Get(ctx, id); err != nil {
		tracing.SetSpanErr(span, err)
		// Do not let user know whether entity with provided ID existed before deleting or not.
		return nil, nil
	}

	return handler(ctx, req)
}
