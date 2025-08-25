package main

import (
	"context"
	_ "net/http/pprof"
	"os/signal"
	"syscall"

	"data-vault/server/internal/config"
	"data-vault/server/internal/handler"
	"data-vault/server/internal/logger"
	"data-vault/server/internal/service"
	"data-vault/server/internal/storage"
	"data-vault/server/internal/transport"

	"net"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// main is the entry point for the Data Vault server application
func main() {
	log := logger.New()

	cfg, err := config.New()
	if err != nil {
		log.Error("Error loading configuration: %v\n", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	defer stop()

	store, err := storage.New(ctx, &cfg)
	if err != nil {
		log.Error("Error creating new storage: %v\n", err)
	}
	defer store.DB.Close()

	s := service.New(log, cfg, store)

	h := handler.New(ctx, s, cfg, log)
	grpcErrCh := make(chan error, 1)

	server := transport.New(h, cfg, log)
	g, err := transport.NewRouter(server)
	if err != nil {
		log.Error("Error creating gRPC server: %v\n", err)
	}

	go func() {
		grpcAddr := cfg.ServerAddr
		lis, err := net.Listen("tcp", grpcAddr)
		if err != nil {
			grpcErrCh <- err
			return
		}
		log.Info("Starting gRPC server", "address", grpcAddr)
		if err := g.Serve(lis); err != nil {
			grpcErrCh <- err
		}
	}()

	select {
	case err := <-grpcErrCh:
		log.Error("gRPC server error", "error", err)
	case <-ctx.Done():
		log.Info("Servers shut down successfully")
	}
}
