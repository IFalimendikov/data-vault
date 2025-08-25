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

// MockService is a mock implementation of the Service interface
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

// Test helper functions

// setupTestHandler creates a test handler with mock service
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

// setupTestHandlerWithConfig creates a test handler with custom config
func setupTestHandlerWithConfig(cfg config.Config) (*Handler, *MockService) {
	mockService := &MockService{}
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	handler := New(context.Background(), mockService, cfg, logger)
	return handler, mockService
}

// createTestUser creates a test user model
func createTestUser(login, password string) models.User {
	return models.User{
		Login:    login,
		Password: password,
	}
}

// createTestData creates test data models
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

// createContextWithUser creates a context with user ID
func createContextWithUser(userID string) context.Context {
	return context.WithValue(context.Background(), userIDKey, userID)
}

// validateJWTToken validates a JWT token and returns the claims
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

// Common test scenarios

// TestHandlerCreation tests the handler constructor
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

// TestJWTIssuance tests the JWT token issuance functionality
func TestJWTIssuance(t *testing.T) {
	handler, _ := setupTestHandler()

	user := createTestUser("testuser", "testpass")

	token, err := handler.IssueJWT(user)
	require.NoError(t, err)
	assert.NotEmpty(t, token)

	// Validate the token
	validateJWTToken(t, token, handler.cfg.JWTSecret, user.Login)
}

// TestJWTIssuance_EmptySecret tests JWT creation with empty secret
func TestJWTIssuance_EmptySecret(t *testing.T) {
	cfg := config.Config{
		ServerAddr:    "localhost:8080",
		JWTSecret:     "", // Empty secret
		EncryptionKey: "test-encryption-key",
		DatabaseURI:   "test-db-uri",
	}
	handler, _ := setupTestHandlerWithConfig(cfg)

	user := createTestUser("testuser", "testpass")

	token, err := handler.IssueJWT(user)
	require.NoError(t, err) // JWT creation should still work with empty secret
	assert.NotEmpty(t, token)
}

// TestJWTIssuance_SpecialCharacters tests JWT with special characters in login
func TestJWTIssuance_SpecialCharacters(t *testing.T) {
	handler, _ := setupTestHandler()

	user := createTestUser("test@user.com", "test-pass-123")

	token, err := handler.IssueJWT(user)
	require.NoError(t, err)
	assert.NotEmpty(t, token)

	// Validate the token
	validateJWTToken(t, token, handler.cfg.JWTSecret, user.Login)
}

// TestContextKey tests the context key functionality
func TestContextKey(t *testing.T) {
	ctx := createContextWithUser("testuser")

	value, ok := ctx.Value(userIDKey).(string)
	assert.True(t, ok)
	assert.Equal(t, "testuser", value)

	// Test with wrong type
	ctx2 := context.WithValue(context.Background(), userIDKey, 123)
	value2, ok2 := ctx2.Value(userIDKey).(string)
	assert.False(t, ok2)
	assert.Empty(t, value2)

	// Test with empty context
	value3, ok3 := context.Background().Value(userIDKey).(string)
	assert.False(t, ok3)
	assert.Empty(t, value3)
}

// TestContextKeyType tests the context key type
func TestContextKeyType(t *testing.T) {
	key := userIDKey
	assert.Equal(t, contextKey("user_id"), key)
	assert.Equal(t, "user_id", string(key))
}

// TestConstants tests the package constants
func TestConstants(t *testing.T) {
	assert.Equal(t, 24, tokenExpiryHours)
	assert.Equal(t, contextKey("user_id"), userIDKey)
}

// Test data factories

// testUserValid returns a valid test user
func testUserValid() models.User {
	return createTestUser("testuser", "testpass123")
}

// testUserEmptyLogin returns a user with empty login
func testUserEmptyLogin() models.User {
	return createTestUser("", "testpass123")
}

// testUserEmptyPassword returns a user with empty password
func testUserEmptyPassword() models.User {
	return createTestUser("testuser", "")
}

// testDataSample returns sample test data
func testDataSample() []models.Data {
	return createTestData("testuser", 2)
}

// testDataLarge returns a large sample of test data
func testDataLarge() []models.Data {
	return createTestData("testuser", 100)
}

// testDataEmpty returns empty test data
func testDataEmpty() []models.Data {
	return []models.Data{}
}
