package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/krixlion/dev_forum-user/cmd/service"
	"github.com/krixlion/dev_forum-user/pkg/env"
	"github.com/krixlion/dev_forum-user/pkg/logging"
	"github.com/krixlion/dev_forum-user/pkg/tracing"
)

var port int

func init() {
	portFlag := flag.Int("p", 50051, "The gRPC server port")
	flag.Parse()
	port = *portFlag
}

// Hardcoded root dir name.
const projectDir = "app"

func main() {
	env.Load(projectDir)

	shutdownTracing, err := tracing.InitProvider()
	if err != nil {
		logging.Log("Failed to initialize tracing", "err", err)
	}

	service := service.NewUserService(port, getServiceDependencies())
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
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
