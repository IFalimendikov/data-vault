package handler

import (
	"context"
	"data-vault/client/internal/config"
	"data-vault/client/internal/models"
	"log/slog"
	"sync"
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

type UserAuth struct {
	mu sync.Mutex
	JWT string
}

// Handler manages HTTP request handling for URL shortening service
type Handler struct {
	service Service
	client  UserAuth
	cfg     config.Config
	log     *slog.Logger
}

// New creates a new Handler instance
func New(s Service, cfg config.Config, log *slog.Logger) *Handler {
	return &Handler{
		service: s,
		client:  UserAuth{},
		cfg:     cfg,
		log:     log,
	}
}

func (c *UserAuth) SetJWT(jwt string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.JWT = jwt
}

func (c *UserAuth) GetJWT() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.JWT
}