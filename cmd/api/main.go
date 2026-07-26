package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"

	"simpkl-api/internal/app"
)

func main() {
	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	dependencies, err := app.BuildDependencies(ctx)
	if err != nil {
		log.Printf("failed to initialize application: %v", err)
		os.Exit(1)
	}
	defer dependencies.Close()

	application := app.New(dependencies)
	if err := application.Run(ctx); err != nil &&
		!errors.Is(err, http.ErrServerClosed) {
		dependencies.Logger.Error("application stopped unexpectedly", zap.Error(err))
		os.Exit(1)
	}
}
