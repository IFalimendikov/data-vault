package handler

import (
	"context"
	"data-vault/server/internal/proto"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (g *Handler) DeleteData(ctx context.Context, in *proto.DeleteDataRequest) (*proto.DeleteDataResponse, error) {
	var response *proto.DeleteDataResponse
	userID, ok := ctx.Value(userIDKey).(string)
	if !ok || userID == "" {
		return nil, status.Error(codes.Unauthenticated, "User ID not found in context")
	}

	dataID := in.Id
	if len(dataID) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Data ID not provided")
	}

	err := g.service.DeleteData(ctx, userID, dataID)
	if err != nil {
		return nil, status.Error(codes.Internal, "Failed to delete data")
	}

	response = &proto.DeleteDataResponse{
		Success: true,
	}

	return response, nil
}
 