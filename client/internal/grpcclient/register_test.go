package grpcclient

import (
	"context"
	"data-vault/client/internal/models"
	"data-vault/client/internal/proto"
	"fmt"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func (m *MockVaultServer) Register(ctx context.Context, req *proto.RegisterRequest) (*proto.RegisterResponse, error) {
	fmt.Printf("DEBUG MockServer: Register called with login: %s, shouldSucceed: %t\n", req.User.Login, m.shouldSucceed)

	if m.registeredUsers == nil {
		m.registeredUsers = make(map[string]string)
	}

	if _, exists := m.registeredUsers[req.User.Login]; exists {
		fmt.Printf("DEBUG MockServer: Duplicate user detected: %s\n", req.User.Login)
		return &proto.RegisterResponse{
			Success:  false,
			JwtToken: "",
		}, nil
	}

	if !m.shouldSucceed {
		fmt.Printf("DEBUG MockServer: shouldSucceed is false, returning failure\n")
		return &proto.RegisterResponse{
			Success:  false,
			JwtToken: "",
		}, nil
	}

	fmt.Printf("DEBUG MockServer: Registration successful for: %s\n", req.User.Login)
	m.registeredUsers[req.User.Login] = req.User.Password

	var jwtToken string
	if m.validateJWT {
		jwtToken = m.GenerateTestJWT(req.User.Login)
		fmt.Printf("DEBUG MockServer: Generated JWT for %s: %s\n", req.User.Login, jwtToken)
	} else {
		jwtToken = m.expectedToken
	}

	return &proto.RegisterResponse{
		Success:  true,
		JwtToken: jwtToken,
	}, nil
}

func TestDataVault_Register(t *testing.T) {
	expectedToken := "test-jwt-token-123"
	_, lis, cleanup := SetupMockServer(true, expectedToken)
	defer cleanup()

	client := SetupTestClient(t, lis)

	tests := []struct {
		name     string
		user     models.User
		wantErr  bool
		errCheck func(error) bool
	}{
		{
			name: "successful registration with random data",
			user: models.User{
				Login:    gofakeit.Username(),
				Password: gofakeit.Password(true, true, true, true, false, 10),
			},
			wantErr: false,
		},
		{
			name: "random email as login",
			user: models.User{
				Login:    gofakeit.Email(),
				Password: gofakeit.Password(true, true, true, true, false, 10),
			},
			wantErr: false,
		},
		{
			name: "very long login",
			user: models.User{
				Login:    gofakeit.LetterN(50),
				Password: gofakeit.Password(true, true, true, true, false, 10),
			},
			wantErr: false,
		},
		{
			name: "special characters in login",
			user: models.User{
				Login:    gofakeit.Username() + "@#$%",
				Password: gofakeit.Password(true, true, true, true, false, 10),
			},
			wantErr: false,
		},
		{
			name: "very long password",
			user: models.User{
				Login:    gofakeit.Username(),
				Password: gofakeit.Password(true, true, true, true, false, 100),
			},
			wantErr: false,
		},
		{
			name: "empty login",
			user: models.User{
				Login:    "",
				Password: gofakeit.Password(true, true, true, true, false, 10),
			},
			wantErr: true,
		},
		{
			name: "empty password",
			user: models.User{
				Login:    gofakeit.Username(),
				Password: "",
			},
			wantErr: true,
		},
		{
			name: "both empty credentials",
			user: models.User{
				Login:    "",
				Password: "",
			},
			wantErr: true,
		},
	}

	ctx := context.Background()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := client.Register(ctx, tt.user)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errCheck != nil {
					assert.True(t, tt.errCheck(err))
				}
				assert.Empty(t, token)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, expectedToken, token)
			}
		})
	}
}

func TestDataVault_Register_WithJWTValidation(t *testing.T) {
	t.Parallel()

	jwtSecret := "test-secret-for-register"
	_, lis, cleanup := SetupMockServerWithJWT(true, "", true, jwtSecret)
	defer cleanup()

	client := SetupTestClient(t, lis)

	user := models.User{
		Login:    "jwttestuser",
		Password: "jwtpassword123",
	}

	jwt, err := client.Register(context.Background(), user)
	assert.NoError(t, err, "Registration should succeed")
	assert.NotEmpty(t, jwt, "JWT should not be empty")
	assert.Contains(t, jwt, ".", "JWT should contain dots")

	mockServer := &MockVaultServer{
		validateJWT: true,
		jwtSecret:   jwtSecret,
	}

	login, valid := mockServer.ValidateTestJWT(jwt)
	assert.True(t, valid, "JWT should be valid")
	assert.Equal(t, user.Login, login, "JWT should contain correct login")
}

func TestDataVault_Register_JWTIntegrity(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		jwtSecret string
		user      models.User
	}{
		{
			name:      "user with alphanumeric login",
			jwtSecret: "secret1",
			user: models.User{
				Login:    "user123",
				Password: "pass123",
			},
		},
		{
			name:      "user with email login",
			jwtSecret: "secret2",
			user: models.User{
				Login:    "user@example.com",
				Password: "securepass",
			},
		},
		{
			name:      "user with special characters",
			jwtSecret: "secret3",
			user: models.User{
				Login:    "user-test_2024",
				Password: "p@ssw0rd!",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, lis, cleanup := SetupMockServerWithJWT(true, "", true, tt.jwtSecret)
			defer cleanup()

			client := SetupTestClient(t, lis)

			jwt, err := client.Register(context.Background(), tt.user)
			assert.NoError(t, err, "Registration should succeed")
			assert.NotEmpty(t, jwt, "JWT should not be empty")

			mockServer := &MockVaultServer{
				validateJWT: true,
				jwtSecret:   tt.jwtSecret,
			}

			login, valid := mockServer.ValidateTestJWT(jwt)
			assert.True(t, valid, "JWT should be valid")
			assert.Equal(t, tt.user.Login, login, "JWT should contain correct login")
		})
	}
}

func TestDataVault_Register_DuplicateUser(t *testing.T) {
	expectedToken := "test-jwt-token-123"
	_, lis, cleanup := SetupMockServer(true, expectedToken)
	defer cleanup()

	client := SetupTestClient(t, lis)

	baseUser := models.User{
		Login:    "duplicate_user_test_login",
		Password: gofakeit.Password(true, true, true, true, false, 10),
	}

	ctx := context.Background()

	token, err := client.Register(ctx, baseUser)
	require.NoError(t, err)
	require.Equal(t, expectedToken, token)

	token, err = client.Register(ctx, baseUser)
	assert.Error(t, err)
	assert.Equal(t, ErrorRegister, err)
	assert.Empty(t, token)
}

func TestDataVault_Register_ServerFailure(t *testing.T) {
	_, lis, cleanup := SetupMockServer(false, "")
	defer cleanup()

	client := SetupTestClient(t, lis)

	user := models.User{
		Login:    gofakeit.Username(),
		Password: gofakeit.Password(true, true, true, true, false, 10),
	}

	ctx := context.Background()
	token, err := client.Register(ctx, user)

	assert.Error(t, err)
	assert.Equal(t, ErrorRegister, err)
	assert.Empty(t, token)
}
