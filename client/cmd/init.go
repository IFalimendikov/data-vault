package main

import (
	"context"

	"data-vault/client/internal/config"
	"data-vault/client/internal/grpcclient"
	"data-vault/client/internal/logger"
	"data-vault/client/internal/services"
)

// initService initializes the service layer with gRPC client
func initService() (*services.Vault, error) {
	cfg, err := config.New()
	if err != nil {
		return nil, err
	}

	log := logger.New()
	ctx := context.Background()

	client, err := grpcclient.New(ctx, cfg)
	if err != nil {
		return nil, err
	}

	service := services.New(ctx, log, client)
	return service, nil
}
