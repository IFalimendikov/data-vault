package service

import (
	"context"
	"data-vault/server/internal/models"
)

func (s *Vault) GetData(ctx context.Context, login string) ([]models.Data, error) {
	var res []models.Data

	if login == "" {
		return nil, ErrMalformedRequest
	}

	data, err := s.Storage.GetData(ctx, login)
	if err != nil {
		return nil, err
	}

	for _, d := range data {
		decryptedData, err := s.decrypt(ctx, d.Data)
		if err != nil {
			return nil, err
		}
		res = append(res, models.Data{
			ID:         d.ID,
			Data:       decryptedData,
			UploadedAt: d.UploadedAt,
			Status:     d.Status,
		})
	}
	return res, nil
}
