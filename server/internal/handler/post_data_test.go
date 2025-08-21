package handler

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ShiraazMoollatjie/goluhn"
	"github.com/brianvoe/gofakeit/v7"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gofermart/internal/config"
	"gofermart/internal/models"
	"gofermart/internal/service"
	"gofermart/internal/storage"
	"log/slog"
	"strconv"
)

func setupPostOrdersTestDB(t *testing.T) *sql.DB {
	cfg := config.Config{}

	err := config.Read(&cfg)
	require.NoError(t, err, "Failed to read config")
	db, err := sql.Open("postgres", cfg.DatabaseURI)
	require.NoError(t, err)
	err = db.Ping()
	require.NoError(t, err)

	_, err = db.Exec(storage.OrdersQuery)
	require.NoError(t, err)
	_, err = db.Exec(storage.UsersQuery)
	require.NoError(t, err)
	_, err = db.Exec(storage.BalancesQuery)
	require.NoError(t, err)
	return db
}

func TestGofermart_PostOrders(t *testing.T) {
	gofakeit.Seed(0)

	db := setupPostOrdersTestDB(t)
	defer db.Close()
	defer func() {
		_, err := db.Exec(`DROP TABLE IF EXISTS orders`)
		require.NoError(t, err)
		_, err = db.Exec(`DROP TABLE IF EXISTS users`)
		require.NoError(t, err)
		_, err = db.Exec(`DROP TABLE IF EXISTS balances`)
		require.NoError(t, err)
	}()

	storage := &storage.Storage{DB: db}
	service := &service.Gofermart{
		Storage: storage,
		Log:     slog.Default(),
	}
	handler := &Handler{
		Service: service,
	}

	testUsers := []models.User{
		{
			Login:    gofakeit.Username(),
			Password: gofakeit.Password(true, true, true, true, false, 10),
		},
		{
			Login:    gofakeit.Username(),
			Password: gofakeit.Password(true, true, true, true, false, 10),
		},
	}

	for _, user := range testUsers {
		err := service.Register(context.Background(), user)
		require.NoError(t, err, "Failed to register test user")
	}

	validOrderNum, _ := strconv.Atoi(goluhn.Generate(8))

	tests := []struct {
		name       string
		orderNum   int
		login      string
		setupFunc  func(*testing.T, *sql.DB, string)
		wantStatus int
	}{
		{
			name:       "valid order number",
			orderNum:   validOrderNum,
			login:      testUsers[0].Login,
			wantStatus: http.StatusAccepted,
		},
		{
			name:       "invalid Luhn number",
			orderNum:   gofakeit.Number(10000, 99999),
			login:      testUsers[0].Login,
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name:     "duplicate order same user",
			orderNum: 17893729974,
			login:    "testuser1",
			setupFunc: func(t *testing.T, db *sql.DB, login string) {
				_, err := db.Exec(`
                INSERT INTO orders (number, login, status, uploaded_at)
                VALUES ($1, $2, 'NEW', NOW())
                `, 17893729974, login)
				require.NoError(t, err)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:     "duplicate order different user",
			orderNum: 17893729974,
			login:    "testuser2",
			setupFunc: func(t *testing.T, db *sql.DB, login string) {
				_, err := db.Exec(`
                INSERT INTO orders (number, login, status, uploaded_at)
                VALUES ($1, $2, 'NEW', NOW())
                `, 17893729974, "testuser1")
				require.NoError(t, err)
			},
			wantStatus: http.StatusConflict,
		},
		{
			name:       "empty login number",
			orderNum:   123,
			login:      "",
			wantStatus: http.StatusUnprocessableEntity,
		},
	}

	gin.SetMode(gin.TestMode)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := db.Exec(`DELETE FROM orders`)
			require.NoError(t, err)
			if tt.setupFunc != nil {
				tt.setupFunc(t, db, tt.login)
			}

			router := gin.New()

			router.Use(func(c *gin.Context) {
				c.Set("login", tt.login)
				c.Next()
			})

			router.POST("/api/user/orders", func(c *gin.Context) {
				handler.PostOrders(c)
			})

			orderJSON, err := json.Marshal(tt.orderNum)
			require.NoError(t, err)

			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/api/user/orders", bytes.NewBuffer(orderJSON))
			req.Header.Set("Content-Type", "application/json")
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code, "Test case: %s - Expected status %d but got %d", tt.name, tt.wantStatus, w.Code)

			if tt.login != "" {
				_, err = db.Exec(`DELETE FROM orders WHERE login = $1`, tt.login)
				require.NoError(t, err)
			}
		})
	}
}