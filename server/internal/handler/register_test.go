package handler

import (
	"context"
	"errors"
	"testing"

	"data-vault/server/internal/models"
	"data-vault/server/internal/proto"
	"data-vault/server/internal/storage"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestRegister(t *testing.T) {
	tests := []struct {
		name           string
		user           models.User
		mockError      error
		expectError    bool
		expectedCode   codes.Code
		expectedMsg    string
		expectSuccess  bool
		expectJWT      bool
	}{
		{
			name:          "success",
			user:          testUserValid(),
			mockError:     nil,
			expectError:   false,
			expectSuccess: true,
			expectJWT:     true,
		},
		{
			name:         "empty login",
			user:         testUserEmptyLogin(),
			expectError:  true,
			expectedCode: codes.InvalidArgument,
			expectedMsg:  "User ID or Password not provided",
		},
		{
			name:         "empty password",
			user:         testUserEmptyPassword(),
			expectError:  true,
			expectedCode: codes.InvalidArgument,
			expectedMsg:  "User ID or Password not provided",
		},
		{
			name:         "duplicate user",
			user:         models.User{Login: "existinguser", Password: "testpass123"},
			mockError:    storage.ErrDuplicateLogin,
			expectError:  true,
			expectedCode: codes.AlreadyExists,
			expectedMsg:  "User already exists",
		},
		{
			name:         "service error",
			user:         models.User{Login: "testuser", Password: "testpass123"},
			mockError:    errors.New("database error"),
			expectError:  true,
			expectedCode: codes.Internal,
			expectedMsg:  "Failed to register user",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mockService := setupTestHandler()

			request := &proto.RegisterRequest{
				User: &proto.User{
					Login:    tt.user.Login,
					Password: tt.user.Password,
				},
			}

			// Only set up mock for valid requests
			if tt.user.Login != "" && tt.user.Password != "" {
				mockService.On("Register", mock.Anything, tt.user).Return(tt.mockError)
			}

			response, err := handler.Register(context.Background(), request)

			if tt.expectError {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				assert.Equal(t, tt.expectedCode, st.Code())
				assert.Contains(t, st.Message(), tt.expectedMsg)
				assert.Nil(t, response)
			} else {
				require.NoError(t, err)
				require.NotNil(t, response)
				assert.Equal(t, tt.expectSuccess, response.Success)
				if tt.expectJWT {
					assert.NotEmpty(t, response.JwtToken)
					validateJWTToken(t, response.JwtToken, handler.cfg.JWTSecret, tt.user.Login)
				}
			}

			mockService.AssertExpectations(t)
		})
	}
}
