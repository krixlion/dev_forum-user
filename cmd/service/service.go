package service

import (
	"context"

	"fmt"
	"net"

	"github.com/Krixlion/def-forum_proto/article_service/pb"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type ArticleService struct {
	grpcPort int
	grpcSrv  *grpc.Server
	srv      server.EntityServer

	// Consumer for events used to update and sync the read model.
	syncEventSource event.Consumer
	broker          event.Broker
	dispatcher      event.Dispatcher

	logger logging.Logger
}

type Dependencies struct {
}

func NewArticleService(grpcPort int, d Dependencies) *ArticleService {
	dispatcher := event.MakeDispatcher()
	dispatcher.Subscribe(event.HandlerFunc(storage.CatchUp), event.ArticleCreated, event.ArticleDeleted, event.ArticleUpdated)

	srv := server.EntityServer{
		Storage: storage,
		Logger:  logger,
	}

	baseSrv := grpc.NewServer(
		// grpc.UnaryInterceptor(srv.Interceptor),
		grpc.UnaryInterceptor(otelgrpc.UnaryServerInterceptor()),
		grpc.StreamInterceptor(otelgrpc.StreamServerInterceptor()),
	)

	s := EntityService{
		grpcPort: grpcPort,
		grpcSrv:  baseSrv,

		srv: srv,

		dispatcher:      dispatcher,
		broker:          broker,
		syncEventSource: &cmd,

		logger: logger,
	}
	reflection.Register(s.grpcSrv)
	pb.RegisterArticleServiceServer(s.grpcSrv, s.srv)
	return s
}

func (s *ArticleService) Run(ctx context.Context) {
	if err := ctx.Err(); err != nil {
		return
	}

	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", s.grpcPort))
	if err != nil {
		s.logger.Log(ctx, "failed to create a listener", "transport", "grpc", "err", err)
	}

	go func() {
		s.dispatcher.AddEventSources(s.SyncEventSources(ctx)...)
		s.dispatcher.Run(ctx)
	}()

	s.logger.Log(ctx, "listening", "transport", "grpc", "port", s.grpcPort)
	err = s.grpcSrv.Serve(lis)
	if err != nil {
		s.logger.Log(ctx, "failed to serve", "transport", "grpc", "err", err)
	}
}

func (s *ArticleService) Close() error {
	s.grpcSrv.GracefulStop()
	return s.srv.Close()
}

func (s *ArticleService) SyncEventSources(ctx context.Context) (chans []<-chan event.Event) {

	aCreated, err := s.syncEventSource.Consume(ctx, "", event.ArticleCreated)
	if err != nil {
		panic(err)
	}

	aDeleted, err := s.syncEventSource.Consume(ctx, "", event.ArticleDeleted)
	if err != nil {
		panic(err)
	}

	aUpdated, err := s.syncEventSource.Consume(ctx, "", event.ArticleUpdated)
	if err != nil {
		panic(err)
	}

	return append(chans, aCreated, aDeleted, aUpdated)
}
