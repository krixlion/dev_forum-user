package service

import (
	"context"

	"fmt"
	"net"

	"github.com/krixlion/dev_forum-proto/user_service/pb"
	"github.com/krixlion/dev_forum-user/pkg/event"
	"github.com/krixlion/dev_forum-user/pkg/logging"
	"github.com/krixlion/dev_forum-user/pkg/net/grpc/server"
	"github.com/krixlion/dev_forum-user/pkg/storage"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type UserService struct {
	grpcPort int
	grpcSrv  *grpc.Server
	srv      server.UserServer

	broker     event.Broker
	dispatcher *event.Dispatcher

	logger logging.Logger
}

type Dependencies struct {
	Storage storage.Storage
	Logger  logging.Logger
	Broker  event.Broker
}

func NewUserService(grpcPort int, d Dependencies) UserService {
	dispatcher := event.MakeDispatcher(20)

	srv := server.UserServer{
		Storage: d.Storage,
		Logger:  d.Logger,
	}

	baseSrv := grpc.NewServer(
		// grpc.UnaryInterceptor(srv.Interceptor),
		grpc.UnaryInterceptor(otelgrpc.UnaryServerInterceptor()),
		grpc.StreamInterceptor(otelgrpc.StreamServerInterceptor()),
	)

	s := UserService{
		grpcPort: grpcPort,
		grpcSrv:  baseSrv,

		srv: srv,

		dispatcher: &dispatcher,
		broker:     d.Broker,

		logger: d.Logger,
	}
	reflection.Register(s.grpcSrv)
	pb.RegisterUserServiceServer(s.grpcSrv, s.srv)
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

	go func() {
		// s.dispatcher.AddEventSources(s.SyncEventSources(ctx)...)
		s.dispatcher.Run(ctx)
	}()

	s.logger.Log(ctx, "listening", "transport", "grpc", "port", s.grpcPort)
	err = s.grpcSrv.Serve(lis)
	if err != nil {
		s.logger.Log(ctx, "failed to serve", "transport", "grpc", "err", err)
	}
}

func (s *UserService) Close() error {
	s.grpcSrv.GracefulStop()
	return s.srv.Close()
}

// func (s *UserService) SyncEventSources(ctx context.Context) (chans []<-chan event.Event) {

// 	aCreated, err := s.syncEventSource.Consume(ctx, "", event.ArticleCreated)
// 	if err != nil {
// 		panic(err)
// 	}

// 	aDeleted, err := s.syncEventSource.Consume(ctx, "", event.ArticleDeleted)
// 	if err != nil {
// 		panic(err)
// 	}

// 	aUpdated, err := s.syncEventSource.Consume(ctx, "", event.ArticleUpdated)
// 	if err != nil {
// 		panic(err)
// 	}

// 	return append(chans, aCreated, aDeleted, aUpdated)
// }
