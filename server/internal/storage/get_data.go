package storage

import (
	"context"
	"data-vault/server/internal/models"

	sq "github.com/Masterminds/squirrel"
)

func (s *Storage) GetData(ctx context.Context, user string) ([]models.Data, error) {
	data := make([]models.Data, 0)

	rows, err := sq.Select("id", "data", "uploaded_at").
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
		err := rows.Scan(&o.ID, &o.Status, &o.Data, &o.UploadedAt)
		if err != nil {
			return nil, err
		}

		data = append(data, o)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	if len(data) == 0 {
		return nil, ErrNoOrdersFound
	}

	return data, nil
}
