package storage

import (
	"context"
	"errors"
	"data-vault/server/internal/models"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

func (s *Storage) Register(ctx context.Context, runner sq.BaseRunner, user models.User) error {
    _, err := sq.Insert("users").
        Columns("login", "password").
        Values(user.Login, user.Password).
        RunWith(runner).
        PlaceholderFormat(sq.Dollar).
        ExecContext(ctx)
    if err != nil {
        var pgErr *pgconn.PgError
        if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
            return ErrDuplicateLogin
        }
        return err
    }

	return nil
}
