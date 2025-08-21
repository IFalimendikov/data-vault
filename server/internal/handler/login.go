package handler

import (
	"context"
	"data-vault/server/internal/models"
	"data-vault/server/internal/proto"
	"data-vault/server/internal/storage"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (g *Handler) Login(ctx context.Context, in *proto.LoginRequest) (*proto.LoginResponse, error) {
	var response *proto.LoginResponse

	user := models.User{
		Login:    in.User.Login,
		Password: in.User.Password,
	}

	if len(user.Login) == 0 || len(user.Password) == 0 {
		return nil, status.Error(codes.InvalidArgument, "User ID or Password not provided")
	}

	err := g.service.Login(ctx, user)
	if err != nil {
		if errors.Is(err, storage.ErrWrongPassword) {
			return nil, status.Error(codes.Unauthenticated, "Wrong password")
		}
		return nil, status.Error(codes.Internal, "Failed to login user")
	}

	jwtToken, err := g.IssueJWT(user)
	if err != nil {
		return nil, status.Error(codes.Internal, "Failed to issue JWT token")
	}

	response = &proto.LoginResponse{
		Success:  true,
		JwtToken: jwtToken,
	}

	return response, nil
}
