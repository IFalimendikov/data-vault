package storage

import (
	"context"

	sq "github.com/Masterminds/squirrel"
)

func (s *Storage) DeleteData(ctx context.Context, login, id string) error {
    _, err := sq.Delete("storage").
        Where(sq.And{
            sq.Eq{"user": login},
            sq.Eq{"id": id},
        }).
        PlaceholderFormat(sq.Dollar).
        RunWith(s.DB).
        ExecContext(ctx)

    if err != nil {
        return err
    }

    return nil
}
