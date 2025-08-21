package grpcclient

import (
	"context"
	"data-vault/client/internal/config"

	"data-vault/client/internal/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// grpcclient holds the file, database and URL mapping information
type Client struct {
	cfg    config.Config
	ClientConn proto.VaultServiceClient
}

// New creates a new grpcclient instance with file and database connections
func New(ctx context.Context, cfg config.Config) (*Client, error) {
	creds, err := credentials.NewClientTLSFromFile("server.crt", "")
	if err != nil {
		return nil, err
	}

	conn, err := grpc.NewClient(
		cfg.ServerAddr,
		grpc.WithTransportCredentials(creds),
	)
	if err != nil {
		return nil, err
	}

	grpcClient := proto.NewVaultServiceClient(conn)

	clientInstance := Client{
		cfg:    cfg,
		ClientConn: grpcClient,
	}

	return &clientInstance, nil
}
