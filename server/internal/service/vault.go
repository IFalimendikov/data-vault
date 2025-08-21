package service

import (
	"context"
	"data-vault/server/internal/config"
	"data-vault/server/internal/models"
	"data-vault/server/internal/storage"
	"log/slog"
)

type Service interface {
	Register(ctx context.Context, user models.User) error
	Login(ctx context.Context, user models.User) error
	PostData(ctx context.Context, login, data string) error
	GetData(ctx context.Context, login string) ([]models.Data, error)
	DeleteData(ctx context.Context, login, id string) error
}

type Vault struct {
	Log     *slog.Logger
	cfg     *config.Config
	Storage *storage.Storage
}

func New(log *slog.Logger, cfg config.Config, storage *storage.Storage) *Vault {
	service := Vault{
		Log:     log,
		cfg:     &cfg,
		Storage: storage,
	}
	return &service
}
