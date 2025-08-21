package grpcclient

import (
	"context"
	"data-vault/client/internal/proto"
)

func (c *Client) PingServer(ctx context.Context) bool {
	req := &proto.PingServerRequest{}

	grpcResp, err := c.ClientConn.PingServer(ctx, req)
	if err != nil || !grpcResp.Success {
		return false
	}

	return true
}
