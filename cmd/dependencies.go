package main

import (
	"os"
	"time"

	rabbitmq "github.com/krixlion/dev_forum-rabbitmq"
	"github.com/krixlion/dev_forum-user/cmd/service"
	"github.com/krixlion/dev_forum-user/pkg/event/broker"
	"github.com/krixlion/dev_forum-user/pkg/logging"
	"github.com/krixlion/dev_forum-user/pkg/storage/db"
	"github.com/krixlion/dev_forum-user/pkg/tracing"
	"go.opentelemetry.io/otel"
)

func getServiceDependencies() service.Dependencies {
	logger, err := logging.NewLogger()
	if err != nil {
		panic(err)
	}

	db_port := os.Getenv("DB_PORT")
	db_host := os.Getenv("DB_HOST")
	db_user := os.Getenv("DB_USER")
	db_pass := os.Getenv("DB_PASS")
	db_name := os.Getenv("DB_NAME")
	storage, err := db.Make(db_host, db_port, db_user, db_pass, db_name)
	if err != nil {
		panic(err)
	}

	mq_port := os.Getenv("MQ_PORT")
	mq_host := os.Getenv("MQ_HOST")
	mq_user := os.Getenv("MQ_USER")
	mq_pass := os.Getenv("MQ_PASS")

	consumer := tracing.ServiceName
	mqConfig := rabbitmq.Config{
		QueueSize:         100,
		MaxWorkers:        100,
		ReconnectInterval: time.Second * 2,
		MaxRequests:       30,
		ClearInterval:     time.Second * 5,
		ClosedTimeout:     time.Second * 15,
	}

	tracer := otel.Tracer(tracing.ServiceName)

	mq := rabbitmq.NewRabbitMQ(consumer, mq_user, mq_pass, mq_host, mq_port, mqConfig, logger, tracer)
	broker := broker.NewBroker(mq, logger)
	return service.Dependencies{
		Storage: storage,
		Logger:  logger,
		Broker:  broker,
	}
}
