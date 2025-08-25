package grpcclient

import (
	"context"
	"data-vault/client/internal/models"
	"data-vault/client/internal/proto"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Login implements the mock Login method
func (m *MockVaultServer) Login(ctx context.Context, req *proto.LoginRequest) (*proto.LoginResponse, error) {
	fmt.Printf("DEBUG MockServer: Login called with login: %s, shouldSucceed: %t\n", req.User.Login, m.shouldSucceed)

	if m.registeredUsers == nil {
		m.registeredUsers = make(map[string]string)
	}

	// If shouldSucceed is false, simulate server failure
	if !m.shouldSucceed {
		fmt.Printf("DEBUG MockServer: shouldSucceed is false, returning failure\n")
		return &proto.LoginResponse{
			Success:  false,
			JwtToken: "",
		}, nil
	}

	// Check if user exists and password matches
	if storedPassword, exists := m.registeredUsers[req.User.Login]; !exists || storedPassword != req.User.Password {
		fmt.Printf("DEBUG MockServer: Login failed - user not found or wrong password: %s\n", req.User.Login)
		return &proto.LoginResponse{
			Success:  false,
			JwtToken: "",
		}, nil
	}

	// Success case
	fmt.Printf("DEBUG MockServer: Login successful for: %s\n", req.User.Login)

	// Generate JWT token if validation is enabled
	var jwtToken string
	if m.validateJWT {
		jwtToken = m.GenerateTestJWT(req.User.Login)
		fmt.Printf("DEBUG MockServer: Generated JWT for %s: %s\n", req.User.Login, jwtToken)
	} else {
		jwtToken = m.expectedToken
	}

	return &proto.LoginResponse{
		Success:  true,
		JwtToken: jwtToken,
	}, nil
} // SetupMockServer creates a mock gRPC server for testing

func TestDataVault_Login(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		setupFunc      func(t *testing.T) (*Client, context.Context, func())
		user           models.User
		expectedToken  string
		expectedError  error
		validateResult func(t *testing.T, token string, err error)
	}{
		{
			name: "successful login with valid credentials",
			setupFunc: func(t *testing.T) (*Client, context.Context, func()) {
				expectedToken := "valid-jwt-token-12345"
				_, lis, cleanup := SetupMockServer(true, expectedToken)
				client := SetupTestClient(t, lis)

				// Pre-register a user for login test
				user := models.User{
					Login:    "testuser",
					Password: "securepassword123",
				}
				_, err := client.Register(context.Background(), user)
				require.NoError(t, err, "Failed to register test user")

				return client, context.Background(), cleanup
			},
			user: models.User{
				Login:    "testuser",
				Password: "securepassword123",
			},
			expectedToken: "valid-jwt-token-12345",
			expectedError: nil,
			validateResult: func(t *testing.T, token string, err error) {
				assert.NoError(t, err)
				assert.Equal(t, "valid-jwt-token-12345", token)
				assert.NotEmpty(t, token)
			},
		},
		{
			name: "failed login with wrong password",
			setupFunc: func(t *testing.T) (*Client, context.Context, func()) {
				expectedToken := "valid-jwt-token-12345"
				_, lis, cleanup := SetupMockServer(true, expectedToken)
				client := SetupTestClient(t, lis)

				// Pre-register a user with correct password
				user := models.User{
					Login:    "testuser2",
					Password: "correctpassword",
				}
				_, err := client.Register(context.Background(), user)
				require.NoError(t, err, "Failed to register test user")

				return client, context.Background(), cleanup
			},
			user: models.User{
				Login:    "testuser2",
				Password: "wrongpassword",
			},
			expectedToken: "",
			expectedError: ErrorLogin,
			validateResult: func(t *testing.T, token string, err error) {
				assert.Error(t, err)
				assert.Equal(t, ErrorLogin, err)
				assert.Empty(t, token)
			},
		},
		{
			name: "failed login with non-existent user",
			setupFunc: func(t *testing.T) (*Client, context.Context, func()) {
				expectedToken := "valid-jwt-token-12345"
				_, lis, cleanup := SetupMockServer(true, expectedToken)
				client := SetupTestClient(t, lis)
				return client, context.Background(), cleanup
			},
			user: models.User{
				Login:    "nonexistentuser",
				Password: "anypassword",
			},
			expectedToken: "",
			expectedError: ErrorLogin,
			validateResult: func(t *testing.T, token string, err error) {
				assert.Error(t, err)
				assert.Equal(t, ErrorLogin, err)
				assert.Empty(t, token)
			},
		},
		{
			name: "failed login with empty credentials",
			setupFunc: func(t *testing.T) (*Client, context.Context, func()) {
				expectedToken := "valid-jwt-token-12345"
				_, lis, cleanup := SetupMockServer(true, expectedToken)
				client := SetupTestClient(t, lis)
				return client, context.Background(), cleanup
			},
			user: models.User{
				Login:    "",
				Password: "",
			},
			expectedToken: "",
			expectedError: ErrorLogin,
			validateResult: func(t *testing.T, token string, err error) {
				assert.Error(t, err)
				assert.Equal(t, ErrorLogin, err)
				assert.Empty(t, token)
			},
		},
		{
			name: "failed login with empty login",
			setupFunc: func(t *testing.T) (*Client, context.Context, func()) {
				expectedToken := "valid-jwt-token-12345"
				_, lis, cleanup := SetupMockServer(true, expectedToken)
				client := SetupTestClient(t, lis)
				return client, context.Background(), cleanup
			},
			user: models.User{
				Login:    "",
				Password: "validpassword123",
			},
			expectedToken: "",
			expectedError: ErrorLogin,
			validateResult: func(t *testing.T, token string, err error) {
				assert.Error(t, err)
				assert.Equal(t, ErrorLogin, err)
				assert.Empty(t, token)
			},
		},
		{
			name: "failed login with empty password",
			setupFunc: func(t *testing.T) (*Client, context.Context, func()) {
				expectedToken := "valid-jwt-token-12345"
				_, lis, cleanup := SetupMockServer(true, expectedToken)
				client := SetupTestClient(t, lis)

				// Pre-register a user
				user := models.User{
					Login:    "testuser3",
					Password: "validpassword123",
				}
				_, err := client.Register(context.Background(), user)
				require.NoError(t, err, "Failed to register test user")

				return client, context.Background(), cleanup
			},
			user: models.User{
				Login:    "testuser3",
				Password: "",
			},
			expectedToken: "",
			expectedError: ErrorLogin,
			validateResult: func(t *testing.T, token string, err error) {
				assert.Error(t, err)
				assert.Equal(t, ErrorLogin, err)
				assert.Empty(t, token)
			},
		},
		{
			name: "server failure during login",
			setupFunc: func(t *testing.T) (*Client, context.Context, func()) {
				// Setup server that always fails
				_, lis, cleanup := SetupMockServer(false, "")
				client := SetupTestClient(t, lis)
				return client, context.Background(), cleanup
			},
			user: models.User{
				Login:    "anyuser",
				Password: "anypassword",
			},
			expectedToken: "",
			expectedError: ErrorLogin,
			validateResult: func(t *testing.T, token string, err error) {
				assert.Error(t, err)
				assert.Equal(t, ErrorLogin, err)
				assert.Empty(t, token)
			},
		},
		{
			name: "login with special characters in credentials",
			setupFunc: func(t *testing.T) (*Client, context.Context, func()) {
				expectedToken := "special-chars-token-789"
				_, lis, cleanup := SetupMockServer(true, expectedToken)
				client := SetupTestClient(t, lis)

				// Pre-register a user with special characters
				user := models.User{
					Login:    "user@example.com",
					Password: "p@ssw0rd!#$%",
				}
				_, err := client.Register(context.Background(), user)
				require.NoError(t, err, "Failed to register test user")

				return client, context.Background(), cleanup
			},
			user: models.User{
				Login:    "user@example.com",
				Password: "p@ssw0rd!#$%",
			},
			expectedToken: "special-chars-token-789",
			expectedError: nil,
			validateResult: func(t *testing.T, token string, err error) {
				assert.NoError(t, err)
				assert.Equal(t, "special-chars-token-789", token)
				assert.NotEmpty(t, token)
			},
		},
	}

	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			client, ctx, cleanup := tt.setupFunc(t)
			defer cleanup()

			// Execute the login function
			token, err := client.Login(ctx, tt.user)

			// Validate the results
			tt.validateResult(t, token, err)
		})
	}
}

func TestDataVault_Login_ServerFailure(t *testing.T) {
	_, lis, cleanup := SetupMockServer(false, "")
	defer cleanup()

	client := SetupTestClient(t, lis)

	user := models.User{
		Login:    "test_user",
		Password: "test_password",
	}

	ctx := context.Background()
	token, err := client.Login(ctx, user)

	assert.Error(t, err)
	assert.Equal(t, ErrorLogin, err)
	assert.Empty(t, token)
}

func TestDataVault_Login_ContextCancellation(t *testing.T) {
	t.Parallel()

	expectedToken := "context-test-token"
	_, lis, cleanup := SetupMockServer(true, expectedToken)
	defer cleanup()

	client := SetupTestClient(t, lis)

	// Register a user first
	user := models.User{
		Login:    "contextuser",
		Password: "contextpassword",
	}
	_, err := client.Register(context.Background(), user)
	require.NoError(t, err)

	// Test with cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	token, err := client.Login(ctx, user)

	// Should get an error due to cancelled context
	assert.Error(t, err)
	assert.Empty(t, token)
}

func TestDataVault_Login_ContextTimeout(t *testing.T) {
	t.Parallel()

	expectedToken := "timeout-test-token"
	_, lis, cleanup := SetupMockServer(true, expectedToken)
	defer cleanup()

	client := SetupTestClient(t, lis)

	// Register a user first
	user := models.User{
		Login:    "timeoutuser",
		Password: "timeoutpassword",
	}
	_, err := client.Register(context.Background(), user)
	require.NoError(t, err)

	// Test with very short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	// Wait a bit to ensure timeout
	time.Sleep(1 * time.Millisecond)

	token, err := client.Login(ctx, user)

	// Should get an error due to timeout
	assert.Error(t, err)
	assert.Empty(t, token)
}

func TestDataVault_Login_WithJWTValidation(t *testing.T) {
	t.Parallel()

	jwtSecret := "test-secret-for-login"
	_, lis, cleanup := SetupMockServerWithJWT(true, "", true, jwtSecret)
	defer cleanup()

	client := SetupTestClient(t, lis)

	// Register a user first
	user := models.User{
		Login:    "jwtloginuser",
		Password: "jwtloginpass",
	}
	_, err := client.Register(context.Background(), user)
	require.NoError(t, err)

	// Test login with JWT validation
	jwt, err := client.Login(context.Background(), user)
	assert.NoError(t, err, "Login should succeed")
	assert.NotEmpty(t, jwt, "JWT should not be empty")
	assert.Contains(t, jwt, ".", "JWT should contain dots")

	// Validate the returned JWT
	mockServer := &MockVaultServer{
		validateJWT: true,
		jwtSecret:   jwtSecret,
	}

	login, valid := mockServer.ValidateTestJWT(jwt)
	assert.True(t, valid, "JWT should be valid")
	assert.Equal(t, user.Login, login, "JWT should contain correct login")
}

func TestDataVault_Login_JWTConsistency(t *testing.T) {
	t.Parallel()

	jwtSecret := "consistent-secret"
	_, lis, cleanup := SetupMockServerWithJWT(true, "", true, jwtSecret)
	defer cleanup()

	client := SetupTestClient(t, lis)

	user := models.User{
		Login:    "consistentuser",
		Password: "consistentpass",
	}

	// Register user
	registerJWT, err := client.Register(context.Background(), user)
	require.NoError(t, err)

	// Login with same user
	loginJWT, err := client.Login(context.Background(), user)
	require.NoError(t, err)

	// Both JWTs should be valid and contain same login
	mockServer := &MockVaultServer{
		validateJWT: true,
		jwtSecret:   jwtSecret,
	}

	regLogin, regValid := mockServer.ValidateTestJWT(registerJWT)
	assert.True(t, regValid, "Register JWT should be valid")
	assert.Equal(t, user.Login, regLogin, "Register JWT should contain correct login")

	loginLogin, loginValid := mockServer.ValidateTestJWT(loginJWT)
	assert.True(t, loginValid, "Login JWT should be valid")
	assert.Equal(t, user.Login, loginLogin, "Login JWT should contain correct login")

	// Both should contain the same login
	assert.Equal(t, regLogin, loginLogin, "Both JWTs should contain same login")
}
