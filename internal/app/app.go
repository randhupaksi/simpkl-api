package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"
)

type App struct {
	dependencies *Dependencies
	server       *http.Server
}

func New(dependencies *Dependencies) *App {
	return &App{
		dependencies: dependencies,
		server: &http.Server{
			Addr:              fmt.Sprintf(":%d", dependencies.Config.App.Port),
			Handler:           NewRouter(dependencies),
			ReadHeaderTimeout: 10 * time.Second,
			ReadTimeout:       30 * time.Second,
			WriteTimeout:      30 * time.Second,
			IdleTimeout:       60 * time.Second,
		},
	}
}

func (a *App) Run(ctx context.Context) error {
	errCh := make(chan error, 1)

	go func() {
		a.dependencies.Logger.Info(
			"api server started",
			zap.String("address", a.server.Addr),
			zap.String("environment", string(a.dependencies.Config.App.Env)),
		)
		errCh <- a.server.ListenAndServe()
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		a.dependencies.Logger.Info("api server shutting down")
		if err := a.server.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("shutdown server: %w", err)
		}

		err := <-errCh
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}

		return nil
	}
}
