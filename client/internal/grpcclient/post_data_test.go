package grpcclient

import (
	"context"
	"data-vault/client/internal/models"
	"data-vault/client/internal/proto"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// PostData implements the mock PostData method
func (m *MockVaultServer) PostData(ctx context.Context, req *proto.PostDataRequest) (*proto.PostDataResponse, error) {
	fmt.Printf("DEBUG MockServer: PostData called with data length: %d, shouldSucceed: %t, validateJWT: %t\n", len(req.Data), m.shouldSucceed, m.validateJWT)

	if m.validateJWT {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			fmt.Printf("DEBUG MockServer: No metadata found\n")
			return nil, status.Error(codes.Unauthenticated, "no metadata found")
		}

		authHeaders := md.Get("authorization")
		if len(authHeaders) == 0 {
			fmt.Printf("DEBUG MockServer: No authorization header found\n")
			return nil, status.Error(codes.Unauthenticated, "no authorization header")
		}

		authHeader := authHeaders[0]
		if !strings.HasPrefix(authHeader, "Bearer ") {
			fmt.Printf("DEBUG MockServer: Invalid authorization header format\n")
			return nil, status.Error(codes.Unauthenticated, "invalid authorization header format")
		}

		token := authHeader[7:]
		login, valid := m.ValidateTestJWT(token)
		if !valid {
			fmt.Printf("DEBUG MockServer: Invalid JWT token\n")
			return nil, status.Error(codes.Unauthenticated, "invalid JWT token")
		}

		fmt.Printf("DEBUG MockServer: JWT validated successfully for user: %s\n", login)
	}

	if !m.shouldSucceed {
		fmt.Printf("DEBUG MockServer: shouldSucceed is false, returning failure\n")
		return &proto.PostDataResponse{
			Success: false,
		}, nil
	}

	if len(req.Data) == 0 {
		fmt.Printf("DEBUG MockServer: Empty data provided\n")
		return nil, status.Error(codes.InvalidArgument, "Data not provided")
	}

	fmt.Printf("DEBUG MockServer: PostData successful\n")
	return &proto.PostDataResponse{
		Success: true,
	}, nil
}

func TestDataVault_PostData_WithJWTIntegration(t *testing.T) {
	t.Parallel()

	jwtSecret := "test-secret-for-post-data"
	_, lis, cleanup := SetupMockServerWithJWT(true, "", true, jwtSecret)
	defer cleanup()

	client := SetupTestClient(t, lis)
	ctx := context.Background()

	user := models.User{
		Login:    "postdatauser",
		Password: "securepassword123",
	}

	jwt, err := client.Register(ctx, user)
	assert.NoError(t, err, "Registration should succeed")
	assert.NotEmpty(t, jwt, "JWT should not be empty")

	testData := "This is some test data to be stored securely"
	err = client.PostData(ctx, jwt, "text", []byte(testData))
	assert.NoError(t, err, "PostData should succeed with valid JWT")
}

func TestDataVault_PostData_WithInvalidJWT(t *testing.T) {
	t.Parallel()

	jwtSecret := "test-secret-for-post-data"
	_, lis, cleanup := SetupMockServerWithJWT(true, "", true, jwtSecret)
	defer cleanup()

	client := SetupTestClient(t, lis)
	ctx := context.Background()

	invalidJWT := "invalid.jwt.token"
	testData := "This is some test data"

	err := client.PostData(ctx, invalidJWT, "text", []byte(testData))
	assert.Error(t, err, "PostData should fail with invalid JWT")
}

func TestDataVault_PostData_WithExpiredJWT(t *testing.T) {
	t.Parallel()

	jwtSecret := "test-secret-for-post-data"
	_, lis, cleanup := SetupMockServerWithJWT(true, "", true, jwtSecret)
	defer cleanup()

	client := SetupTestClient(t, lis)
	ctx := context.Background()

	expiredJWT := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJsb2dpbiI6InRlc3R1c2VyIiwiZXhwIjoxNjAwMDAwMDAwLCJpYXQiOjE2MDAwMDAwMDB9.invalid"

	testData := "This is some test data"
	err := client.PostData(ctx, expiredJWT, "text", []byte(testData))
	assert.Error(t, err, "PostData should fail with expired JWT")
}

func TestDataVault_PostData_MultipleUsersWithDifferentJWTs(t *testing.T) {
	t.Parallel()

	jwtSecret := "test-secret-for-multi-user"
	_, lis, cleanup := SetupMockServerWithJWT(true, "", true, jwtSecret)
	defer cleanup()

	client := SetupTestClient(t, lis)
	ctx := context.Background()

	users := []models.User{
		{Login: "user1", Password: "password1"},
		{Login: "user2", Password: "password2"},
		{Login: "user3", Password: "password3"},
	}

	for i, user := range users {
		t.Run(fmt.Sprintf("user_%d", i+1), func(t *testing.T) {
			jwt, err := client.Register(ctx, user)
			require.NoError(t, err)
			require.NotEmpty(t, jwt)

			err = client.PostData(ctx, jwt, "text", []byte("test data for "+user.Login))
			require.NoError(t, err)
		})
	}
}

func TestDataVault_PostData_CrossSecretValidation(t *testing.T) {
	t.Parallel()

	_, lis1, cleanup1 := SetupMockServerWithJWT(true, "", true, "secret1")
	defer cleanup1()

	_, lis2, cleanup2 := SetupMockServerWithJWT(true, "", true, "secret2")
	defer cleanup2()

	client1 := SetupTestClient(t, lis1)
	client2 := SetupTestClient(t, lis2)
	ctx := context.Background()

	user := models.User{Login: "testuser", Password: "password123"}
	jwt, err := client1.Register(ctx, user)
	require.NoError(t, err)

	err = client2.PostData(ctx, jwt, "text", []byte("test data"))
	require.Error(t, err, "Should fail to use JWT from server1 with server2")
}

func TestDataVault_PostData_JWTCrossValidation(t *testing.T) {
	t.Parallel()

	jwtSecret1 := "secret1"
	jwtSecret2 := "secret2"

	_, lis1, cleanup1 := SetupMockServerWithJWT(true, "", true, jwtSecret1)
	defer cleanup1()
	client1 := SetupTestClient(t, lis1)

	_, lis2, cleanup2 := SetupMockServerWithJWT(true, "", true, jwtSecret2)
	defer cleanup2()
	client2 := SetupTestClient(t, lis2)

	ctx := context.Background()
	user := models.User{Login: "testuser", Password: "testpass"}

	jwt1, err := client1.Register(ctx, user)
	assert.NoError(t, err, "Registration should succeed with server1")

	testData := "Cross-validation test data"
	err = client2.PostData(ctx, jwt1, "text", []byte(testData))
	assert.Error(t, err, "PostData should fail when using JWT from different server")
}

func TestDataVault_PostData(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		setupFunc      func(t *testing.T) (*Client, context.Context, func())
		jwt            string
		data           string
		expectedError  bool
		validateResult func(t *testing.T, err error)
	}{
		{
			name: "successful post data with valid JWT and data",
			setupFunc: func(t *testing.T) (*Client, context.Context, func()) {
				expectedToken := "valid-jwt-token-123"
				_, lis, cleanup := SetupMockServer(true, expectedToken)
				client := SetupTestClient(t, lis)
				return client, context.Background(), cleanup
			},
			jwt:           "valid-jwt-token-123",
			data:          "sample encrypted data content",
			expectedError: false,
			validateResult: func(t *testing.T, err error) {
				assert.NoError(t, err, "Expected post data to succeed")
			},
		},
		{
			name: "failed post data with server failure",
			setupFunc: func(t *testing.T) (*Client, context.Context, func()) {
				_, lis, cleanup := SetupMockServer(false, "")
				client := SetupTestClient(t, lis)
				return client, context.Background(), cleanup
			},
			jwt:           "any-jwt-token",
			data:          "sample data",
			expectedError: true,
			validateResult: func(t *testing.T, err error) {
				assert.Error(t, err, "Expected post data to fail")
				assert.Contains(t, err.Error(), "failed to post data")
			},
		},
		{
			name: "failed post data with empty data",
			setupFunc: func(t *testing.T) (*Client, context.Context, func()) {
				expectedToken := "valid-jwt-token-123"
				_, lis, cleanup := SetupMockServer(true, expectedToken)
				client := SetupTestClient(t, lis)
				return client, context.Background(), cleanup
			},
			jwt:           "valid-jwt-token-123",
			data:          "",
			expectedError: true,
			validateResult: func(t *testing.T, err error) {
				assert.Error(t, err, "Expected post data to fail with empty data")
			},
		},
		{
			name: "failed post data with empty JWT",
			setupFunc: func(t *testing.T) (*Client, context.Context, func()) {
				expectedToken := "valid-jwt-token-123"
				_, lis, cleanup := SetupMockServer(true, expectedToken)
				client := SetupTestClient(t, lis)
				return client, context.Background(), cleanup
			},
			jwt:           "",
			data:          "sample data",
			expectedError: true,
			validateResult: func(t *testing.T, err error) {
				assert.Error(t, err, "Expected post data to fail with empty JWT")
			},
		},
		{
			name: "post data with large content",
			setupFunc: func(t *testing.T) (*Client, context.Context, func()) {
				expectedToken := "valid-jwt-token-456"
				_, lis, cleanup := SetupMockServer(true, expectedToken)
				client := SetupTestClient(t, lis)
				return client, context.Background(), cleanup
			},
			jwt:           "valid-jwt-token-456",
			data:          string(make([]byte, 10000)),
			expectedError: false,
			validateResult: func(t *testing.T, err error) {
				assert.NoError(t, err, "Expected post data with large content to succeed")
			},
		},
		{
			name: "post data with special characters",
			setupFunc: func(t *testing.T) (*Client, context.Context, func()) {
				expectedToken := "special-chars-token"
				_, lis, cleanup := SetupMockServer(true, expectedToken)
				client := SetupTestClient(t, lis)
				return client, context.Background(), cleanup
			},
			jwt:           "special-chars-token",
			data:          "special chars: !@#$%^&*()_+ äöü 中文 🚀",
			expectedError: false,
			validateResult: func(t *testing.T, err error) {
				assert.NoError(t, err, "Expected post data with special characters to succeed")
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			client, ctx, cleanup := tt.setupFunc(t)
			defer cleanup()

			err := client.PostData(ctx, tt.jwt, "text", []byte(tt.data))
			tt.validateResult(t, err)
			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDataVault_PostData_ContextCancellation(t *testing.T) {
	t.Parallel()

	expectedToken := "context-test-token"
	_, lis, cleanup := SetupMockServer(true, expectedToken)
	defer cleanup()

	client := SetupTestClient(t, lis)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := client.PostData(ctx, expectedToken, "text", []byte("test data"))

	assert.Error(t, err)
}

func TestDataVault_PostData_ContextTimeout(t *testing.T) {
	t.Parallel()

	expectedToken := "timeout-test-token"
	_, lis, cleanup := SetupMockServer(true, expectedToken)
	defer cleanup()

	client := SetupTestClient(t, lis)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	time.Sleep(1 * time.Millisecond)

	err := client.PostData(ctx, expectedToken, "text", []byte("test data"))

	assert.Error(t, err)
}

func TestDataVault_PostData_MultipleConsecutive(t *testing.T) {
	t.Parallel()

	expectedToken := "multi-post-token"
	_, lis, cleanup := SetupMockServer(true, expectedToken)
	defer cleanup()

	client := SetupTestClient(t, lis)
	ctx := context.Background()

	for i := 0; i < 3; i++ {
		data := fmt.Sprintf("test data item %d", i+1)
		err := client.PostData(ctx, expectedToken, "text", []byte(data))
		assert.NoError(t, err, "Expected post data %d to succeed", i+1)
	}
}
