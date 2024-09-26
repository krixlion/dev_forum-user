package server

import (
	"context"
	"errors"
	"html"
	"net/mail"
	"slices"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/selector"
	"github.com/krixlion/dev_forum-lib/filter"
	"github.com/krixlion/dev_forum-lib/tracing"
	pb "github.com/krixlion/dev_forum-user/pkg/grpc/v1"
	"github.com/krixlion/dev_forum-user/pkg/storage"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// AuthMatcher returns a callback used to determine whether current gRPC path should be protected by auth middleware.
func (UserServer) AuthMatcher() selector.Matcher {
	return selector.MatchFunc(func(ctx context.Context, callMeta interceptors.CallMeta) bool {
		// List of paths excluded from auth middleware.
		disabledAuthPaths := []string{
			"/user.UserService/Create",
			"/user.UserService/Get",
			"/user.UserService/GetSecret",
		}
		return !slices.Contains(disabledAuthPaths, callMeta.FullMethod())
	})
}

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

func (s UserServer) validateCreate(ctx context.Context, req *pb.CreateUserRequest, handler grpc.UnaryHandler) (_ interface{}, err error) {
	ctx, span := s.tracer.Start(ctx, "server.validateCreate")
	defer span.End()
	defer tracing.SetSpanErr(span, err)

	user := req.GetUser()
	if user == nil {
		return nil, status.Error(codes.InvalidArgument, "User not provided")
	}

	// Sanitize user input.
	// Assign a new ID: do not let users assign custom IDs.
	id, err := uuid.NewV4()
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	user.Id = id.String()
	user.Name = html.EscapeString(user.GetName())

	// Validate email.
	if _, err := mail.ParseAddress(user.Email); err != nil {
		return nil, err
	}

	// Password has to be at least 8 characters long.
	if len(user.GetPassword()) < 8 {
		return nil, status.Error(codes.FailedPrecondition, "Provided password is too short")
	}

	// Hash password before saving.
	hash, err := bcrypt.GenerateFromPassword([]byte(user.GetPassword()), bcrypt.MinCost)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to generate hash from password: %v", err.Error())
	}

	user.Password = string(hash)
	user.CreatedAt = timestamppb.New(time.Now())
	user.UpdatedAt = timestamppb.New(time.Time{})

	return handler(ctx, req)
}

func (s UserServer) validateUpdate(ctx context.Context, req *pb.UpdateUserRequest, handler grpc.UnaryHandler) (_ interface{}, err error) {
	ctx, span := s.tracer.Start(ctx, "server.validateUpdate")
	defer span.End()
	defer tracing.SetSpanErr(span, err)

	user := req.GetUser()

	if user == nil {
		return nil, status.Error(codes.FailedPrecondition, "User not provided")
	}

	// Sanitize user input.
	user.Id = ""
	user.Name = html.EscapeString(user.GetName())
	user.CreatedAt = timestamppb.New(time.Time{})
	user.UpdatedAt = timestamppb.New(time.Now())

	// Validate email.
	if _, err := mail.ParseAddress(user.GetEmail()); err != nil {
		return nil, err
	}

	// Password has to be at least 8 characters long.
	if len(user.Password) < 8 {
		return nil, status.Error(codes.FailedPrecondition, "Provided password is too short")
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

	query := filter.Filter{{
		Attribute: "id",
		Operator:  filter.Equal,
		Value:     id,
	}}

	if _, err := s.storage.Get(ctx, query); err != nil {
		tracing.SetSpanErr(span, err)
		if errors.Is(err, storage.ErrNotFound) {
			// Do not let the user know whether user with provided ID existed or not.
			return nil, nil
		}
		return nil, status.Error(codes.Internal, "Failed to delete user")
	}

	return handler(ctx, req)
}
