package handler

import (
	"context"
	"data-vault/server/internal/proto"
	"data-vault/server/internal/storage"
	"errors"

	"data-vault/server/internal/models"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Register handles user registration requests
func (g *Handler) Register(ctx context.Context, in *proto.RegisterRequest) (*proto.RegisterResponse, error) {
	var response *proto.RegisterResponse

	user := models.User{
		Login:    in.User.Login,
		Password: in.User.Password,
	}

	if len(user.Login) == 0 || len(user.Password) == 0 {
		return nil, status.Error(codes.InvalidArgument, "User ID or Password not provided")
	}

	err := g.service.Register(ctx, user)
	if err != nil {
		if errors.Is(err, storage.ErrDuplicateLogin) {
			return nil, status.Error(codes.AlreadyExists, "User already exists")
		}
		return nil, status.Error(codes.Internal, "Failed to register user")
	}

	jwtToken, err := g.IssueJWT(user)
	if err != nil {
		return nil, status.Error(codes.Internal, "Failed to issue JWT token")
	}

	response = &proto.RegisterResponse{
		Success:  true,
		JwtToken: jwtToken,
	}

	return response, nil
}
