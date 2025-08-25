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

// DeleteData implements the mock DeleteData method
func (m *MockVaultServer) DeleteData(ctx context.Context, req *proto.DeleteDataRequest) (*proto.DeleteDataResponse, error) {
	fmt.Printf("DEBUG MockServer: DeleteData called with ID: %s, shouldSucceed: %t, validateJWT: %t\n", req.Id, m.shouldSucceed, m.validateJWT)

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
		return &proto.DeleteDataResponse{
			Success: false,
		}, nil
	}

	// Check if ID is empty
	if req.Id == "" {
		fmt.Printf("DEBUG MockServer: Empty ID provided\n")
		return nil, status.Error(codes.InvalidArgument, "Data ID not provided")
	}

	// Success case
	fmt.Printf("DEBUG MockServer: DeleteData successful for ID: %s\n", req.Id)
	return &proto.DeleteDataResponse{
		Success: true,
	}, nil
}

func TestDataVault_DeleteData_WithJWTIntegration(t *testing.T) {
	t.Parallel()

	jwtSecret := "test-secret-for-delete-data"
	_, lis, cleanup := SetupMockServerWithJWT(true, "", true, jwtSecret)
	defer cleanup()

	client := SetupTestClient(t, lis)
	ctx := context.Background()

	// First, register a user to get a real JWT token
	user := models.User{
		Login:    "deletedatauser",
		Password: "securepassword123",
	}

	jwt, err := client.Register(ctx, user)
	assert.NoError(t, err, "Registration should succeed")
	assert.NotEmpty(t, jwt, "JWT should not be empty")

	// Now test deleting data with the real JWT token
	dataID := "test-data-id-123"
	err = client.DeleteData(ctx, jwt, dataID)
	assert.NoError(t, err, "DeleteData should succeed with valid JWT")
}

func TestDataVault_DeleteData_WithInvalidJWT(t *testing.T) {
	t.Parallel()

	jwtSecret := "test-secret-for-delete-data"
	_, lis, cleanup := SetupMockServerWithJWT(true, "", true, jwtSecret)
	defer cleanup()

	client := SetupTestClient(t, lis)
	ctx := context.Background()

	// Test with invalid JWT
	invalidJWT := "invalid.jwt.token"
	dataID := "test-data-id-123"

	err := client.DeleteData(ctx, invalidJWT, dataID)
	assert.Error(t, err, "DeleteData should fail with invalid JWT")
}

func TestDataVault_DeleteData_WithoutJWT(t *testing.T) {
	t.Parallel()

	jwtSecret := "test-secret-for-delete-data"
	_, lis, cleanup := SetupMockServerWithJWT(true, "", true, jwtSecret)
	defer cleanup()

	client := SetupTestClient(t, lis)
	ctx := context.Background()

	// Test with empty JWT
	dataID := "test-data-id-123"
	err := client.DeleteData(ctx, "", dataID)
	assert.Error(t, err, "DeleteData should fail with empty JWT")
}

func TestDataVault_DeleteData(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		setupFunc      func(t *testing.T) (*Client, context.Context, func())
		jwt            string
		id             string
		expectedError  bool
		validateResult func(t *testing.T, err error)
	}{
		{
			name: "successful delete data with valid JWT and ID",
			setupFunc: func(t *testing.T) (*Client, context.Context, func()) {
				expectedToken := "valid-jwt-token-123"
				_, lis, cleanup := SetupMockServer(true, expectedToken)
				client := SetupTestClient(t, lis)
				return client, context.Background(), cleanup
			},
			jwt:           "valid-jwt-token-123",
			id:            "data-123",
			expectedError: false,
			validateResult: func(t *testing.T, err error) {
				assert.NoError(t, err, "Expected delete data to succeed")
			},
		},
		{
			name: "failed delete data with server failure",
			setupFunc: func(t *testing.T) (*Client, context.Context, func()) {
				_, lis, cleanup := SetupMockServer(false, "")
				client := SetupTestClient(t, lis)
				return client, context.Background(), cleanup
			},
			jwt:           "any-jwt-token",
			id:            "data-456",
			expectedError: true,
			validateResult: func(t *testing.T, err error) {
				assert.Error(t, err, "Expected delete data to fail")
				assert.Equal(t, ErrorDelete, err, "Expected specific delete error")
			},
		},
		{
			name: "failed delete data with empty ID",
			setupFunc: func(t *testing.T) (*Client, context.Context, func()) {
				expectedToken := "valid-jwt-token-123"
				_, lis, cleanup := SetupMockServer(true, expectedToken)
				client := SetupTestClient(t, lis)
				return client, context.Background(), cleanup
			},
			jwt:           "valid-jwt-token-123",
			id:            "",
			expectedError: true,
			validateResult: func(t *testing.T, err error) {
				assert.Error(t, err, "Expected delete data to fail with empty ID")
			},
		},
		{
			name: "failed delete data with empty JWT",
			setupFunc: func(t *testing.T) (*Client, context.Context, func()) {
				expectedToken := "valid-jwt-token-123"
				_, lis, cleanup := SetupMockServer(true, expectedToken)
				client := SetupTestClient(t, lis)
				return client, context.Background(), cleanup
			},
			jwt:           "",
			id:            "data-789",
			expectedError: true,
			validateResult: func(t *testing.T, err error) {
				assert.Error(t, err, "Expected delete data to fail with empty JWT")
			},
		},
		{
			name: "delete data with UUID format ID",
			setupFunc: func(t *testing.T) (*Client, context.Context, func()) {
				expectedToken := "uuid-test-token"
				_, lis, cleanup := SetupMockServer(true, expectedToken)
				client := SetupTestClient(t, lis)
				return client, context.Background(), cleanup
			},
			jwt:           "uuid-test-token",
			id:            "550e8400-e29b-41d4-a716-446655440000",
			expectedError: false,
			validateResult: func(t *testing.T, err error) {
				assert.NoError(t, err, "Expected delete data with UUID format to succeed")
			},
		},
		{
			name: "delete data with numeric ID",
			setupFunc: func(t *testing.T) (*Client, context.Context, func()) {
				expectedToken := "numeric-test-token"
				_, lis, cleanup := SetupMockServer(true, expectedToken)
				client := SetupTestClient(t, lis)
				return client, context.Background(), cleanup
			},
			jwt:           "numeric-test-token",
			id:            "12345",
			expectedError: false,
			validateResult: func(t *testing.T, err error) {
				assert.NoError(t, err, "Expected delete data with numeric ID to succeed")
			},
		},
		{
			name: "delete data with special characters in ID",
			setupFunc: func(t *testing.T) (*Client, context.Context, func()) {
				expectedToken := "special-chars-token"
				_, lis, cleanup := SetupMockServer(true, expectedToken)
				client := SetupTestClient(t, lis)
				return client, context.Background(), cleanup
			},
			jwt:           "special-chars-token",
			id:            "data-item_2024@test.com",
			expectedError: false,
			validateResult: func(t *testing.T, err error) {
				assert.NoError(t, err, "Expected delete data with special characters to succeed")
			},
		},
	}

	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			client, ctx, cleanup := tt.setupFunc(t)
			defer cleanup()

			// Execute the delete data function
			err := client.DeleteData(ctx, tt.jwt, tt.id)

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

func TestDataVault_DeleteData_ContextCancellation(t *testing.T) {
	t.Parallel()

	expectedToken := "context-test-token"
	_, lis, cleanup := SetupMockServer(true, expectedToken)
	defer cleanup()

	client := SetupTestClient(t, lis)

	// Test with cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	err := client.DeleteData(ctx, expectedToken, "test-data-id")

	// Should get an error due to cancelled context
	assert.Error(t, err)
}

func TestDataVault_DeleteData_ContextTimeout(t *testing.T) {
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

	err := client.DeleteData(ctx, expectedToken, "test-data-id")

	// Should get an error due to timeout
	assert.Error(t, err)
}

func TestDataVault_DeleteData_MultipleConsecutive(t *testing.T) {
	t.Parallel()

	expectedToken := "multi-delete-token"
	_, lis, cleanup := SetupMockServer(true, expectedToken)
	defer cleanup()

	client := SetupTestClient(t, lis)
	ctx := context.Background()

	// Test multiple consecutive deletes
	for i := 0; i < 3; i++ {
		id := fmt.Sprintf("test-data-id-%d", i+1)
		err := client.DeleteData(ctx, expectedToken, id)
		assert.NoError(t, err, "Expected delete data %d to succeed", i+1)
	}
}

func TestDataVault_DeleteData_NonExistentID(t *testing.T) {
	t.Parallel()

	expectedToken := "nonexistent-test-token"
	_, lis, cleanup := SetupMockServer(true, expectedToken)
	defer cleanup()

	client := SetupTestClient(t, lis)
	ctx := context.Background()

	// In mock server, all valid IDs succeed
	// In real implementation, this might return a specific error for non-existent IDs
	err := client.DeleteData(ctx, expectedToken, "non-existent-data-id")
	assert.NoError(t, err, "Mock server allows deletion of non-existent ID")
}

func TestDataVault_DeleteData_LongID(t *testing.T) {
	t.Parallel()

	expectedToken := "long-id-test-token"
	_, lis, cleanup := SetupMockServer(true, expectedToken)
	defer cleanup()

	client := SetupTestClient(t, lis)
	ctx := context.Background()

	// Test with very long ID
	longID := string(make([]byte, 1000))
	for i := range longID {
		longID = longID[:i] + "a" + longID[i+1:]
	}

	err := client.DeleteData(ctx, expectedToken, longID)
	assert.NoError(t, err, "Expected delete data with long ID to succeed")
}
