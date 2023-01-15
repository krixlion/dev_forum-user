package main

import (
	"os"
	"time"

	rabbitmq "github.com/krixlion/dev-forum_rabbitmq"
	"go.opentelemetry.io/otel"
)

func getServiceDependencies() Dependencies {
	logger, err := logging.NewLogger()
	if err != nil {
		panic(err)
	}

	cmd_port := os.Getenv("DB_WRITE_PORT")
	cmd_host := os.Getenv("DB_WRITE_HOST")
	cmd_user := os.Getenv("DB_WRITE_USER")
	cmd_pass := os.Getenv("DB_WRITE_PASS")
	// cmd, err := eventstore.MakeDB(cmd_port, cmd_host, cmd_user, cmd_pass, logger)
	if err != nil {
		panic(err)
	}

	query_port := os.Getenv("DB_READ_PORT")
	query_host := os.Getenv("DB_READ_HOST")
	query_pass := os.Getenv("DB_READ_PASS")
	// query, err := query.MakeDB(query_host, query_port, query_pass, logger)
	if err != nil {
		panic(err)
	}
	storage := storage.NewStorage(cmd, query, logger)

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
}
