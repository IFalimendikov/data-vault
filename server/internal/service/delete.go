package service

import (
	"context"
)

func (s *Vault) DeleteData(ctx context.Context, login, id string) error {
	err := s.Storage.DeleteData(ctx, login, id)
	if err != nil {
		return err
	}
	return nil
}
