package storage

import (
	"context"
	"time"

	sq "github.com/Masterminds/squirrel"
)

// PostData stores user data in the database with timestamp
func (s *Storage) PostData(ctx context.Context, login, dataType string, data []byte) error {
	_, err := sq.Insert("storage").
		Columns("user", "status", "type", "data", "uploaded_at").
		Values(login, "NEW", dataType, data, time.Now().UTC().Format(time.RFC3339)).
		RunWith(s.DB).
		PlaceholderFormat(sq.Dollar).
		ExecContext(ctx)

	if err != nil {
		return err
	}

	return nil
}
