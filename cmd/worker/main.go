package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/novriyantoAli/wallet-ms-backend/internal/config"
	"github.com/novriyantoAli/wallet-ms-backend/internal/pkg/database"
	"github.com/novriyantoAli/wallet-ms-backend/internal/pkg/logger"
	"github.com/novriyantoAli/wallet-ms-backend/internal/pkg/queue"
	"github.com/novriyantoAli/wallet-ms-backend/internal/server/worker"

	"go.uber.org/fx"
)

func main() {
	app := fx.New(
		fx.Provide(
			config.NewConfig,
			logger.NewLogger,
			database.NewDatabase,
			queue.NewClient,
			queue.NewServer,
		),
		worker.Module,
		fx.Invoke(runWorker),
		fx.StartTimeout(config.DefaultStartTimeout),
		fx.StopTimeout(config.DefaultStopTimeout),
	)

	ctx := context.Background()
	if err := app.Start(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to start worker application: %v\n", err)
		os.Exit(1)
	}

	// Setup graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Wait for shutdown signal
	<-sigChan
	fmt.Println("\nReceived shutdown signal, stopping worker gracefully...")

	if err := app.Stop(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to stop worker application gracefully: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Worker stopped successfully")
}

func runWorker(lifecycle fx.Lifecycle, workerServer *worker.Server, queueServer *queue.Server) {
	// Register worker handlers
	workerServer.RegisterHandlers()

	// Start the queue api (it manages its own lifecycle)
	queueServer.Start(lifecycle)
}
