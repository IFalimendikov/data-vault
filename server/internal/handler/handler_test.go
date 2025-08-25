package handler

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"testing"
	"time"

	"data-vault/server/internal/config"
	"data-vault/server/internal/models"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockService struct {
	mock.Mock
}

func (m *MockService) Register(ctx context.Context, user models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockService) Login(ctx context.Context, user models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockService) PostData(ctx context.Context, login, dataType string, data []byte) error {
	args := m.Called(ctx, login, dataType, data)
	return args.Error(0)
}

func (m *MockService) GetData(ctx context.Context, login string) ([]models.Data, error) {
	args := m.Called(ctx, login)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Data), args.Error(1)
}

func (m *MockService) DeleteData(ctx context.Context, login, id string) error {
	args := m.Called(ctx, login, id)
	return args.Error(0)
}

func setupTestHandler() (*Handler, *MockService) {
	mockService := &MockService{}
	cfg := config.Config{
		ServerAddr:    "localhost:8080",
		JWTSecret:     "test-jwt-secret-key-for-testing",
		EncryptionKey: "test-encryption-key",
		DatabaseURI:   "test-db-uri",
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	handler := New(context.Background(), mockService, cfg, logger)
	return handler, mockService
}

func setupTestHandlerWithConfig(cfg config.Config) (*Handler, *MockService) {
	mockService := &MockService{}
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	handler := New(context.Background(), mockService, cfg, logger)
	return handler, mockService
}

func createTestUser(login, password string) models.User {
	return models.User{
		Login:    login,
		Password: password,
	}
}

func createTestData(userLogin string, count int) []models.Data {
	data := make([]models.Data, count)
	for i := 0; i < count; i++ {
		data[i] = models.Data{
			ID:         fmt.Sprintf("data%d", i+1),
			User:       userLogin,
			Status:     "active",
			Type:       "text",
			Data:       []byte(fmt.Sprintf("encrypted data %d", i+1)),
			UploadedAt: time.Now().Add(-time.Duration(i) * time.Hour).Format(time.RFC3339),
		}
	}
	return data
}

func createContextWithUser(userID string) context.Context {
	return context.WithValue(context.Background(), userIDKey, userID)
}

func validateJWTToken(t *testing.T, tokenString, secret, expectedLogin string) *Claim {
	token, err := jwt.ParseWithClaims(tokenString, &Claim{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	require.NoError(t, err)
	require.True(t, token.Valid)

	claims, ok := token.Claims.(*Claim)
	require.True(t, ok)
	assert.Equal(t, expectedLogin, claims.Login)
	assert.True(t, claims.ExpiresAt.Time.After(time.Now()))
	assert.True(t, claims.IssuedAt.Time.Before(time.Now().Add(time.Minute)))

	return claims
}

func TestHandlerCreation(t *testing.T) {
	mockService := &MockService{}
	cfg := config.Config{
		ServerAddr:    "localhost:8080",
		JWTSecret:     "test-jwt-secret",
		EncryptionKey: "test-encryption-key",
		DatabaseURI:   "test-db-uri",
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	handler := New(context.Background(), mockService, cfg, logger)

	assert.NotNil(t, handler)
	assert.Equal(t, mockService, handler.service)
	assert.Equal(t, cfg, handler.cfg)
	assert.Equal(t, logger, handler.log)
}

func TestJWTIssuance(t *testing.T) {
	handler, _ := setupTestHandler()

	user := createTestUser("testuser", "testpass")

	token, err := handler.IssueJWT(user)
	require.NoError(t, err)
	assert.NotEmpty(t, token)

	validateJWTToken(t, token, handler.cfg.JWTSecret, user.Login)
}

func TestJWTIssuance_EmptySecret(t *testing.T) {
	cfg := config.Config{
		ServerAddr:    "localhost:8080",
		JWTSecret:     "",
		EncryptionKey: "test-encryption-key",
		DatabaseURI:   "test-db-uri",
	}
	handler, _ := setupTestHandlerWithConfig(cfg)

	user := createTestUser("testuser", "testpass")

	token, err := handler.IssueJWT(user)
	require.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestJWTIssuance_SpecialCharacters(t *testing.T) {
	handler, _ := setupTestHandler()

	user := createTestUser("test@user.com", "test-pass-123")

	token, err := handler.IssueJWT(user)
	require.NoError(t, err)
	assert.NotEmpty(t, token)

	validateJWTToken(t, token, handler.cfg.JWTSecret, user.Login)
}

func TestContextKey(t *testing.T) {
	ctx := createContextWithUser("testuser")

	value, ok := ctx.Value(userIDKey).(string)
	assert.True(t, ok)
	assert.Equal(t, "testuser", value)

	ctx2 := context.WithValue(context.Background(), userIDKey, 123)
	value2, ok2 := ctx2.Value(userIDKey).(string)
	assert.False(t, ok2)
	assert.Empty(t, value2)

	ctx3 := context.Background()
	value3, ok3 := ctx3.Value(userIDKey).(string)
	assert.False(t, ok3)
	assert.Empty(t, value3)
}

func TestContextKeyType(t *testing.T) {
	key := userIDKey
	assert.Equal(t, contextKey("user_id"), key)
	assert.Equal(t, "user_id", string(key))
}

func TestConstants(t *testing.T) {
	assert.Equal(t, 24, tokenExpiryHours)
	assert.Equal(t, contextKey("user_id"), userIDKey)
}

func testUserValid() models.User {
	return createTestUser("testuser", "testpass123")
}

func testUserEmptyLogin() models.User {
	return createTestUser("", "testpass123")
}

func testUserEmptyPassword() models.User {
	return createTestUser("testuser", "")
}

func testDataSample() []models.Data {
	return createTestData("testuser", 2)
}

func testDataLarge() []models.Data {
	return createTestData("testuser", 100)
}

func testDataEmpty() []models.Data {
	return []models.Data{}
}