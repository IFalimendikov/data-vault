package storage

import (
	"context"
	"time"

	sq "github.com/Masterminds/squirrel"
)

func (s *Storage) PostData(ctx context.Context, login, data string) error {
	_, err := sq.Insert("storage").
		Columns("user", "status", "data", "uploaded_at").
		Values(login, "NEW", data, time.Now().UTC().Format(time.RFC3339)).
		RunWith(s.DB).
		PlaceholderFormat(sq.Dollar).
		ExecContext(ctx)

	if err != nil {
		return err
	}

	return nil
}
