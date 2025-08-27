package grpcclient

import (
	"context"

	"data-vault/client/internal/proto"

	"google.golang.org/grpc/metadata"
)

// DeleteData removes a specific data entry from the vault via gRPC
func (c *Client) DeleteData(ctx context.Context, jwt, id string) error {
	md := metadata.New(map[string]string{
		"authorization": "Bearer " + jwt,
	})
	ctx = metadata.NewOutgoingContext(ctx, md)

	if id == "" || jwt == "" {
		return ErrorDelete
	}

	req := &proto.DeleteDataRequest{
		Id: id,
	}

	grpcResp, err := c.ClientConn.DeleteData(ctx, req)
	if err != nil || !grpcResp.Success {
		return ErrorDelete
	}

	return nil
}
