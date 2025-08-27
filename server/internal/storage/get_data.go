package storage

import (
	"context"
	"data-vault/server/internal/models"

	sq "github.com/Masterminds/squirrel"
)

// GetData retrieves all data for a specific user from the database
func (s *Storage) GetData(ctx context.Context, user string) ([]models.Data, error) {
	data := make([]models.Data, 0)

	rows, err := sq.Select("id", "user", "status", "type", "data", "uploaded_at").
		From("storage").
		Where(sq.Eq{"user": user}).
		OrderBy("uploaded_at DESC").
		RunWith(s.DB).
		PlaceholderFormat(sq.Dollar).
		QueryContext(ctx)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var o models.Data
		err := rows.Scan(&o.ID, &o.User, &o.Status, &o.Type, &o.Data, &o.UploadedAt)
		if err != nil {
			return nil, err
		}

		data = append(data, o)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	if len(data) == 0 {
		return nil, ErrNoDataFound
	}

	return data, nil
}
