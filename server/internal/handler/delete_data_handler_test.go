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

func TestDeleteData_Success(t *testing.T) {
	handler, mockService := setupTestHandler()

	request := &proto.DeleteDataRequest{
		Id: "data123",
	}
	ctx := context.WithValue(context.Background(), userIDKey, "testuser")

	mockService.On("DeleteData", mock.Anything, "testuser", "data123").Return(nil)

	response, err := handler.DeleteData(ctx, request)

	require.NoError(t, err)
	require.NotNil(t, response)
	assert.True(t, response.Success)

	mockService.AssertExpectations(t)
}

func TestDeleteData_EmptyDataID(t *testing.T) {
	handler, mockService := setupTestHandler()

	request := &proto.DeleteDataRequest{
		Id: "",
	}
	ctx := context.WithValue(context.Background(), userIDKey, "testuser")

	response, err := handler.DeleteData(ctx, request)

	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.InvalidArgument, st.Code())
	assert.Contains(t, st.Message(), "Data ID not provided")
	assert.Nil(t, response)

	mockService.AssertExpectations(t)
}

func TestDeleteData_MissingUserIDInContext(t *testing.T) {
	handler, mockService := setupTestHandler()

	request := &proto.DeleteDataRequest{
		Id: "data123",
	}
	ctx := context.Background()

	response, err := handler.DeleteData(ctx, request)

	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.Unauthenticated, st.Code())
	assert.Contains(t, st.Message(), "User ID not found in context")
	assert.Nil(t, response)

	mockService.AssertExpectations(t)
}

func TestDeleteData_EmptyUserIDInContext(t *testing.T) {
	handler, mockService := setupTestHandler()

	request := &proto.DeleteDataRequest{
		Id: "data123",
	}
	ctx := context.WithValue(context.Background(), userIDKey, "")

	response, err := handler.DeleteData(ctx, request)

	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.Unauthenticated, st.Code())
	assert.Contains(t, st.Message(), "User ID not found in context")
	assert.Nil(t, response)

	mockService.AssertExpectations(t)
}

func TestDeleteData_WrongUserIDType(t *testing.T) {
	handler, mockService := setupTestHandler()

	request := &proto.DeleteDataRequest{
		Id: "data123",
	}
	ctx := context.WithValue(context.Background(), userIDKey, 123) // Wrong type

	response, err := handler.DeleteData(ctx, request)

	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.Unauthenticated, st.Code())
	assert.Contains(t, st.Message(), "User ID not found in context")
	assert.Nil(t, response)

	mockService.AssertExpectations(t)
}

func TestDeleteData_ServiceError(t *testing.T) {
	handler, mockService := setupTestHandler()

	request := &proto.DeleteDataRequest{
		Id: "data123",
	}
	ctx := context.WithValue(context.Background(), userIDKey, "testuser")

	mockService.On("DeleteData", mock.Anything, "testuser", "data123").Return(errors.New("database error"))

	response, err := handler.DeleteData(ctx, request)

	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.Internal, st.Code())
	assert.Contains(t, st.Message(), "Failed to delete data")
	assert.Nil(t, response)

	mockService.AssertExpectations(t)
}

func TestDeleteData_NonExistentData(t *testing.T) {
	handler, mockService := setupTestHandler()

	request := &proto.DeleteDataRequest{
		Id: "nonexistent123",
	}
	ctx := context.WithValue(context.Background(), userIDKey, "testuser")

	// Service should handle non-existent data gracefully
	mockService.On("DeleteData", mock.Anything, "testuser", "nonexistent123").Return(nil)

	response, err := handler.DeleteData(ctx, request)

	require.NoError(t, err)
	require.NotNil(t, response)
	assert.True(t, response.Success)

	mockService.AssertExpectations(t)
}
