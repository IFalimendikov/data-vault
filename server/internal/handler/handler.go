package handler

import (
	"context"
	"log/slog"

	"time"

	"data-vault/server/internal/config"
	"data-vault/server/internal/models"
	"data-vault/server/internal/proto"
	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const (
	tokenExpiryHours            = 24
	userIDKey        contextKey = "user_id"
)

// Service defines the interface for URL shortening operations
type Service interface {
	Register(ctx context.Context, user models.User) error
	Login(ctx context.Context, user models.User) error
	PostData(ctx context.Context, login, data string) error
	GetData(ctx context.Context, shortURL string) ([]models.Data, error)
	DeleteData(ctx context.Context, login, id string) error
}

// Handler manages GRPC request handling for URL shortening service
type Handler struct {
	proto.UnimplementedVaultServiceServer
	service Service
	cfg     config.Config
	log     *slog.Logger
}

type Claim struct {
	jwt.RegisteredClaims
	Login string
}

// New creates a new Handler instance
func New(ctx context.Context, s Service, cfg config.Config, log *slog.Logger) *Handler {
	return &Handler{
		service: s,
		cfg:     cfg,
		log:     log,
	}
}

func (g *Handler) IssueJWT(user models.User) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claim{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(tokenExpiryHours * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		},
		Login: user.Login,
	})

	signedToken, err := token.SignedString([]byte(g.cfg.JWTSecret))
	if err != nil {
		return "", err
	}

	return signedToken, nil
}
