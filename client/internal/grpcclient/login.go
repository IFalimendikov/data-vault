package grpcclient

import (
	"context"
	"data-vault/client/internal/models"
	"data-vault/client/internal/proto"
)

// Login authenticates a user via gRPC and returns a JWT token
func (c *Client) Login(ctx context.Context, user models.User) (string, error) {
	req := &proto.LoginRequest{
		User: &proto.User{
			Login:    user.Login,
			Password: user.Password,
		},
	}

	if user.Login == "" || user.Password == "" {
		return "", ErrorLogin
	}

	grpcResp, err := c.ClientConn.Login(ctx, req)
	if err != nil || !grpcResp.Success {
		return "", ErrorLogin
	}

	jwt := grpcResp.JwtToken

	return jwt, nil
}
