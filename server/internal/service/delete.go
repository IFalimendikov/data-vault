package service

import (
	"context"
)

// DeleteData removes a specific data entry for a user
func (s *Vault) DeleteData(ctx context.Context, login, id string) error {
	err := s.Storage.DeleteData(ctx, login, id)
	if err != nil {
		return err
	}
	return nil
}
