package grpcclient

import (
	"context"
	"data-vault/client/internal/proto"
	"errors"

	"google.golang.org/grpc/metadata"
)

// PostData stores encrypted data in the vault via gRPC
func (c *Client) PostData(ctx context.Context, jwt, dataType string, data []byte) error {
	md := metadata.New(map[string]string{
		"authorization": "Bearer " + jwt,
	})
	ctx = metadata.NewOutgoingContext(ctx, md)

	if len(data) == 0 || jwt == "" || dataType == "" {
		return errors.New("data, data type, or JWT token is empty")
	}

	req := &proto.PostDataRequest{
		Type: dataType,
		Data: data,
	}

	grpcResp, err := c.ClientConn.PostData(ctx, req)
	if err != nil || !grpcResp.Success {
		return errors.New("failed to post data")
	}

	return nil
}
