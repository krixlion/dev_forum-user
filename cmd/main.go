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

	dbPort := os.Getenv("DB_PORT")
	dbHost := os.Getenv("DB_HOST")
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASS")
	dbName := os.Getenv("DB_NAME")
	storage, err := db.Make(dbHost, dbPort, dbUser, dbPass, dbName, tracer)
	if err != nil {
		panic(err)
	}

	mqPort := os.Getenv("MQ_PORT")
	mqHost := os.Getenv("MQ_HOST")
	mqUser := os.Getenv("MQ_USER")
	mqPass := os.Getenv("MQ_PASS")

	consumer := serviceName
	mqConfig := rabbitmq.Config{
		QueueSize:         100,
		MaxWorkers:        100,
		ReconnectInterval: time.Second * 2,
		MaxRequests:       30,
		ClearInterval:     time.Second * 5,
		ClosedTimeout:     time.Second * 15,
	}

	messageQueue := rabbitmq.NewRabbitMQ(
		consumer,
		mqUser,
		mqPass,
		mqHost,
		mqPort,
		mqConfig,
		rabbitmq.WithLogger(logger),
		rabbitmq.WithTracer(tracer),
	)
	broker := broker.NewBroker(messageQueue, logger, tracer)
	dispatcher := dispatcher.NewDispatcher(broker, 20)

	userServer := server.NewUserServer(storage, logger, dispatcher)
	grpcServer := grpc.NewServer(
		grpc.StreamInterceptor(otelgrpc.StreamServerInterceptor()),

		grpc_middleware.WithUnaryServerChain(
			grpc_recovery.UnaryServerInterceptor(),
			grpc_zap.UnaryServerInterceptor(zap.L()),
			otelgrpc.UnaryServerInterceptor(),
			userServer.ValidateRequestInterceptor(),
		),
	)
	reflection.Register(grpcServer)
	pb.RegisterUserServiceServer(grpcServer, userServer)

	closeFunc := func() error {
		grpcServer.GracefulStop()
		return userServer.Close()
	}

	return service.Dependencies{
		Logger:       logger,
		Dispatcher:   dispatcher,
		GRPCServer:   grpcServer,
		Broker:       broker,
		ShutdownFunc: closeFunc,
	}
}
