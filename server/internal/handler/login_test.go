package handler

import (
	"context"
	"errors"
	"testing"

	"data-vault/server/internal/models"
	"data-vault/server/internal/proto"
	"data-vault/server/internal/storage"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestLogin(t *testing.T) {
	tests := []struct {
		name          string
		user          models.User
		mockError     error
		expectError   bool
		expectedCode  codes.Code
		expectedMsg   string
		expectSuccess bool
		expectJWT     bool
	}{
		{
			name:          "success",
			user:          models.User{Login: "testuser", Password: "testpass123"},
			mockError:     nil,
			expectError:   false,
			expectSuccess: true,
			expectJWT:     true,
		},
		{
			name:         "empty login",
			user:         models.User{Login: "", Password: "testpass123"},
			expectError:  true,
			expectedCode: codes.InvalidArgument,
			expectedMsg:  "User ID or Password not provided",
		},
		{
			name:         "empty password",
			user:         models.User{Login: "testuser", Password: ""},
			expectError:  true,
			expectedCode: codes.InvalidArgument,
			expectedMsg:  "User ID or Password not provided",
		},
		{
			name:         "wrong password",
			user:         models.User{Login: "testuser", Password: "wrongpass"},
			mockError:    storage.ErrWrongPassword,
			expectError:  true,
			expectedCode: codes.Unauthenticated,
			expectedMsg:  "Wrong password",
		},
		{
			name:         "service error",
			user:         models.User{Login: "testuser", Password: "testpass123"},
			mockError:    errors.New("database error"),
			expectError:  true,
			expectedCode: codes.Internal,
			expectedMsg:  "Failed to login user",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mockService := setupTestHandler()

			request := &proto.LoginRequest{
				User: &proto.User{
					Login:    tt.user.Login,
					Password: tt.user.Password,
				},
			}

			// Only set up mock for valid requests
			if tt.user.Login != "" && tt.user.Password != "" {
				mockService.On("Login", mock.Anything, tt.user).Return(tt.mockError)
			}

			response, err := handler.Login(context.Background(), request)

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
					// Validate JWT token structure
					token, err := jwt.ParseWithClaims(response.JwtToken, &Claim{}, func(token *jwt.Token) (interface{}, error) {
						return []byte(handler.cfg.JWTSecret), nil
					})
					require.NoError(t, err)
					assert.True(t, token.Valid)

					claims, ok := token.Claims.(*Claim)
					require.True(t, ok)
					assert.Equal(t, tt.user.Login, claims.Login)
				}
			}

			mockService.AssertExpectations(t)
		})
	}
}
