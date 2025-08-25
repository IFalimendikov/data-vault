package grpcclient

import (
	"context"
	"data-vault/client/internal/proto"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// PingServer implements the mock PingServer method
func (m *MockVaultServer) PingServer(ctx context.Context, req *proto.PingServerRequest) (*proto.PingServerResponse, error) {
	fmt.Printf("DEBUG MockServer: PingServer called, shouldSucceed: %t\n", m.shouldSucceed)

	// If shouldSucceed is false, simulate server failure
	if !m.shouldSucceed {
		fmt.Printf("DEBUG MockServer: shouldSucceed is false, returning failure\n")
		return &proto.PingServerResponse{
			Success: false,
		}, nil
	}

	// Success case
	fmt.Printf("DEBUG MockServer: PingServer successful\n")
	return &proto.PingServerResponse{
		Success: true,
	}, nil
}

func TestDataVault_PingServer(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		setupFunc      func(t *testing.T) (*Client, context.Context, func())
		expectedResult bool
		validateResult func(t *testing.T, result bool)
	}{
		{
			name: "successful ping to healthy server",
			setupFunc: func(t *testing.T) (*Client, context.Context, func()) {
				_, lis, cleanup := SetupMockServer(true, "")
				client := SetupTestClient(t, lis)
				return client, context.Background(), cleanup
			},
			expectedResult: true,
			validateResult: func(t *testing.T, result bool) {
				assert.True(t, result, "Expected ping to succeed")
			},
		},
		{
			name: "failed ping to unhealthy server",
			setupFunc: func(t *testing.T) (*Client, context.Context, func()) {
				_, lis, cleanup := SetupMockServer(false, "")
				client := SetupTestClient(t, lis)
				return client, context.Background(), cleanup
			},
			expectedResult: false,
			validateResult: func(t *testing.T, result bool) {
				assert.False(t, result, "Expected ping to fail")
			},
		},
		{
			name: "ping with cancelled context",
			setupFunc: func(t *testing.T) (*Client, context.Context, func()) {
				_, lis, cleanup := SetupMockServer(true, "")
				client := SetupTestClient(t, lis)
				ctx, cancel := context.WithCancel(context.Background())
				cancel() // Cancel immediately
				return client, ctx, cleanup
			},
			expectedResult: false,
			validateResult: func(t *testing.T, result bool) {
				assert.False(t, result, "Expected ping to fail with cancelled context")
			},
		},
		{
			name: "ping with timeout context",
			setupFunc: func(t *testing.T) (*Client, context.Context, func()) {
				_, lis, cleanup := SetupMockServer(true, "")
				client := SetupTestClient(t, lis)
				ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
				t.Cleanup(cancel)
				// Wait a bit to ensure timeout
				time.Sleep(1 * time.Millisecond)
				return client, ctx, cleanup
			},
			expectedResult: false,
			validateResult: func(t *testing.T, result bool) {
				assert.False(t, result, "Expected ping to fail with timeout context")
			},
		},
		{
			name: "multiple consecutive pings to healthy server",
			setupFunc: func(t *testing.T) (*Client, context.Context, func()) {
				_, lis, cleanup := SetupMockServer(true, "")
				client := SetupTestClient(t, lis)
				return client, context.Background(), cleanup
			},
			expectedResult: true,
			validateResult: func(t *testing.T, result bool) {
				assert.True(t, result, "Expected first ping to succeed")
			},
		},
	}

	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			client, ctx, cleanup := tt.setupFunc(t)
			defer cleanup()

			// Execute the ping function
			result := client.PingServer(ctx)

			// Validate the results
			tt.validateResult(t, result)
			assert.Equal(t, tt.expectedResult, result)
		})
	}
}

func TestDataVault_PingServer_MultipleConsecutive(t *testing.T) {
	t.Parallel()

	_, lis, cleanup := SetupMockServer(true, "")
	defer cleanup()

	client := SetupTestClient(t, lis)
	ctx := context.Background()

	// Test multiple consecutive pings
	for i := 0; i < 5; i++ {
		result := client.PingServer(ctx)
		assert.True(t, result, "Expected ping %d to succeed", i+1)
	}
}

func TestDataVault_PingServer_AlternatingServerStates(t *testing.T) {
	t.Parallel()

	// Test with healthy server
	_, lis1, cleanup1 := SetupMockServer(true, "")
	defer cleanup1()
	client1 := SetupTestClient(t, lis1)

	result1 := client1.PingServer(context.Background())
	assert.True(t, result1, "Expected ping to healthy server to succeed")

	// Test with unhealthy server
	_, lis2, cleanup2 := SetupMockServer(false, "")
	defer cleanup2()
	client2 := SetupTestClient(t, lis2)

	result2 := client2.PingServer(context.Background())
	assert.False(t, result2, "Expected ping to unhealthy server to fail")
}

func TestDataVault_PingServer_ContextWithDeadline(t *testing.T) {
	t.Parallel()

	_, lis, cleanup := SetupMockServer(true, "")
	defer cleanup()

	client := SetupTestClient(t, lis)

	// Test with a reasonable deadline
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(5*time.Second))
	defer cancel()

	result := client.PingServer(ctx)
	assert.True(t, result, "Expected ping with reasonable deadline to succeed")
}

func TestDataVault_PingServer_EmptyContext(t *testing.T) {
	t.Parallel()

	_, lis, cleanup := SetupMockServer(true, "")
	defer cleanup()

	client := SetupTestClient(t, lis)

	// Test with empty context (should still work)
	result := client.PingServer(context.Background())
	assert.True(t, result, "Expected ping with background context to succeed")
}
