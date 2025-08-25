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

func TestLogin_Success(t *testing.T) {
	handler, mockService := setupTestHandler()

	request := &proto.LoginRequest{
		User: &proto.User{
			Login:    "testuser",
			Password: "testpass123",
		},
	}

	mockService.On("Login", mock.Anything, models.User{
		Login:    "testuser",
		Password: "testpass123",
	}).Return(nil)

	response, err := handler.Login(context.Background(), request)

	require.NoError(t, err)
	require.NotNil(t, response)
	assert.True(t, response.Success)
	assert.NotEmpty(t, response.JwtToken)

	// Validate JWT token structure
	token, err := jwt.ParseWithClaims(response.JwtToken, &Claim{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(handler.cfg.JWTSecret), nil
	})
	require.NoError(t, err)
	assert.True(t, token.Valid)

	claims, ok := token.Claims.(*Claim)
	require.True(t, ok)
	assert.Equal(t, request.User.Login, claims.Login)

	mockService.AssertExpectations(t)
}

func TestLogin_EmptyLogin(t *testing.T) {
	handler, mockService := setupTestHandler()

	request := &proto.LoginRequest{
		User: &proto.User{
			Login:    "",
			Password: "testpass123",
		},
	}

	response, err := handler.Login(context.Background(), request)

	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.InvalidArgument, st.Code())
	assert.Contains(t, st.Message(), "User ID or Password not provided")
	assert.Nil(t, response)

	mockService.AssertExpectations(t)
}

func TestLogin_EmptyPassword(t *testing.T) {
	handler, mockService := setupTestHandler()

	request := &proto.LoginRequest{
		User: &proto.User{
			Login:    "testuser",
			Password: "",
		},
	}

	response, err := handler.Login(context.Background(), request)

	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.InvalidArgument, st.Code())
	assert.Contains(t, st.Message(), "User ID or Password not provided")
	assert.Nil(t, response)

	mockService.AssertExpectations(t)
}

func TestLogin_WrongPassword(t *testing.T) {
	handler, mockService := setupTestHandler()

	request := &proto.LoginRequest{
		User: &proto.User{
			Login:    "testuser",
			Password: "wrongpass",
		},
	}

	mockService.On("Login", mock.Anything, models.User{
		Login:    "testuser",
		Password: "wrongpass",
	}).Return(storage.ErrWrongPassword)

	response, err := handler.Login(context.Background(), request)

	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.Unauthenticated, st.Code())
	assert.Contains(t, st.Message(), "Wrong password")
	assert.Nil(t, response)

	mockService.AssertExpectations(t)
}

func TestLogin_ServiceError(t *testing.T) {
	handler, mockService := setupTestHandler()

	request := &proto.LoginRequest{
		User: &proto.User{
			Login:    "testuser",
			Password: "testpass123",
		},
	}

	mockService.On("Login", mock.Anything, models.User{
		Login:    "testuser",
		Password: "testpass123",
	}).Return(errors.New("database error"))

	response, err := handler.Login(context.Background(), request)

	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.Internal, st.Code())
	assert.Contains(t, st.Message(), "Failed to login user")
	assert.Nil(t, response)

	mockService.AssertExpectations(t)
}
