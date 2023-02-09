package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/krixlion/dev_forum-lib/env"
	"github.com/krixlion/dev_forum-lib/event/broker"
	"github.com/krixlion/dev_forum-lib/logging"
	"github.com/krixlion/dev_forum-lib/tracing"
	rabbitmq "github.com/krixlion/dev_forum-rabbitmq"
	"github.com/krixlion/dev_forum-user/pkg/service"
	"github.com/krixlion/dev_forum-user/pkg/storage/db"
	"go.opentelemetry.io/otel"
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

	mq := rabbitmq.NewRabbitMQ(consumer, mq_user, mq_pass, mq_host, mq_port, mqConfig, logger, tracer)
	broker := broker.NewBroker(mq, logger, tracer)

	return service.Dependencies{
		Storage: storage,
		Logger:  logger,
		Broker:  broker,
	}
}
