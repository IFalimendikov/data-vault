package storage

import (
	"context"
	"data-vault/server/internal/models"
	"database/sql"
	"errors"

	sq "github.com/Masterminds/squirrel"
)

// Login validates user credentials against the database
func (s *Storage) Login(ctx context.Context, user models.User) error {
	var login string

	row := sq.Select("login").
		From("users").
		Where(sq.Eq{
			"login":    user.Login,
			"password": user.Password,
		}).
		RunWith(s.DB).
		PlaceholderFormat(sq.Dollar).
		QueryRowContext(ctx)

	err := row.Scan(&login)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrWrongPassword
		}
		return err
	}

	return nil
}
