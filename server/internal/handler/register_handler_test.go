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

func TestRegister_Success(t *testing.T) {
	handler, mockService := setupTestHandler()

	user := testUserValid()
	request := &proto.RegisterRequest{
		User: &proto.User{
			Login:    user.Login,
			Password: user.Password,
		},
	}

	mockService.On("Register", mock.Anything, user).Return(nil)

	response, err := handler.Register(context.Background(), request)

	require.NoError(t, err)
	require.NotNil(t, response)
	assert.True(t, response.Success)
	assert.NotEmpty(t, response.JwtToken)

	// Validate JWT token using helper function
	validateJWTToken(t, response.JwtToken, handler.cfg.JWTSecret, user.Login)

	mockService.AssertExpectations(t)
}

func TestRegister_EmptyLogin(t *testing.T) {
	handler, mockService := setupTestHandler()

	user := testUserEmptyLogin()
	request := &proto.RegisterRequest{
		User: &proto.User{
			Login:    user.Login,
			Password: user.Password,
		},
	}

	response, err := handler.Register(context.Background(), request)

	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.InvalidArgument, st.Code())
	assert.Contains(t, st.Message(), "User ID or Password not provided")
	assert.Nil(t, response)

	mockService.AssertExpectations(t)
}

func TestRegister_EmptyPassword(t *testing.T) {
	handler, mockService := setupTestHandler()

	user := testUserEmptyPassword()
	request := &proto.RegisterRequest{
		User: &proto.User{
			Login:    user.Login,
			Password: user.Password,
		},
	}

	response, err := handler.Register(context.Background(), request)

	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.InvalidArgument, st.Code())
	assert.Contains(t, st.Message(), "User ID or Password not provided")
	assert.Nil(t, response)

	mockService.AssertExpectations(t)
}

func TestRegister_DuplicateUser(t *testing.T) {
	handler, mockService := setupTestHandler()

	request := &proto.RegisterRequest{
		User: &proto.User{
			Login:    "existinguser",
			Password: "testpass123",
		},
	}

	mockService.On("Register", mock.Anything, models.User{
		Login:    "existinguser",
		Password: "testpass123",
	}).Return(storage.ErrDuplicateLogin)

	response, err := handler.Register(context.Background(), request)

	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.AlreadyExists, st.Code())
	assert.Contains(t, st.Message(), "User already exists")
	assert.Nil(t, response)

	mockService.AssertExpectations(t)
}

func TestRegister_ServiceError(t *testing.T) {
	handler, mockService := setupTestHandler()

	request := &proto.RegisterRequest{
		User: &proto.User{
			Login:    "testuser",
			Password: "testpass123",
		},
	}

	mockService.On("Register", mock.Anything, models.User{
		Login:    "testuser",
		Password: "testpass123",
	}).Return(errors.New("database error"))

	response, err := handler.Register(context.Background(), request)

	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.Internal, st.Code())
	assert.Contains(t, st.Message(), "Failed to register user")
	assert.Nil(t, response)

	mockService.AssertExpectations(t)
}
