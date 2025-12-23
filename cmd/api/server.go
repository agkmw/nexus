package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/agkmw/reddit-clone/internal/platform/logger"
)

func serve(
	ctx context.Context,
	cfg config,
	mux http.Handler,
	log *logger.Logger,
) error {
	server := http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.port),
		Handler:      mux,
		IdleTimeout:  2 * time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		ErrorLog:     logger.NewStdLogger(log, logger.LevelError),
	}

	shutdownErr := make(chan error)

	go func() {
		shutdown := make(chan os.Signal, 1)
		signal.Notify(shutdown, syscall.SIGTERM, syscall.SIGINT)
		sig := <-shutdown

		log.Info(ctx, "gracefully shutting down the server", "signal", sig.String())

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		err := server.Shutdown(ctx)
		if err != nil {
			shutdownErr <- err
		}

		log.Info(ctx, "completing background tasks", "addr", server.Addr)

		// app.wg.Wait()
		shutdownErr <- nil
	}()

	log.Info(ctx, "starting server", "addr", server.Addr, "env", cfg.environment)

	if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	if err := <-shutdownErr; err != nil {
		return err
	}

	log.Info(ctx, "server gracefully shut down")

	return nil
}
