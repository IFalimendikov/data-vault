package handler

import (
	"context"
	"data-vault/server/internal/proto"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (g *Handler) PostData(ctx context.Context, in *proto.PostDataRequest) (*proto.PostDataResponse, error) {
	var response *proto.PostDataResponse

	data := in.Data
	if len(data) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Data not provided")
	}

	login, ok := ctx.Value(userIDKey).(string)
	if !ok || len(login) == 0 {
		return nil, status.Error(codes.Unauthenticated, "User ID not found in context")
	}

	err := g.service.PostData(ctx, login, data)
	if err != nil {
		return nil, status.Error(codes.Internal, "Failed to post data")
	}

	response = &proto.PostDataResponse{
		Success: true,
	}

	return response, nil
}
