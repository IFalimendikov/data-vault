package grpcclient

import (
	"context"
	"data-vault/client/internal/models"
	"data-vault/client/internal/proto"
	"errors"

	"google.golang.org/grpc/metadata"
)

// GetData retrieves all user data from the vault via gRPC
func (c *Client) GetData(ctx context.Context, jwt string) ([]models.Data, error) {
	var resp []models.Data

	md := metadata.New(map[string]string{
		"authorization": "Bearer " + jwt,
	})
	ctx = metadata.NewOutgoingContext(ctx, md)

	if jwt == "" {
		return nil, errors.New("JWT token is empty")
	}

	req := &proto.GetDataRequest{}

	grpcResp, err := c.ClientConn.GetData(ctx, req)
	if err != nil {
		return nil, err
	}

	for _, d := range grpcResp.Data {
		resp = append(resp, models.Data{
			ID:         d.Id,
			User:       d.User,
			Status:     d.Status,
			Type:       d.Type,
			Data:       d.Data,
			UploadedAt: d.UploadedAt,
		})
	}

	return resp, nil
}
