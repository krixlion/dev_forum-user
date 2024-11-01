package main

import (
	"context"
	"errors"
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/auth"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/selector"
	"github.com/krixlion/dev_forum-auth/pkg/grpc/auth"
	authPb "github.com/krixlion/dev_forum-auth/pkg/grpc/v1"
	"github.com/krixlion/dev_forum-auth/pkg/tokens"
	"github.com/krixlion/dev_forum-auth/pkg/tokens/validator"
	"github.com/krixlion/dev_forum-lib/cert"
	"github.com/krixlion/dev_forum-lib/env"
	"github.com/krixlion/dev_forum-lib/event/broker"
	"github.com/krixlion/dev_forum-lib/event/dispatcher"
	"github.com/krixlion/dev_forum-lib/logging"
	"github.com/krixlion/dev_forum-lib/rabbitmq"
	"github.com/krixlion/dev_forum-lib/tracing"
	"github.com/krixlion/dev_forum-user/pkg/grpc/server"
	pb "github.com/krixlion/dev_forum-user/pkg/grpc/v1"
	"github.com/krixlion/dev_forum-user/pkg/service"
	"github.com/krixlion/dev_forum-user/pkg/storage/cockroach"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/grpclog"
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
	if err := env.Load(projectDir); err != nil {
		logging.Log("Failed to read env file", "err", err)
		return
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	deps, err := getServiceDependencies(ctx, serviceName, isTLS)
	if err != nil {
		logging.Log("Failed to initialize service dependencies", "err", err)
		return
	}

	service := service.MakeUserService(port, deps)
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
	clientCreds := insecure.NewCredentials()
	if isTLS {
		caCertPool, err := cert.LoadCaPool(os.Getenv("TLS_CA_PATH"))
		if err != nil {
			return service.Dependencies{}, err
		}

		clientCreds = credentials.NewClientTLSFromCert(caCertPool, "")

		serverCert, err := cert.LoadX509KeyPair(os.Getenv("TLS_CERT_PATH"), os.Getenv("TLS_KEY_PATH"))
		if err != nil {
			return service.Dependencies{}, err
		}

		serverCreds = cert.NewServerOptionalMTLSCreds(caCertPool, serverCert)
	}

	shutdownTracing, err := tracing.InitProvider(ctx, serviceName, os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"))
	if err != nil {
		return service.Dependencies{}, err
	}

	tracer := otel.Tracer(serviceName)

	logger, err := logging.NewLogger()
	if err != nil {
		return service.Dependencies{}, err
	}
	grpclog.SetLoggerV2(logger)

	storage, err := cockroach.Make(os.Getenv("DB_HOST"), os.Getenv("DB_PORT"), os.Getenv("DB_USER"), os.Getenv("DB_PASS"), os.Getenv("DB_NAME"), tracer)
	if err != nil {
		return service.Dependencies{}, err
	}

	authConn, err := grpc.NewClient(os.Getenv("AUTH_SERVICE_URL"),
		grpc.WithTransportCredentials(clientCreds),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
	)
	if err != nil {
		return service.Dependencies{}, err
	}
	tokenValidator, err := validator.NewValidator(tokens.DefaultIssuer, validator.DefaultRefreshFunc(authPb.NewAuthServiceClient(authConn), tracer), validator.WithLogger(logger))
	if err != nil {
		return service.Dependencies{}, err
	}

	go tokenValidator.Run(ctx)

	mqConfig := rabbitmq.Config{
		QueueSize:         100,
		MaxWorkers:        100,
		ReconnectInterval: time.Second * 2,
		MaxRequests:       30,
		ClearInterval:     time.Second * 5,
		ClosedTimeout:     time.Second * 15,
	}

	mq := rabbitmq.NewRabbitMQ(
		serviceName,
		os.Getenv("MQ_USER"),
		os.Getenv("MQ_PASS"),
		os.Getenv("MQ_HOST"),
		os.Getenv("MQ_PORT"),
		mqConfig,
		rabbitmq.WithLogger(logger),
		rabbitmq.WithTracer(tracer),
	)
	broker := broker.NewBroker(mq, logger, tracer)
	dispatcher := dispatcher.NewDispatcher(20)

	dispatcher.Register(tokenValidator)

	userConfig := server.Config{
		VerifyClientCert: isTLS,
	}

	userServer := server.MakeUserServer(server.Dependencies{
		Storage: storage,
		Logger:  logger,
		Broker:  broker,
		Tracer:  tracer,
		Config:  userConfig,
	})

	grpcServer := grpc.NewServer(
		grpc.Creds(serverCreds),
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
		grpc.ChainUnaryInterceptor(
			grpc_recovery.UnaryServerInterceptor(),
			selector.UnaryServerInterceptor(grpc_auth.UnaryServerInterceptor(auth.NewAuthFunc(tokenValidator, tracer)), userServer.AuthMatcher()),
			userServer.ValidateRequestInterceptor(),
		),
	)

	reflection.Register(grpcServer)
	pb.RegisterUserServiceServer(grpcServer, userServer)

	return service.Dependencies{
		Logger:     logger,
		Dispatcher: dispatcher,
		GRPCServer: grpcServer,
		Broker:     broker,
		ShutdownFunc: func() error {
			grpcServer.GracefulStop()
			return errors.Join(authConn.Close(), storage.Close(), mq.Close(), shutdownTracing(), logger.Sync())
		},
	}, nil
}
