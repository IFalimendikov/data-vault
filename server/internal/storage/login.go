package storage

import (
	"context"
	"database/sql"
	"errors"
	"data-vault/server/internal/models"

    sq "github.com/Masterminds/squirrel"
)

func (s *Storage) Login(ctx context.Context, user models.User) error {
    var login string
    
    row := sq.Select("login").
        From("users").
        Where(sq.Eq{
            "login": user.Login,
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