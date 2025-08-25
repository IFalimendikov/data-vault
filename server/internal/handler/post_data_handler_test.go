package handler

import (
	"context"
	"errors"
	"testing"

	"data-vault/server/internal/proto"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestPostData_Success(t *testing.T) {
	handler, mockService := setupTestHandler()

	request := &proto.PostDataRequest{
		Data: "sensitive data to store",
	}
	ctx := context.WithValue(context.Background(), userIDKey, "testuser")

	mockService.On("PostData", mock.Anything, "testuser", "sensitive data to store").Return(nil)

	response, err := handler.PostData(ctx, request)

	require.NoError(t, err)
	require.NotNil(t, response)
	assert.True(t, response.Success)

	mockService.AssertExpectations(t)
}

func TestPostData_EmptyData(t *testing.T) {
	handler, mockService := setupTestHandler()

	request := &proto.PostDataRequest{
		Data: "",
	}
	ctx := context.WithValue(context.Background(), userIDKey, "testuser")

	response, err := handler.PostData(ctx, request)

	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.InvalidArgument, st.Code())
	assert.Contains(t, st.Message(), "Data not provided")
	assert.Nil(t, response)

	mockService.AssertExpectations(t)
}

func TestPostData_MissingUserIDInContext(t *testing.T) {
	handler, mockService := setupTestHandler()

	request := &proto.PostDataRequest{
		Data: "sensitive data to store",
	}
	ctx := context.Background()

	response, err := handler.PostData(ctx, request)

	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.Unauthenticated, st.Code())
	assert.Contains(t, st.Message(), "User ID not found in context")
	assert.Nil(t, response)

	mockService.AssertExpectations(t)
}

func TestPostData_EmptyUserIDInContext(t *testing.T) {
	handler, mockService := setupTestHandler()

	request := &proto.PostDataRequest{
		Data: "sensitive data to store",
	}
	ctx := context.WithValue(context.Background(), userIDKey, "")

	response, err := handler.PostData(ctx, request)

	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.Unauthenticated, st.Code())
	assert.Contains(t, st.Message(), "User ID not found in context")
	assert.Nil(t, response)

	mockService.AssertExpectations(t)
}

func TestPostData_WrongUserIDType(t *testing.T) {
	handler, mockService := setupTestHandler()

	request := &proto.PostDataRequest{
		Data: "sensitive data to store",
	}
	ctx := context.WithValue(context.Background(), userIDKey, 123) // Wrong type

	response, err := handler.PostData(ctx, request)

	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.Unauthenticated, st.Code())
	assert.Contains(t, st.Message(), "User ID not found in context")
	assert.Nil(t, response)

	mockService.AssertExpectations(t)
}

func TestPostData_ServiceError(t *testing.T) {
	handler, mockService := setupTestHandler()

	request := &proto.PostDataRequest{
		Data: "sensitive data to store",
	}
	ctx := context.WithValue(context.Background(), userIDKey, "testuser")

	mockService.On("PostData", mock.Anything, "testuser", "sensitive data to store").Return(errors.New("storage error"))

	response, err := handler.PostData(ctx, request)

	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.Internal, st.Code())
	assert.Contains(t, st.Message(), "Failed to post data")
	assert.Nil(t, response)

	mockService.AssertExpectations(t)
}

func TestPostData_LargeData(t *testing.T) {
	handler, mockService := setupTestHandler()

	largeData := string(make([]byte, 10000)) // 10KB of data
	request := &proto.PostDataRequest{
		Data: largeData,
	}
	ctx := context.WithValue(context.Background(), userIDKey, "testuser")

	mockService.On("PostData", mock.Anything, "testuser", largeData).Return(nil)

	response, err := handler.PostData(ctx, request)

	require.NoError(t, err)
	require.NotNil(t, response)
	assert.True(t, response.Success)

	mockService.AssertExpectations(t)
}
