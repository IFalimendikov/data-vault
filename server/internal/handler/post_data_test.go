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

func TestPostData(t *testing.T) {
	tests := []struct {
		name          string
		data          []byte
		dataType      string
		userID        interface{}
		mockError     error
		expectError   bool
		expectedCode  codes.Code
		expectedMsg   string
		expectSuccess bool
	}{
		{
			name:          "success",
			data:          []byte("sensitive data to store"),
			dataType:      "text",
			userID:        "testuser",
			mockError:     nil,
			expectError:   false,
			expectSuccess: true,
		},
		{
			name:         "empty data",
			data:         []byte{},
			dataType:     "text",
			userID:       "testuser",
			expectError:  true,
			expectedCode: codes.InvalidArgument,
			expectedMsg:  "Data not provided",
		},
		{
			name:         "empty data type",
			data:         []byte("sensitive data to store"),
			dataType:     "",
			userID:       "testuser",
			expectError:  true,
			expectedCode: codes.InvalidArgument,
			expectedMsg:  "Data not provided",
		},
		{
			name:         "missing user ID in context",
			data:         []byte("sensitive data to store"),
			dataType:     "text",
			userID:       nil,
			expectError:  true,
			expectedCode: codes.Unauthenticated,
			expectedMsg:  "User ID not found in context",
		},
		{
			name:         "empty user ID in context",
			data:         []byte("sensitive data to store"),
			dataType:     "text",
			userID:       "",
			expectError:  true,
			expectedCode: codes.Unauthenticated,
			expectedMsg:  "User ID not found in context",
		},
		{
			name:         "wrong user ID type",
			data:         []byte("sensitive data to store"),
			dataType:     "text",
			userID:       123,
			expectError:  true,
			expectedCode: codes.Unauthenticated,
			expectedMsg:  "User ID not found in context",
		},
		{
			name:         "service error",
			data:         []byte("sensitive data to store"),
			dataType:     "text",
			userID:       "testuser",
			mockError:    errors.New("storage error"),
			expectError:  true,
			expectedCode: codes.Internal,
			expectedMsg:  "Failed to post data",
		},
		{
			name:          "large data",
			data:          make([]byte, 10000),
			dataType:      "binary",
			userID:        "testuser",
			mockError:     nil,
			expectError:   false,
			expectSuccess: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mockService := setupTestHandler()

			request := &proto.PostDataRequest{
				Type: tt.dataType,
				Data: tt.data,
			}

			var ctx context.Context
			if tt.userID != nil {
				ctx = context.WithValue(context.Background(), userIDKey, tt.userID)
			} else {
				ctx = context.Background()
			}

			if len(tt.data) > 0 && tt.dataType != "" && tt.userID == "testuser" {
				mockService.On("PostData", mock.Anything, "testuser", tt.dataType, tt.data).Return(tt.mockError)
			}

			response, err := handler.PostData(ctx, request)

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
