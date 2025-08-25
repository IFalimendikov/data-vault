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

func TestGetData(t *testing.T) {
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

	tests := []struct {
		name         string
		userID       interface{}
		mockData     []models.Data
		mockError    error
		expectError  bool
		expectedCode codes.Code
		expectedMsg  string
		expectedLen  int
	}{
		{
			name:        "success",
			userID:      "testuser",
			mockData:    mockData,
			mockError:   nil,
			expectError: false,
			expectedLen: 2,
		},
		{
			name:        "empty result",
			userID:      "testuser",
			mockData:    []models.Data{},
			mockError:   nil,
			expectError: false,
			expectedLen: 0,
		},
		{
			name:         "missing user ID in context",
			userID:       nil,
			expectError:  true,
			expectedCode: codes.Unauthenticated,
			expectedMsg:  "User ID not found in context",
		},
		{
			name:         "empty user ID in context",
			userID:       "",
			expectError:  true,
			expectedCode: codes.Unauthenticated,
			expectedMsg:  "User ID not found in context",
		},
		{
			name:         "wrong user ID type",
			userID:       123,
			expectError:  true,
			expectedCode: codes.Unauthenticated,
			expectedMsg:  "User ID not found in context",
		},
		{
			name:         "service error",
			userID:       "testuser",
			mockData:     []models.Data{},
			mockError:    errors.New("database error"),
			expectError:  true,
			expectedCode: codes.Internal,
			expectedMsg:  "Failed to get data",
		},
		{
			name:        "nil data from service",
			userID:      "testuser",
			mockData:    nil,
			mockError:   nil,
			expectError: false,
			expectedLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mockService := setupTestHandler()

			request := &proto.GetDataRequest{}

			var ctx context.Context
			if tt.userID != nil {
				ctx = context.WithValue(context.Background(), userIDKey, tt.userID)
			} else {
				ctx = context.Background()
			}

			// Only set up mock for valid requests
			if tt.userID == "testuser" {
				mockService.On("GetData", mock.Anything, "testuser").Return(tt.mockData, tt.mockError)
			}

			response, err := handler.GetData(ctx, request)

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
				assert.Len(t, response.Data, tt.expectedLen)

				if tt.expectedLen > 0 && tt.mockData != nil {
					for i, expectedItem := range tt.mockData {
						assert.Equal(t, expectedItem.ID, response.Data[i].Id)
						assert.Equal(t, expectedItem.Data, response.Data[i].Data)
						assert.Equal(t, expectedItem.Status, response.Data[i].Status)
						assert.Equal(t, expectedItem.UploadedAt, response.Data[i].UploadedAt)
					}
				}
			}

			mockService.AssertExpectations(t)
		})
	}
}
