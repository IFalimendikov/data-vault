package storage

import (
	"context"
	"database/sql"
	"data-vault/server/internal/config"
)

type Storage struct {
	cfg *config.Config
	DB  *sql.DB
}

var (
	UsersQuery       = `CREATE TABLE IF NOT EXISTS users (login text PRIMARY KEY, password text);`
	StorageQuery     = `CREATE TABLE IF NOT EXISTS storage (id SERIAL PRIMARY KEY, user text, status text, data text, uploaded_at text);`
)

func New(ctx context.Context, cfg *config.Config) (*Storage, error) {
	if cfg.DatabaseURI == "" {
		return nil, ErrBadConn
	}

	db, err := sql.Open("pgx", cfg.DatabaseURI)
	if err != nil {
		return nil, ErrBadConn
	}

	err = db.Ping()
	if err != nil {
		return nil, ErrBadConn
	}

	tables := []string{UsersQuery, StorageQuery}

	for _, q := range tables {
		_, err = db.ExecContext(ctx, q)
		if err != nil {
			return nil, err
		}
	}

	storage := Storage{
		cfg: cfg,
		DB:  db,
	}

	return &storage, nil
}
