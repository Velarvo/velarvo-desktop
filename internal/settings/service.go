package settings

import (
	"context"
	"fmt"

	"github.com/Velarvo/velarvo-desktop/internal/storage"
)

const (
	keyLanguage = "ui.language"
)

type Service struct {
	repo *Repository
}

func NewService(db *storage.DB) *Service {
	return &Service{repo: NewRepository(db)}
}

func (s *Service) GetAll(ctx context.Context) (Settings, error) {
	language, err := s.GetLanguage(ctx)
	if err != nil {
		return Settings{}, err
	}
	return Settings{Language: language}, nil
}

func (s *Service) GetLanguage(ctx context.Context) (LanguageCode, error) {
	value, ok, err := s.repo.Get(ctx, keyLanguage)
	if err != nil {
		return "", err
	}
	if !ok {
		return DefaultLanguage, nil
	}
	code := LanguageCode(value)
	if !code.Valid() {
		return DefaultLanguage, nil
	}
	return code, nil
}

func (s *Service) SetLanguage(ctx context.Context, code LanguageCode) error {
	if !code.Valid() {
		return fmt.Errorf("%w: %q", ErrUnsupportedLanguage, code)
	}
	return s.repo.Set(ctx, keyLanguage, string(code))
}
