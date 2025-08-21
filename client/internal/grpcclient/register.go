package grpcclient

import (
	"context"
	"data-vault/client/internal/models"
	"data-vault/client/internal/proto"
)

func (c *Client) Register(ctx context.Context, user models.User) (string, error) {
	req := &proto.RegisterRequest{
		User: &proto.User{
			Login:    user.Login,
			Password: user.Password,
		},
	}

	grpcResp, err := c.ClientConn.Register(ctx, req)
	if err != nil || !grpcResp.Success {
		return "", ErrorRegister
	}

	jwt := grpcResp.JwtToken

	return jwt, nil
}
