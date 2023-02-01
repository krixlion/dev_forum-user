package service

import (
	"context"

	"fmt"
	"net"

	"github.com/krixlion/dev_forum-proto/user_service/pb"
	"github.com/krixlion/dev_forum-user/pkg/event"
	"github.com/krixlion/dev_forum-user/pkg/event/dispatcher"
	"github.com/krixlion/dev_forum-user/pkg/grpc/server"
	"github.com/krixlion/dev_forum-user/pkg/logging"
	"github.com/krixlion/dev_forum-user/pkg/storage"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type UserService struct {
	grpcPort   int
	grpcServer *grpc.Server
	server     server.UserServer

	broker     event.Broker
	dispatcher *dispatcher.Dispatcher

	logger logging.Logger
}

type Dependencies struct {
	Storage storage.Storage
	Logger  logging.Logger
	Broker  event.Broker
}

func NewUserService(grpcPort int, d Dependencies) UserService {
	dispatcher := dispatcher.NewDispatcher(d.Broker, 20)

	s := UserService{
		grpcPort: grpcPort,
		server: server.UserServer{
			Dispatcher: dispatcher,
			Storage:    d.Storage,
			Logger:     d.Logger,
		},
		dispatcher: dispatcher,
		broker:     d.Broker,
		logger:     d.Logger,
	}

	baseSrv := grpc.NewServer(
		grpc.UnaryInterceptor(otelgrpc.UnaryServerInterceptor()),
		grpc.StreamInterceptor(otelgrpc.StreamServerInterceptor()),
	)
	s.grpcServer = baseSrv
	reflection.Register(s.grpcServer)
	pb.RegisterUserServiceServer(s.grpcServer, s.server)
	return s
}

func (s *UserService) Run(ctx context.Context) {
	if err := ctx.Err(); err != nil {
		return
	}

	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", s.grpcPort))
	if err != nil {
		s.logger.Log(ctx, "failed to create a listener", "transport", "grpc", "err", err)
	}

	go s.dispatcher.Run(ctx)

	s.logger.Log(ctx, "listening", "transport", "grpc", "port", s.grpcPort)
	err = s.grpcServer.Serve(lis)
	if err != nil {
		s.logger.Log(ctx, "failed to serve", "transport", "grpc", "err", err)
	}
}

func (s *UserService) Close() error {
	s.grpcServer.GracefulStop()
	return s.server.Close()
}
