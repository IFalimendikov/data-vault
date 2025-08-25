package storage

import (
	"context"
	"data-vault/server/internal/config"
	"database/sql"
)

// Storage handles database operations and configuration
type Storage struct {
	cfg *config.Config
	DB  *sql.DB
}

// Database table creation queries
var (
	UsersQuery   = `CREATE TABLE IF NOT EXISTS users (login text PRIMARY KEY, password text);`
	StorageQuery = `CREATE TABLE IF NOT EXISTS storage (id SERIAL PRIMARY KEY, user text, status text, type text, data bytea, uploaded_at text);`
)

// New creates and initializes a new storage instance with database connection
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
