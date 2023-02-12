package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"github.com/krixlion/dev_forum-lib/env"
	"github.com/krixlion/dev_forum-lib/event/broker"
	"github.com/krixlion/dev_forum-lib/event/dispatcher"
	"github.com/krixlion/dev_forum-lib/logging"
	"github.com/krixlion/dev_forum-lib/tracing"
	"github.com/krixlion/dev_forum-proto/user_service/pb"
	rabbitmq "github.com/krixlion/dev_forum-rabbitmq"
	"github.com/krixlion/dev_forum-user/pkg/grpc/server"
	"github.com/krixlion/dev_forum-user/pkg/service"
	"github.com/krixlion/dev_forum-user/pkg/storage/db"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var port int

// Hardcoded root dir name.
const projectDir = "app"
const serviceName = "user-service"

func init() {
	portFlag := flag.Int("p", 50051, "The gRPC server port")
	flag.Parse()
	port = *portFlag
}

func main() {
	env.Load(projectDir)
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	shutdownTracing, err := tracing.InitProvider(ctx, serviceName)
	if err != nil {
		logging.Log("Failed to initialize tracing", "err", err)
	}

	service := service.NewUserService(port, getServiceDependencies())
	service.Run(ctx)

	<-ctx.Done()
	logging.Log("Service shutting down")

	defer func() {
		cancel()
		shutdownTracing()
		err := service.Close()
		if err != nil {
			logging.Log("Failed to shutdown service", "err", err)
		} else {
			logging.Log("Service shutdown properly")
		}
	}()
}

// getServiceDependencies is a Composition root.
// Panics on any non-nil error.
func getServiceDependencies() service.Dependencies {
	tracer := otel.Tracer(serviceName)
	logger, err := logging.NewLogger()
	if err != nil {
		panic(err)
	}

	db_port := os.Getenv("DB_PORT")
	db_host := os.Getenv("DB_HOST")
	db_user := os.Getenv("DB_USER")
	db_pass := os.Getenv("DB_PASS")
	db_name := os.Getenv("DB_NAME")
	storage, err := db.Make(db_host, db_port, db_user, db_pass, db_name, tracer)
	if err != nil {
		panic(err)
	}

	mq_port := os.Getenv("MQ_PORT")
	mq_host := os.Getenv("MQ_HOST")
	mq_user := os.Getenv("MQ_USER")
	mq_pass := os.Getenv("MQ_PASS")

	consumer := serviceName
	mqConfig := rabbitmq.Config{
		QueueSize:         100,
		MaxWorkers:        100,
		ReconnectInterval: time.Second * 2,
		MaxRequests:       30,
		ClearInterval:     time.Second * 5,
		ClosedTimeout:     time.Second * 15,
	}

	mq := rabbitmq.NewRabbitMQ(consumer, mq_user, mq_pass, mq_host, mq_port, mqConfig, rabbitmq.WithLogger(logger), rabbitmq.WithTracer(tracer))
	broker := broker.NewBroker(mq, logger, tracer)
	dispatcher := dispatcher.NewDispatcher(broker, 20)

	server := server.NewUserServer(storage, logger, dispatcher)
	grpcServer := grpc.NewServer(
		grpc.StreamInterceptor(otelgrpc.StreamServerInterceptor()),

		grpc_middleware.WithUnaryServerChain(
			grpc_recovery.UnaryServerInterceptor(),
			grpc_zap.UnaryServerInterceptor(zap.L()),
			otelgrpc.UnaryServerInterceptor(),
			server.ValidateRequestInterceptor(),
		),
	)
	reflection.Register(grpcServer)
	pb.RegisterUserServiceServer(grpcServer, server)

	closeFunc := func() error {
		grpcServer.GracefulStop()
		return server.Close()
	}

	return service.Dependencies{
		Logger:       logger,
		Dispatcher:   dispatcher,
		GRPCServer:   grpcServer,
		Broker:       broker,
		ShutdownFunc: closeFunc,
	}
}
