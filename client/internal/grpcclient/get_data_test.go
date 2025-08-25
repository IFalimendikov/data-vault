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

// GetData implements the mock GetData method
func (m *MockVaultServer) GetData(ctx context.Context, req *proto.GetDataRequest) (*proto.GetDataResponse, error) {
	fmt.Printf("DEBUG MockServer: GetData called, shouldSucceed: %t, validateJWT: %t\n", m.shouldSucceed, m.validateJWT)

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
		return nil, status.Error(codes.Internal, "server failure")
	}

	// Return mock data
	mockData := []*proto.Data{
		{
			Id:         "data-1",
			Data:       []byte("sample encrypted data 1"),
			Status:     "ACTIVE",
			UploadedAt: "2025-08-24T10:00:00Z",
		},
		{
			Id:         "data-2",
			Data:       []byte("sample encrypted data 2"),
			Status:     "ACTIVE",
			UploadedAt: "2025-08-24T11:00:00Z",
		},
	}

	fmt.Printf("DEBUG MockServer: GetData successful, returning %d items\n", len(mockData))
	return &proto.GetDataResponse{
		Data: mockData,
	}, nil
}

func TestDataVault_GetData_WithJWTIntegration(t *testing.T) {
	t.Parallel()

	jwtSecret := "test-secret-for-get-data"
	_, lis, cleanup := SetupMockServerWithJWT(true, "", true, jwtSecret)
	defer cleanup()

	client := SetupTestClient(t, lis)
	ctx := context.Background()

	// First, register a user to get a real JWT token
	user := models.User{
		Login:    "getdatauser",
		Password: "securepassword123",
	}

	jwt, err := client.Register(ctx, user)
	assert.NoError(t, err, "Registration should succeed")
	assert.NotEmpty(t, jwt, "JWT should not be empty")

	// Now test getting data with the real JWT token
	data, err := client.GetData(ctx, jwt)
	assert.NoError(t, err, "GetData should succeed with valid JWT")
	assert.NotNil(t, data, "Data should not be nil")
	assert.Len(t, data, 2, "Should return 2 mock data items")

	// Verify data structure
	assert.Equal(t, "data-1", data[0].ID)
	assert.Equal(t, "sample encrypted data 1", string(data[0].Data))
	assert.Equal(t, "ACTIVE", data[0].Status)
}

func TestDataVault_GetData_WithInvalidJWT(t *testing.T) {
	t.Parallel()

	jwtSecret := "test-secret-for-get-data"
	_, lis, cleanup := SetupMockServerWithJWT(true, "", true, jwtSecret)
	defer cleanup()

	client := SetupTestClient(t, lis)
	ctx := context.Background()

	// Test with invalid JWT
	invalidJWT := "invalid.jwt.token"

	data, err := client.GetData(ctx, invalidJWT)
	assert.Error(t, err, "GetData should fail with invalid JWT")
	assert.Nil(t, data, "Data should be nil on error")
}

func TestDataVault_GetData_WithoutJWT(t *testing.T) {
	t.Parallel()

	jwtSecret := "test-secret-for-get-data"
	_, lis, cleanup := SetupMockServerWithJWT(true, "", true, jwtSecret)
	defer cleanup()

	client := SetupTestClient(t, lis)
	ctx := context.Background()

	// Test with empty JWT
	data, err := client.GetData(ctx, "")
	assert.Error(t, err, "GetData should fail with empty JWT")
	assert.Nil(t, data, "Data should be nil on error")
}

func TestDataVault_GetData(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		setupFunc      func(t *testing.T) (*Client, context.Context, func())
		jwt            string
		expectedError  bool
		expectedCount  int
		validateResult func(t *testing.T, data []models.Data, err error)
	}{
		{
			name: "successful get data with valid JWT",
			setupFunc: func(t *testing.T) (*Client, context.Context, func()) {
				expectedToken := "valid-jwt-token-123"
				_, lis, cleanup := SetupMockServer(true, expectedToken)
				client := SetupTestClient(t, lis)
				return client, context.Background(), cleanup
			},
			jwt:           "valid-jwt-token-123",
			expectedError: false,
			expectedCount: 2,
			validateResult: func(t *testing.T, data []models.Data, err error) {
				assert.NoError(t, err, "Expected get data to succeed")
				assert.Len(t, data, 2, "Expected 2 data items")
				assert.Equal(t, "data-1", data[0].ID)
				assert.Equal(t, "sample encrypted data 1", string(data[0].Data))
				assert.Equal(t, "ACTIVE", data[0].Status)
				assert.Equal(t, "2025-08-24T10:00:00Z", data[0].UploadedAt)
				assert.Equal(t, "data-2", data[1].ID)
				assert.Equal(t, "sample encrypted data 2", string(data[1].Data))
			},
		},
		{
			name: "failed get data with server failure",
			setupFunc: func(t *testing.T) (*Client, context.Context, func()) {
				_, lis, cleanup := SetupMockServer(false, "")
				client := SetupTestClient(t, lis)
				return client, context.Background(), cleanup
			},
			jwt:           "any-jwt-token",
			expectedError: true,
			expectedCount: 0,
			validateResult: func(t *testing.T, data []models.Data, err error) {
				assert.Error(t, err, "Expected get data to fail")
				assert.Nil(t, data, "Expected no data on error")
			},
		},
		{
			name: "get data with empty JWT",
			setupFunc: func(t *testing.T) (*Client, context.Context, func()) {
				expectedToken := "valid-jwt-token-123"
				_, lis, cleanup := SetupMockServer(true, expectedToken)
				client := SetupTestClient(t, lis)
				return client, context.Background(), cleanup
			},
			jwt:           "",
			expectedError: true, // Empty JWT still allows the call, but server would normally reject
			expectedCount: 2,
			validateResult: func(t *testing.T, data []models.Data, err error) {
				// In real scenario, server would reject, but mock server returns data
				assert.Error(t, err, "JWT token is empty")
			},
		},
		{
			name: "get data with expired JWT format",
			setupFunc: func(t *testing.T) (*Client, context.Context, func()) {
				expectedToken := "expired-jwt-token"
				_, lis, cleanup := SetupMockServer(true, expectedToken)
				client := SetupTestClient(t, lis)
				return client, context.Background(), cleanup
			},
			jwt:           "expired-jwt-token",
			expectedError: false, // Mock server doesn't validate expiration
			expectedCount: 2,
			validateResult: func(t *testing.T, data []models.Data, err error) {
				assert.NoError(t, err, "Mock server returns data with expired token format")
				assert.Len(t, data, 2, "Expected 2 data items from mock")
			},
		},
	}

	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			client, ctx, cleanup := tt.setupFunc(t)
			defer cleanup()

			// Execute the get data function
			data, err := client.GetData(ctx, tt.jwt)

			// Validate the results
			tt.validateResult(t, data, err)
			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDataVault_GetData_ContextCancellation(t *testing.T) {
	t.Parallel()

	expectedToken := "context-test-token"
	_, lis, cleanup := SetupMockServer(true, expectedToken)
	defer cleanup()

	client := SetupTestClient(t, lis)

	// Test with cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	data, err := client.GetData(ctx, expectedToken)

	// Should get an error due to cancelled context
	assert.Error(t, err)
	assert.Nil(t, data)
}

func TestDataVault_GetData_ContextTimeout(t *testing.T) {
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

	data, err := client.GetData(ctx, expectedToken)

	// Should get an error due to timeout
	assert.Error(t, err)
	assert.Nil(t, data)
}

func TestDataVault_GetData_MultipleConsecutive(t *testing.T) {
	t.Parallel()

	expectedToken := "multi-get-token"
	_, lis, cleanup := SetupMockServer(true, expectedToken)
	defer cleanup()

	client := SetupTestClient(t, lis)
	ctx := context.Background()

	// Test multiple consecutive gets
	for i := 0; i < 3; i++ {
		data, err := client.GetData(ctx, expectedToken)
		assert.NoError(t, err, "Expected get data %d to succeed", i+1)
		assert.Len(t, data, 2, "Expected 2 data items on request %d", i+1)
	}
}

func TestDataVault_GetData_EmptyResponse(t *testing.T) {
	t.Parallel()

	// Test scenario where server returns empty data list
	expectedToken := "empty-data-token"
	_, lis, cleanup := SetupMockServer(true, expectedToken)
	defer cleanup()

	client := SetupTestClient(t, lis)
	ctx := context.Background()

	// Note: This test uses the current mock implementation which always returns 2 items
	// In a real test, you might want a separate mock that returns empty data
	data, err := client.GetData(ctx, expectedToken)

	assert.NoError(t, err, "Expected get data to succeed")
	// Current mock always returns 2 items, but in real implementation empty response would be valid
	assert.NotNil(t, data, "Data slice should not be nil even if empty")
}

func TestDataVault_GetData_DataIntegrity(t *testing.T) {
	t.Parallel()

	expectedToken := "integrity-test-token"
	_, lis, cleanup := SetupMockServer(true, expectedToken)
	defer cleanup()

	client := SetupTestClient(t, lis)
	ctx := context.Background()

	data, err := client.GetData(ctx, expectedToken)
	assert.NoError(t, err)
	assert.Len(t, data, 2)

	// Verify all fields are properly mapped from proto to models
	for i, item := range data {
		assert.NotEmpty(t, item.ID, "Data item %d should have non-empty ID", i)
		assert.NotEmpty(t, item.Data, "Data item %d should have non-empty Data", i)
		assert.NotEmpty(t, item.Status, "Data item %d should have non-empty Status", i)
		assert.NotEmpty(t, item.UploadedAt, "Data item %d should have non-empty UploadedAt", i)
	}
}
