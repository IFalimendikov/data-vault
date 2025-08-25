package service

import (
	"context"
	"data-vault/server/internal/models"
)

// GetData retrieves and decrypts all data for a specific user
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
		decryptedData, err := s.decryptBytes(ctx, d.Data)
		if err != nil {
			return nil, err
		}
		res = append(res, models.Data{
			ID:         d.ID,
			User:       d.User,
			Status:     d.Status,
			Type:       d.Type,
			Data:       decryptedData,
			UploadedAt: d.UploadedAt,
		})
	}
	return res, nil
}
