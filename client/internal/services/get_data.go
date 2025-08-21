package services

import (
	"context"
	"data-vault/client/internal/models"
)

// GetURL retrieves the original URL from grpcclient using the shortened URL as a key
func (v *Vault) GetData(ctx context.Context, jwt string) ([]models.Data, error) {
	var res []models.Data

	data, err := v.grpcclient.GetData(ctx, jwt)
	if err != nil {
		return res, nil
	}

	for _, d := range data {
		res = append(res, d)
	}

	return res, nil
}
