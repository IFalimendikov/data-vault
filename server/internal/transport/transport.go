package transport

import (
	"context"
	"log/slog"

	"data-vault/server/internal/config"
	"data-vault/server/internal/handler"
	"data-vault/server/internal/proto"

	"google.golang.org/grpc"

	"time"

	"github.com/golang-jwt/jwt/v4"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// contextKey is a type for keys used in context to avoid collisions
type contextKey string

const (
	registerMethod            = "/proto.VaultService/Register"
	loginMethod               = "/proto.VaultService/Login"
	userIDKey      contextKey = "user_id"
)

// Transport handles gRPC transport layer operations including middleware and routing
type Transport struct {
	handler *handler.Handler
	cfg     config.Config
	log     *slog.Logger
}

// New creates a new Transport instance with the provided configuration and handlers
func New(h *handler.Handler, cfg config.Config, log *slog.Logger) *Transport {
	return &Transport{
		handler: h,
		cfg:     cfg,
		log:     log,
	}
}

// Claims represents JWT claims structure
type Claim struct {
	jwt.RegisteredClaims
	Login    string
}

// NewRouter creates and returns a new configured gRPC server with interceptors
func NewRouter(g *Transport) (*grpc.Server, error) {
	creds, err := credentials.NewServerTLSFromFile("server.crt", "server.key")
	if err != nil {
		return nil, err
	}

	server := grpc.NewServer(
		grpc.Creds(creds),
		grpc.ChainUnaryInterceptor(
			LoggingInterceptor(g.log),
			AuthInterceptor(g.cfg.JWTSecret),
		),
	)

	proto.RegisterVaultServiceServer(server, g.handler)

	return server, nil
}

// LoggingInterceptor adds request logging that records method, duration, and status
func LoggingInterceptor(log *slog.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		start := time.Now()
		method := info.FullMethod

		resp, err := handler(ctx, req)

		latency := time.Since(start)
		code := codes.OK
		if err != nil {
			if st, ok := status.FromError(err); ok {
				code = st.Code()
			} else {
				code = codes.Internal
			}
		}

		log.Info("request completed",
			"method", method,
			"duration", latency.String(),
			"status", code.String(),
		)

		return resp, err
	}
}

// AuthInterceptor handles JWT authentication via metadata
func AuthInterceptor(JWTSecret string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if info.FullMethod == registerMethod || info.FullMethod == loginMethod {
			return handler(ctx, req)
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Internal, "failed to get metadata from context")
		}

		var userID string

		authHeaders := md.Get("authorization")
		if len(authHeaders) > 0 {
			authHeader := authHeaders[0]
			if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
				tokenString := authHeader[7:]

				claims := &Claim{}
				token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
					if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
						return nil, status.Error(codes.InvalidArgument, "unexpected signing method")
					}
					return []byte(JWTSecret), nil
				})

				if err == nil && token.Valid {
					userID = claims.Login
					ctx = context.WithValue(ctx, userIDKey, userID)
					return handler(ctx, req)
				}
			}
		}

		return nil, status.Error(codes.Unauthenticated, "valid authentication required")
	}
}
