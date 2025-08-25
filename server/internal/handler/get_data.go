package handler

import (
	"context"
	"data-vault/server/internal/proto"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// GetData handles data retrieval requests
func (g *Handler) GetData(ctx context.Context, in *proto.GetDataRequest) (*proto.GetDataResponse, error) {
	userID, ok := ctx.Value(userIDKey).(string)
	if !ok || userID == "" {
		return nil, status.Error(codes.Unauthenticated, "User ID not found in context")
	}

	data, err := g.service.GetData(ctx, userID)
	if err != nil {
		return nil, status.Error(codes.Internal, "Failed to get data")
	}

	response := &proto.GetDataResponse{
		Data: make([]*proto.Data, 0, len(data)),
	}

	for _, d := range data {
		response.Data = append(response.Data, &proto.Data{
			Id:         d.ID,
			User:       d.User,
			Status:     d.Status,
			Type:       d.Type,
			Data:       d.Data,
			UploadedAt: d.UploadedAt,
		})
	}

	return response, nil
}
