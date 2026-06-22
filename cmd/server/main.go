package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	slog.Info("data_sync_event starting")

	// TODO(phase-07): wire config -> store -> publisher -> manager -> metrics -> http server.

	<-ctx.Done()
	slog.Info("shutdown signal received, exiting")
}
