package service

import (
	"context"

	"fmt"
	"net"

	"github.com/krixlion/dev_forum-lib/event"
	"github.com/krixlion/dev_forum-lib/event/dispatcher"
	"github.com/krixlion/dev_forum-lib/logging"
	"google.golang.org/grpc"
)

type UserService struct {
	grpcPort   int
	grpcServer *grpc.Server
	broker     event.Broker
	dispatcher *dispatcher.Dispatcher
	logger     logging.Logger
	shutdown   func() error
}

type Dependencies struct {
	Logger       logging.Logger
	Broker       event.Broker
	Dispatcher   *dispatcher.Dispatcher
	GRPCServer   *grpc.Server
	ShutdownFunc func() error
}

func NewUserService(grpcPort int, d Dependencies) UserService {
	return UserService{
		grpcPort:   grpcPort,
		grpcServer: d.GRPCServer,
		dispatcher: d.Dispatcher,
		broker:     d.Broker,
		logger:     d.Logger,
		shutdown:   d.ShutdownFunc,
	}
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

	if err := s.grpcServer.Serve(lis); err != nil {
		s.logger.Log(ctx, "failed to serve", "transport", "grpc", "err", err)
	}
}

func (s *UserService) Close() error {
	return s.shutdown()
}
