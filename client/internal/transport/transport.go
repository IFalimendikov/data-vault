package transport

import (
	"context"
	"log/slog"
	"time"

	"data-vault/client/internal/config"
	"data-vault/client/internal/handler"
	"data-vault/client/internal/models"

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
)

type Service interface {
	Register(ctx context.Context, user models.User) (string, error)
	Login(ctx context.Context, user models.User) (string, error)
	PostData(ctx context.Context, jwt, data string) error
	GetData(ctx context.Context, jwt string) ([]models.Data, error)
	DeleteData(ctx context.Context, jwt, id string) error
	PingServer(ctx context.Context) bool
}

type Transport struct {
	Handler *handler.Handler
	Log     *slog.Logger
	Config  config.Config
}

func New(cfg config.Config, h *handler.Handler, log *slog.Logger) *Transport {
	return &Transport{
		Handler: h,
		Log:     log,
	}
}

func (t *Transport) NewRouter() *gin.Engine {
	r := gin.Default()

	r.Use(gin.Recovery())
	r.Use(t.withLogging())
	r.Use(gzip.Gzip(gzip.DefaultCompression, gzip.WithDecompressFn(gzip.DefaultDecompressHandle)))

	authorized := r.Group("api/user")

	r.POST("/register", func(c *gin.Context) {
		t.Handler.Register(c)
	})

	r.POST("/login", func(c *gin.Context) {
		t.Handler.Login(c)
	})

	authorized.POST("/orders", func(c *gin.Context) {
		t.Handler.PostData(c)
	})

	authorized.GET("/orders", func(c *gin.Context) {
		t.Handler.GetData(c)
	})

	authorized.GET("/balance", func(c *gin.Context) {
		t.Handler.DeleteData(c)
	})

	return r
}

func (t *Transport) withLogging() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now().UTC()
		uri := c.Request.RequestURI
		method := c.Request.Method

		c.Next()

		status := c.Writer.Status()
		size := c.Writer.Size()
		latency := time.Since(start)

		t.Log.Info("request completed",
			"uri", uri,
			"method", method,
			"duration", latency.String(),
			"status", status,
			"size", size,
		)
		c.Next()
	}
}
