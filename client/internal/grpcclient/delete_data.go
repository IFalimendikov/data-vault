package grpcclient

import (
	"context"

	"data-vault/client/internal/proto"
	"google.golang.org/grpc/metadata"
)

// Delete marks multiple URLs as deleted in the database for a given user
func (c *Client) DeleteData(ctx context.Context, jwt, id string) error {
	md := metadata.New(map[string]string{
		"authorization": "Bearer " + jwt,
	})
	ctx = metadata.NewOutgoingContext(ctx, md)

	req := &proto.DeleteDataRequest{
		Id: id,
	}

	grpcResp, err := c.ClientConn.DeleteData(ctx, req)
	if err != nil || !grpcResp.Success {
		return ErrorDelete
	}

	return nil
}
