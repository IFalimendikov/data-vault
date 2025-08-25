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
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// PostData implements the mock PostData method
func (m *MockVaultServer) PostData(ctx context.Context, req *proto.PostDataRequest) (*proto.PostDataResponse, error) {
	fmt.Printf("DEBUG MockServer: PostData called with data length: %d, shouldSucceed: %t, validateJWT: %t\n", len(req.Data), m.shouldSucceed, m.validateJWT)

	// Check JWT token from metadata if validation is enabled
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

		token := authHeader[7:] // Remove "Bearer " prefix
		login, valid := m.ValidateTestJWT(token)
		if !valid {
			fmt.Printf("DEBUG MockServer: Invalid JWT token\n")
			return nil, status.Error(codes.Unauthenticated, "invalid JWT token")
		}

		fmt.Printf("DEBUG MockServer: JWT validated successfully for user: %s\n", login)
	}

	// If shouldSucceed is false, simulate server failure
	if !m.shouldSucceed {
		fmt.Printf("DEBUG MockServer: shouldSucceed is false, returning failure\n")
		return &proto.PostDataResponse{
			Success: false,
		}, nil
	}

	// Check if data is empty
	if req.Data == "" {
		fmt.Printf("DEBUG MockServer: Empty data provided\n")
		return nil, status.Error(codes.InvalidArgument, "Data not provided")
	}

	// Success case
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

	// First, register a user to get a real JWT token
	user := models.User{
		Login:    "postdatauser",
		Password: "securepassword123",
	}

	jwt, err := client.Register(ctx, user)
	assert.NoError(t, err, "Registration should succeed")
	assert.NotEmpty(t, jwt, "JWT should not be empty")

	// Now test posting data with the real JWT token
	testData := "This is some test data to be stored securely"
	err = client.PostData(ctx, jwt, testData)
	assert.NoError(t, err, "PostData should succeed with valid JWT")
}

func TestDataVault_PostData_WithInvalidJWT(t *testing.T) {
	t.Parallel()

	jwtSecret := "test-secret-for-post-data"
	_, lis, cleanup := SetupMockServerWithJWT(true, "", true, jwtSecret)
	defer cleanup()

	client := SetupTestClient(t, lis)
	ctx := context.Background()

	// Test with invalid JWT
	invalidJWT := "invalid.jwt.token"
	testData := "This is some test data"

	err := client.PostData(ctx, invalidJWT, testData)
	assert.Error(t, err, "PostData should fail with invalid JWT")
}

func TestDataVault_PostData_WithExpiredJWT(t *testing.T) {
	t.Parallel()

	jwtSecret := "test-secret-for-post-data"
	_, lis, cleanup := SetupMockServerWithJWT(true, "", true, jwtSecret)
	defer cleanup()

	client := SetupTestClient(t, lis)
	ctx := context.Background()

	// Test with a malformed JWT that will be rejected
	expiredJWT := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJsb2dpbiI6InRlc3R1c2VyIiwiZXhwIjoxNjAwMDAwMDAwLCJpYXQiOjE2MDAwMDAwMDB9.invalid"

	testData := "This is some test data"
	err := client.PostData(ctx, expiredJWT, testData)
	assert.Error(t, err, "PostData should fail with expired JWT")
}

func TestDataVault_PostData_MultipleUsersWithDifferentJWTs(t *testing.T) {
	t.Parallel()

	jwtSecret := "test-secret-for-multi-user"
	_, lis, cleanup := SetupMockServerWithJWT(true, "", true, jwtSecret)
	defer cleanup()

	client := SetupTestClient(t, lis)
	ctx := context.Background()

	// Register multiple users
	users := []models.User{
		{Login: "user1", Password: "password1"},
		{Login: "user2", Password: "password2"},
		{Login: "user3", Password: "password3"},
	}

	// Test each user can post data with their own JWT
	for i, user := range users {
		t.Run(fmt.Sprintf("user_%d", i+1), func(t *testing.T) {
			// Register user and get JWT
			jwt, err := client.Register(ctx, user)
			assert.NoError(t, err, "Registration should succeed for %s", user.Login)
			assert.NotEmpty(t, jwt, "JWT should not be empty for %s", user.Login)

			// Post data with user's JWT
			testData := fmt.Sprintf("Data from %s", user.Login)
			err = client.PostData(ctx, jwt, testData)
			assert.NoError(t, err, "PostData should succeed for %s", user.Login)
		})
	}
}

func TestDataVault_PostData_JWTCrossValidation(t *testing.T) {
	t.Parallel()

	// Test that JWTs generated with different secrets don't work
	jwtSecret1 := "secret1"
	jwtSecret2 := "secret2"

	// Setup server with secret1
	_, lis1, cleanup1 := SetupMockServerWithJWT(true, "", true, jwtSecret1)
	defer cleanup1()
	client1 := SetupTestClient(t, lis1)

	// Setup server with secret2
	_, lis2, cleanup2 := SetupMockServerWithJWT(true, "", true, jwtSecret2)
	defer cleanup2()
	client2 := SetupTestClient(t, lis2)

	ctx := context.Background()
	user := models.User{Login: "testuser", Password: "testpass"}

	// Register user with server1 and get JWT
	jwt1, err := client1.Register(ctx, user)
	assert.NoError(t, err, "Registration should succeed with server1")

	// Try to use JWT from server1 with server2 (should fail)
	testData := "Cross-validation test data"
	err = client2.PostData(ctx, jwt1, testData)
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
			data:          string(make([]byte, 10000)), // Large data content
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
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			client, ctx, cleanup := tt.setupFunc(t)
			defer cleanup()

			// Execute the post data function
			err := client.PostData(ctx, tt.jwt, tt.data)

			// Validate the results
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

	// Test with cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	err := client.PostData(ctx, expectedToken, "test data")

	// Should get an error due to cancelled context
	assert.Error(t, err)
}

func TestDataVault_PostData_ContextTimeout(t *testing.T) {
	t.Parallel()

	expectedToken := "timeout-test-token"
	_, lis, cleanup := SetupMockServer(true, expectedToken)
	defer cleanup()

	client := SetupTestClient(t, lis)

	// Test with very short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	// Wait a bit to ensure timeout
	time.Sleep(1 * time.Millisecond)

	err := client.PostData(ctx, expectedToken, "test data")

	// Should get an error due to timeout
	assert.Error(t, err)
}

func TestDataVault_PostData_MultipleConsecutive(t *testing.T) {
	t.Parallel()

	expectedToken := "multi-post-token"
	_, lis, cleanup := SetupMockServer(true, expectedToken)
	defer cleanup()

	client := SetupTestClient(t, lis)
	ctx := context.Background()

	// Test multiple consecutive posts
	for i := 0; i < 3; i++ {
		data := fmt.Sprintf("test data item %d", i+1)
		err := client.PostData(ctx, expectedToken, data)
		assert.NoError(t, err, "Expected post data %d to succeed", i+1)
	}
}
