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

func TestDeleteData(t *testing.T) {
	tests := []struct {
		name          string
		dataID        string
		userID        interface{}
		mockError     error
		expectError   bool
		expectedCode  codes.Code
		expectedMsg   string
		expectSuccess bool
	}{
		{
			name:          "success",
			dataID:        "data123",
			userID:        "testuser",
			mockError:     nil,
			expectError:   false,
			expectSuccess: true,
		},
		{
			name:         "empty data ID",
			dataID:       "",
			userID:       "testuser",
			expectError:  true,
			expectedCode: codes.InvalidArgument,
			expectedMsg:  "Data ID not provided",
		},
		{
			name:         "missing user ID in context",
			dataID:       "data123",
			userID:       nil,
			expectError:  true,
			expectedCode: codes.Unauthenticated,
			expectedMsg:  "User ID not found in context",
		},
		{
			name:         "empty user ID in context",
			dataID:       "data123",
			userID:       "",
			expectError:  true,
			expectedCode: codes.Unauthenticated,
			expectedMsg:  "User ID not found in context",
		},
		{
			name:         "wrong user ID type",
			dataID:       "data123",
			userID:       123,
			expectError:  true,
			expectedCode: codes.Unauthenticated,
			expectedMsg:  "User ID not found in context",
		},
		{
			name:         "service error",
			dataID:       "data123",
			userID:       "testuser",
			mockError:    errors.New("database error"),
			expectError:  true,
			expectedCode: codes.Internal,
			expectedMsg:  "Failed to delete data",
		},
		{
			name:          "non-existent data",
			dataID:        "nonexistent123",
			userID:        "testuser",
			mockError:     nil,
			expectError:   false,
			expectSuccess: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mockService := setupTestHandler()

			request := &proto.DeleteDataRequest{
				Id: tt.dataID,
			}

			var ctx context.Context
			if tt.userID != nil {
				ctx = context.WithValue(context.Background(), userIDKey, tt.userID)
			} else {
				ctx = context.Background()
			}

			// Only set up mock for valid requests
			if tt.dataID != "" && tt.userID == "testuser" {
				mockService.On("DeleteData", mock.Anything, "testuser", tt.dataID).Return(tt.mockError)
			}

			response, err := handler.DeleteData(ctx, request)

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
			}

			mockService.AssertExpectations(t)
		})
	}
}
