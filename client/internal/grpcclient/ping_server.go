package grpcclient

import (
	"context"
	"data-vault/client/internal/proto"
)

// PingServer checks server connectivity via gRPC
func (c *Client) PingServer(ctx context.Context) bool {
	req := &proto.PingDBRequest{}

	grpcResp, err := c.ClientConn.PingDB(ctx, req)
	if err != nil || !grpcResp.Success {
		return false
	}

	return true
}
