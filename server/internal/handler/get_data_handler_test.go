package handler

import (
	"context"
	"errors"
	"testing"

	"data-vault/server/internal/models"
	"data-vault/server/internal/proto"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestGetData_Success(t *testing.T) {
	handler, mockService := setupTestHandler()

	mockData := []models.Data{
		{
			ID:         "data1",
			User:       "testuser",
			Status:     "active",
			Data:       "encrypted data 1",
			UploadedAt: "2024-01-01T00:00:00Z",
		},
		{
			ID:         "data2",
			User:       "testuser",
			Status:     "active",
			Data:       "encrypted data 2",
			UploadedAt: "2024-01-02T00:00:00Z",
		},
	}

	request := &proto.GetDataRequest{}
	ctx := context.WithValue(context.Background(), userIDKey, "testuser")

	mockService.On("GetData", mock.Anything, "testuser").Return(mockData, nil)

	response, err := handler.GetData(ctx, request)

	require.NoError(t, err)
	require.NotNil(t, response)
	assert.Len(t, response.Data, 2)

	for i, expectedItem := range mockData {
		assert.Equal(t, expectedItem.ID, response.Data[i].Id)
		assert.Equal(t, expectedItem.Data, response.Data[i].Data)
		assert.Equal(t, expectedItem.Status, response.Data[i].Status)
		assert.Equal(t, expectedItem.UploadedAt, response.Data[i].UploadedAt)
	}

	mockService.AssertExpectations(t)
}

func TestGetData_EmptyResult(t *testing.T) {
	handler, mockService := setupTestHandler()

	request := &proto.GetDataRequest{}
	ctx := context.WithValue(context.Background(), userIDKey, "testuser")

	mockService.On("GetData", mock.Anything, "testuser").Return([]models.Data{}, nil)

	response, err := handler.GetData(ctx, request)

	require.NoError(t, err)
	require.NotNil(t, response)
	assert.Len(t, response.Data, 0)

	mockService.AssertExpectations(t)
}

func TestGetData_MissingUserIDInContext(t *testing.T) {
	handler, mockService := setupTestHandler()

	request := &proto.GetDataRequest{}
	ctx := context.Background()

	response, err := handler.GetData(ctx, request)

	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.Unauthenticated, st.Code())
	assert.Contains(t, st.Message(), "User ID not found in context")
	assert.Nil(t, response)

	mockService.AssertExpectations(t)
}

func TestGetData_EmptyUserIDInContext(t *testing.T) {
	handler, mockService := setupTestHandler()

	request := &proto.GetDataRequest{}
	ctx := context.WithValue(context.Background(), userIDKey, "")

	response, err := handler.GetData(ctx, request)

	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.Unauthenticated, st.Code())
	assert.Contains(t, st.Message(), "User ID not found in context")
	assert.Nil(t, response)

	mockService.AssertExpectations(t)
}

func TestGetData_WrongUserIDType(t *testing.T) {
	handler, mockService := setupTestHandler()

	request := &proto.GetDataRequest{}
	ctx := context.WithValue(context.Background(), userIDKey, 123) // Wrong type

	response, err := handler.GetData(ctx, request)

	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.Unauthenticated, st.Code())
	assert.Contains(t, st.Message(), "User ID not found in context")
	assert.Nil(t, response)

	mockService.AssertExpectations(t)
}

func TestGetData_ServiceError(t *testing.T) {
	handler, mockService := setupTestHandler()

	request := &proto.GetDataRequest{}
	ctx := context.WithValue(context.Background(), userIDKey, "testuser")

	mockService.On("GetData", mock.Anything, "testuser").Return([]models.Data{}, errors.New("database error"))

	response, err := handler.GetData(ctx, request)

	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.Internal, st.Code())
	assert.Contains(t, st.Message(), "Failed to get data")
	assert.Nil(t, response)

	mockService.AssertExpectations(t)
}

func TestGetData_NilDataFromService(t *testing.T) {
	handler, mockService := setupTestHandler()

	request := &proto.GetDataRequest{}
	ctx := context.WithValue(context.Background(), userIDKey, "testuser")

	mockService.On("GetData", mock.Anything, "testuser").Return(nil, nil)

	response, err := handler.GetData(ctx, request)

	require.NoError(t, err)
	require.NotNil(t, response)
	assert.Len(t, response.Data, 0)

	mockService.AssertExpectations(t)
}
