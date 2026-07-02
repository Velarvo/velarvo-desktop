package keychain

import (
	"errors"

	"github.com/zalando/go-keyring"
)

const (
	service      = "dev.velarvo.desktop"
	accessToken  = "access_token"
	refreshToken = "refresh_token"
)

type Service struct{}

func New() *Service {
	return &Service{}
}

func (s *Service) SaveTokens(access, refresh string) error {
	if err := keyring.Set(service, accessToken, access); err != nil {
		return err
	}
	return keyring.Set(service, refreshToken, refresh)
}

func (s *Service) GetAccessToken() (string, error) {
	return keyring.Get(service, accessToken)
}

func (s *Service) GetRefreshToken() (string, error) {
	return keyring.Get(service, refreshToken)
}

func (s *Service) ClearTokens() error {
	if err := keyring.Delete(service, accessToken); err != nil {
		if errors.Is(err, keyring.ErrNotFound) {
			return keyring.Delete(service, refreshToken)
		}
		return err
	}

	if err := keyring.Delete(service, refreshToken); err != nil && !errors.Is(err, keyring.ErrNotFound) {
		return err
	}

	return nil
}

func (s *Service) SetSecret(name, value string) error {
	return keyring.Set(service, name, value)
}

func (s *Service) GetSecret(name string) (string, error) {
	return keyring.Get(service, name)
}

func (s *Service) DeleteSecret(name string) error {
	return keyring.Delete(service, name)
}

func ErrNotFound() error {
	return keyring.ErrNotFound
}
