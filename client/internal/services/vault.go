package services

import (
	"context"
	"data-vault/client/internal/grpcclient"
	"data-vault/client/internal/models"
	"log/slog"
)

// Service defines the interface for URL shortening operations
type Service interface {
	Register(ctx context.Context, user models.User) (string, error)
	Login(ctx context.Context, user models.User) (string, error)
	PostData(ctx context.Context, jwt, data string) error
	GetData(ctx context.Context, jwt string) ([]models.Data, error)
	DeleteData(ctx context.Context, jwt, id string) error
	PingServer(ctx context.Context) bool
}

// URLs implements the Service interface and manages URL shortening operations
type Vault struct {
	ctx        context.Context
	Log        *slog.Logger       // Logger for service operations
	grpcclient *grpcclient.Client // grpcclient interface for persistence
}

// New creates and initializes a new URLs service instance
func New(ctx context.Context, log *slog.Logger, grpcclient *grpcclient.Client) *Vault {
	service := Vault{
		ctx:        ctx,
		grpcclient: grpcclient,
		Log:        log,
	}
	return &service
}
