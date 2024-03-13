package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"github.com/krixlion/dev_forum-lib/cert"
	"github.com/krixlion/dev_forum-lib/env"
	"github.com/krixlion/dev_forum-lib/event/broker"
	"github.com/krixlion/dev_forum-lib/event/dispatcher"
	"github.com/krixlion/dev_forum-lib/logging"
	"github.com/krixlion/dev_forum-lib/tracing"
	rabbitmq "github.com/krixlion/dev_forum-rabbitmq"
	"github.com/krixlion/dev_forum-user/pkg/grpc/server"
	pb "github.com/krixlion/dev_forum-user/pkg/grpc/v1"
	"github.com/krixlion/dev_forum-user/pkg/service"
	"github.com/krixlion/dev_forum-user/pkg/storage/cockroach"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
)

var port int
var isTLS bool

// Hardcoded root dir name.
const projectDir = "app"
const serviceName = "user-service"

func init() {
	portFlag := flag.Int("p", 50051, "The gRPC server port")
	insecureFlag := flag.Bool("insecure", false, "Whether to not use TLS over gRPC")
	flag.Parse()
	port = *portFlag
	isTLS = !(*insecureFlag)
}

func main() {
	env.Load(projectDir)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	deps, err := getServiceDependencies(ctx, serviceName, isTLS)
	if err != nil {
		logging.Log("Failed to initialize service dependencies", "err", err)
		return
	}

	service := service.NewUserService(port, deps)
	service.Run(ctx)

	<-ctx.Done()
	logging.Log("Service shutting down")

	defer func() {
		cancel()

		if err := service.Close(); err != nil {
			logging.Log("Failed to shutdown service", "err", err)
			return
		}

		logging.Log("Service shutdown successful")
	}()
}

// getServiceDependencies is a Composition root.
// Panics on any non-nil error.
func getServiceDependencies(ctx context.Context, serviceName string, isTLS bool) (service.Dependencies, error) {
	serverCreds := insecure.NewCredentials()
	if isTLS {
		caCertPool, err := cert.LoadCaPool(os.Getenv("TLS_CA_PATH"))
		if err != nil {
			return service.Dependencies{}, err
		}

		serverCert, err := cert.LoadX509KeyPair(os.Getenv("TLS_CERT_PATH"), os.Getenv("TLS_KEY_PATH"))
		if err != nil {
			return service.Dependencies{}, err
		}

		serverCreds = cert.NewServerOptionalMTLSCreds(caCertPool, serverCert)
	}

	shutdownTracing, err := tracing.InitProvider(ctx, serviceName)
	if err != nil {
		return service.Dependencies{}, err
	}

	tracer := otel.Tracer(serviceName)

	logger, err := logging.NewLogger()
	if err != nil {
		return service.Dependencies{}, err
	}

	storage, err := cockroach.Make(os.Getenv("DB_HOST"), os.Getenv("DB_PORT"), os.Getenv("DB_USER"), os.Getenv("DB_PASS"), os.Getenv("DB_NAME"), tracer)
	if err != nil {
		return service.Dependencies{}, err
	}

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
		os.Getenv("MQ_USER"),
		os.Getenv("MQ_PASS"),
		os.Getenv("MQ_HOST"),
		os.Getenv("MQ_PORT"),
		mqConfig,
		rabbitmq.WithLogger(logger),
		rabbitmq.WithTracer(tracer),
	)
	broker := broker.NewBroker(messageQueue, logger, tracer)
	dispatcher := dispatcher.NewDispatcher(20)

	userServer := server.MakeUserServer(server.Dependencies{
		Storage:    storage,
		Logger:     logger,
		Broker:     broker,
		Tracer:     tracer,
		Dispatcher: dispatcher,
	})

	grpcServer := grpc.NewServer(
		grpc.Creds(serverCreds),
		grpc.StreamInterceptor(otelgrpc.StreamServerInterceptor()),
		grpc.ChainUnaryInterceptor(
			grpc_recovery.UnaryServerInterceptor(),
			otelgrpc.UnaryServerInterceptor(),
			userServer.ValidateRequestInterceptor(),
		),
	)
	reflection.Register(grpcServer)
	pb.RegisterUserServiceServer(grpcServer, userServer)

	closeFunc := func() error {
		grpcServer.GracefulStop()
		shutdownTracing()
		return userServer.Close()
	}

	return service.Dependencies{
		Logger:       logger,
		Dispatcher:   dispatcher,
		GRPCServer:   grpcServer,
		Broker:       broker,
		ShutdownFunc: closeFunc,
	}, nil
}
