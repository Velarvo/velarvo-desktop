package settings

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/Velarvo/velarvo-desktop/internal/storage"
)

type Repository struct {
	db *storage.DB
}

func NewRepository(db *storage.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Get(ctx context.Context, key string) (string, bool, error) {
	var value string
	err := r.db.Read().
		QueryRowContext(ctx, `SELECT value FROM app_meta WHERE key = ?`, key).
		Scan(&value)
	if errors.Is(err, sql.ErrNoRows) {
		return "", false, nil
	}
	if err != nil {
		return "", false, fmt.Errorf("read app_meta %q: %w", key, err)
	}
	return value, true, nil
}

func (r *Repository) Set(ctx context.Context, key, value string) error {
	_, err := r.db.Write().ExecContext(ctx, `
		INSERT INTO app_meta (key, value) VALUES (?, ?)
		ON CONFLICT(key) DO UPDATE SET value = excluded.value`, key, value)
	if err != nil {
		return fmt.Errorf("write app_meta %q: %w", key, err)
	}
	return nil
}
