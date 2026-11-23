package settings

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/Velarvo/velarvo-desktop/internal/apperr"
)

var ErrUnsupportedLanguage = apperr.New("SETTINGS_UNSUPPORTED_VALUE", "unsupported language")

type Language string

const (
	English Language = "en"
	Czech   Language = "cs"
)

const keyLanguage = "ui.language"

type Settings struct {
	Language Language `json:"language"`
}

type Service struct {
	db *sql.DB
}

func NewService(db *sql.DB) *Service {
	return &Service{db: db}
}

func (s *Service) Get(ctx context.Context) (Settings, error) {
	var value string
	err := s.db.QueryRowContext(ctx, `SELECT value FROM settings WHERE key = ?`, keyLanguage).Scan(&value)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return Settings{}, fmt.Errorf("read language: %w", err)
	}

	language := Language(value)
	if language != Czech {
		language = English
	}
	return Settings{Language: language}, nil
}

func (s *Service) SetLanguage(ctx context.Context, language Language) (Settings, error) {
	if language != English && language != Czech {
		return Settings{}, ErrUnsupportedLanguage
	}
	if _, err := s.db.ExecContext(ctx, `
		INSERT INTO settings (key, value) VALUES (?, ?)
		ON CONFLICT (key) DO UPDATE SET value = excluded.value`, keyLanguage, string(language)); err != nil {
		return Settings{}, fmt.Errorf("save language: %w", err)
	}
	return Settings{Language: language}, nil
}
