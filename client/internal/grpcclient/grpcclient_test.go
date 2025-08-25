package grpcclient

import (
	"context"
	"data-vault/client/internal/models"
	"data-vault/client/internal/proto"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

// MockClaims represents JWT claims for testing
type MockClaims struct {
	Login     string `json:"login"`
	ExpiresAt int64  `json:"exp"`
	IssuedAt  int64  `json:"iat"`
}

// MockVaultServer implements the VaultServiceServer interface for testing
type MockVaultServer struct {
	proto.UnimplementedVaultServiceServer
	registeredUsers map[string]string
	shouldSucceed   bool
	expectedToken   string
	jwtSecret       string
	validateJWT     bool
}

// GenerateTestJWT creates a mock JWT token for testing
func (m *MockVaultServer) GenerateTestJWT(login string) string {
	if !m.validateJWT {
		return m.expectedToken
	}

	claims := MockClaims{
		Login:     login,
		ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
		IssuedAt:  time.Now().Unix(),
	}

	// Create a simple base64 encoded token for testing
	// Format: header.payload.signature (simplified)
	header := `{"alg":"HS256","typ":"JWT"}`
	payload, _ := json.Marshal(claims)

	headerB64 := base64.RawURLEncoding.EncodeToString([]byte(header))
	payloadB64 := base64.RawURLEncoding.EncodeToString(payload)

	// Simple signature using secret
	signature := m.generateSignature(headerB64 + "." + payloadB64)
	signatureB64 := base64.RawURLEncoding.EncodeToString([]byte(signature))

	return fmt.Sprintf("%s.%s.%s", headerB64, payloadB64, signatureB64)
}

// ValidateTestJWT validates a mock JWT token
func (m *MockVaultServer) ValidateTestJWT(token string) (string, bool) {
	if !m.validateJWT {
		return "test-user", true
	}

	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return "", false
	}

	// Verify signature
	expectedSig := m.generateSignature(parts[0] + "." + parts[1])
	actualSig, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil || string(actualSig) != expectedSig {
		return "", false
	}

	// Decode payload
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return "", false
	}

	var claims MockClaims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return "", false
	}

	// Check expiration
	if time.Now().Unix() > claims.ExpiresAt {
		return "", false
	}

	return claims.Login, true
}

// generateSignature creates a simple signature for testing
func (m *MockVaultServer) generateSignature(data string) string {
	// Simple hash-like signature using secret
	secret := m.jwtSecret
	if secret == "" {
		secret = "test-secret"
	}

	// Create a deterministic "signature" based on data and secret
	combined := data + secret
	hash := 0
	for _, c := range combined {
		hash = hash*31 + int(c)
	}

	return fmt.Sprintf("sig_%x", hash)
}

func SetupMockServer(shouldSucceed bool, expectedToken string) (*grpc.Server, *bufconn.Listener, func()) {
	return SetupMockServerWithJWT(shouldSucceed, expectedToken, false, "")
}

func SetupMockServerWithJWT(shouldSucceed bool, expectedToken string, validateJWT bool, jwtSecret string) (*grpc.Server, *bufconn.Listener, func()) {
	buffer := 1024 * 1024
	lis := bufconn.Listen(buffer)

	baseServer := grpc.NewServer()
	mockServer := &MockVaultServer{
		shouldSucceed:   shouldSucceed,
		expectedToken:   expectedToken,
		registeredUsers: make(map[string]string),
		validateJWT:     validateJWT,
		jwtSecret:       jwtSecret,
	}
	proto.RegisterVaultServiceServer(baseServer, mockServer)

	go func() {
		if err := baseServer.Serve(lis); err != nil {
			// Log error in real scenario
		}
	}()

	// Give the server time to start
	time.Sleep(100 * time.Millisecond)

	cleanup := func() {
		baseServer.Stop()
		lis.Close()
	}

	return baseServer, lis, cleanup
}

// bufDialer returns a dialer function for the bufconn listener
func BufDialer(lis *bufconn.Listener) func(context.Context, string) (net.Conn, error) {
	return func(ctx context.Context, url string) (net.Conn, error) {
		return lis.Dial()
	}
}

// setupTestClient creates a gRPC client for testing
func SetupTestClient(t *testing.T, lis *bufconn.Listener) *Client {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	t.Cleanup(cancel)

	conn, err := grpc.DialContext(
		ctx,
		"bufnet",
		grpc.WithContextDialer(BufDialer(lis)),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(), // ensure connection established before returning
	)
	require.NoError(t, err, "Failed to create gRPC connection")

	grpcClient := proto.NewVaultServiceClient(conn)
	return &Client{
		ClientConn: grpcClient,
	}
}

// TestJWTGeneration tests JWT token generation and validation
func TestJWTGeneration(t *testing.T) {
	mockServer := &MockVaultServer{
		validateJWT: true,
		jwtSecret:   "test-secret-key",
	}

	// Test JWT generation
	token := mockServer.GenerateTestJWT("testuser")
	require.NotEmpty(t, token, "Generated token should not be empty")
	require.Contains(t, token, ".", "JWT should contain dots")

	// Test JWT validation
	login, valid := mockServer.ValidateTestJWT(token)
	require.True(t, valid, "Token should be valid")
	require.Equal(t, "testuser", login, "Login should match")
}

func TestJWTValidation(t *testing.T) {
	mockServer := &MockVaultServer{
		validateJWT: true,
		jwtSecret:   "test-secret-key",
	}

	tests := []struct {
		name        string
		token       string
		expectValid bool
		expectLogin string
	}{
		{
			name:        "valid token",
			token:       mockServer.GenerateTestJWT("validuser"),
			expectValid: true,
			expectLogin: "validuser",
		},
		{
			name:        "invalid format",
			token:       "invalid.token",
			expectValid: false,
			expectLogin: "",
		},
		{
			name:        "empty token",
			token:       "",
			expectValid: false,
			expectLogin: "",
		},
		{
			name:        "malformed token",
			token:       "header.payload.signature.extra",
			expectValid: false,
			expectLogin: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			login, valid := mockServer.ValidateTestJWT(tt.token)
			require.Equal(t, tt.expectValid, valid, "Validation result should match expected")
			require.Equal(t, tt.expectLogin, login, "Login should match expected")
		})
	}
}

func TestJWTWithDifferentSecrets(t *testing.T) {
	server1 := &MockVaultServer{
		validateJWT: true,
		jwtSecret:   "secret1",
	}
	server2 := &MockVaultServer{
		validateJWT: true,
		jwtSecret:   "secret2",
	}

	// Generate token with server1
	token := server1.GenerateTestJWT("testuser")

	// Validate with server1 (should work)
	login1, valid1 := server1.ValidateTestJWT(token)
	require.True(t, valid1, "Token should be valid with original secret")
	require.Equal(t, "testuser", login1)

	// Validate with server2 (should fail due to different secret)
	_, valid2 := server2.ValidateTestJWT(token)
	require.False(t, valid2, "Token should be invalid with different secret")
}

func TestFullWorkflowWithJWTIntegration(t *testing.T) {
	t.Parallel()

	jwtSecret := "test-secret-full-workflow"
	_, lis, cleanup := SetupMockServerWithJWT(true, "", true, jwtSecret)
	defer cleanup()

	client := SetupTestClient(t, lis)
	ctx := context.Background()

	// Step 1: Register a user and get JWT
	user := models.User{
		Login:    "workflowuser",
		Password: "workflowpassword123",
	}

	jwt, err := client.Register(ctx, user)
	require.NoError(t, err, "Registration should succeed")
	require.NotEmpty(t, jwt, "JWT should not be empty")
	require.Contains(t, jwt, ".", "JWT should be in proper format")

	// Step 2: Validate the JWT token
	mockServer := &MockVaultServer{
		validateJWT: true,
		jwtSecret:   jwtSecret,
	}
	login, valid := mockServer.ValidateTestJWT(jwt)
	require.True(t, valid, "JWT should be valid")
	require.Equal(t, user.Login, login, "JWT should contain correct login")

	// Step 3: Use JWT to post data
	testData := "This is test data for the full workflow"
	err = client.PostData(ctx, jwt, testData)
	require.NoError(t, err, "PostData should succeed with valid JWT")

	// Step 4: Use JWT to get data
	data, err := client.GetData(ctx, jwt)
	require.NoError(t, err, "GetData should succeed with valid JWT")
	require.NotNil(t, data, "Data should not be nil")
	require.Len(t, data, 2, "Should return mock data")

	// Step 5: Use JWT to delete data
	err = client.DeleteData(ctx, jwt, "test-data-id")
	require.NoError(t, err, "DeleteData should succeed with valid JWT")

	// Step 6: Test ping server (doesn't require JWT)
	pingResult := client.PingServer(ctx)
	require.True(t, pingResult, "PingServer should succeed")
}

func TestJWTSecurityValidation(t *testing.T) {
	t.Parallel()

	jwtSecret := "security-test-secret"
	_, lis, cleanup := SetupMockServerWithJWT(true, "", true, jwtSecret)
	defer cleanup()

	client := SetupTestClient(t, lis)
	ctx := context.Background()

	// Register a legitimate user
	legitimateUser := models.User{
		Login:    "legitimate",
		Password: "password123",
	}

	validJWT, err := client.Register(ctx, legitimateUser)
	require.NoError(t, err, "Registration should succeed")

	testData := "Sensitive test data"

	// Test 1: Valid JWT should work
	err = client.PostData(ctx, validJWT, testData)
	assert.NoError(t, err, "Valid JWT should allow data posting")

	// Test 2: Invalid JWT should be rejected
	invalidJWT := "invalid.jwt.token"
	err = client.PostData(ctx, invalidJWT, testData)
	assert.Error(t, err, "Invalid JWT should be rejected")

	// Test 3: Empty JWT should be rejected
	err = client.PostData(ctx, "", testData)
	assert.Error(t, err, "Empty JWT should be rejected")

	// Test 4: Malformed JWT should be rejected
	malformedJWT := "header.payload" // Missing signature
	err = client.PostData(ctx, malformedJWT, testData)
	assert.Error(t, err, "Malformed JWT should be rejected")

	// Test 5: JWT with wrong signature should be rejected
	wrongSignatureJWT := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJsb2dpbiI6ImF0dGFja2VyIiwiZXhwIjo5OTk5OTk5OTk5LCJpYXQiOjE2MDAwMDAwMDB9.wrong_signature"
	err = client.PostData(ctx, wrongSignatureJWT, testData)
	assert.Error(t, err, "JWT with wrong signature should be rejected")
}
