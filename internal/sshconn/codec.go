package sshconn

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Velarvo/velarvo-desktop/internal/storage"
	"github.com/awnumar/memguard"
)

const defaultPort = 22

func normalizeListing(data *ListingData) error {
	data.Name = strings.TrimSpace(data.Name)
	data.Host = strings.TrimSpace(data.Host)
	data.Username = strings.TrimSpace(data.Username)

	if data.Host == "" {
		return fmt.Errorf("%w: host is required", ErrInvalidInput)
	}
	if data.Username == "" {
		return fmt.Errorf("%w: username is required", ErrInvalidInput)
	}
	if data.Port <= 0 || data.Port > 65535 {
		data.Port = defaultPort
	}
	if data.Name == "" {
		data.Name = data.Host
	}
	return nil
}

func MarshalListing(data ListingData) (*memguard.LockedBuffer, storage.DataVersion, error) {
	if err := normalizeListing(&data); err != nil {
		return nil, 0, err
	}

	plaintext, err := json.Marshal(data)
	if err != nil {
		return nil, 0, fmt.Errorf("marshal ssh listing: %w", err)
	}
	return memguard.NewBufferFromBytes(plaintext), storage.CurrentSSHListingVersion, nil
}

func UnmarshalListing(plaintext []byte, version storage.DataVersion) (*ListingData, error) {
	switch version {
	case storage.DataVersionV1:
		var data ListingDataV1
		if err := json.Unmarshal(plaintext, &data); err != nil {
			return nil, fmt.Errorf("unmarshal ssh listing v1: %w", err)
		}
		return &data, nil
	default:
		return nil, fmt.Errorf("unsupported ssh listing version: %d", version)
	}
}

func MarshalSecret(data SecretData) (*memguard.LockedBuffer, storage.DataVersion, error) {
	plaintext, err := json.Marshal(data)
	if err != nil {
		return nil, 0, fmt.Errorf("marshal ssh secret: %w", err)
	}
	return memguard.NewBufferFromBytes(plaintext), storage.CurrentSSHSecretVersion, nil
}

func UnmarshalSecret(plaintext []byte, version storage.DataVersion) (*SecretData, error) {
	switch version {
	case storage.DataVersionV1:
		var data SecretDataV1
		if err := json.Unmarshal(plaintext, &data); err != nil {
			return nil, fmt.Errorf("unmarshal ssh secret v1: %w", err)
		}
		return &data, nil
	default:
		return nil, fmt.Errorf("unsupported ssh secret version: %d", version)
	}
}
