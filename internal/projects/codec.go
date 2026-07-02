package projects

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Velarvo/velarvo-desktop/internal/storage"
	"github.com/awnumar/memguard"
)

func MarshalData(data ProjectData) (*memguard.LockedBuffer, storage.DataVersion, error) {
	data.Name = normalizeName(data.Name)
	if data.Name == "" {
		return nil, 0, fmt.Errorf("%w: project name is required", ErrInvalidInput)
	}

	plaintext, err := json.Marshal(data)
	if err != nil {
		return nil, 0, fmt.Errorf("marshal project data: %w", err)
	}
	return memguard.NewBufferFromBytes(plaintext), storage.CurrentProjectDataVersion, nil
}

func UnmarshalData(plaintext []byte, dataVersion storage.DataVersion) (*ProjectData, error) {
	switch dataVersion {
	case storage.DataVersionV1:
		var data ProjectDataV1
		if err := json.Unmarshal(plaintext, &data); err != nil {
			return nil, fmt.Errorf("unmarshal project data v1: %w", err)
		}
		data.Name = normalizeName(data.Name)
		return &data, nil
	default:
		return nil, fmt.Errorf("unsupported project data version: %d", dataVersion)
	}
}

func normalizeName(name string) string {
	return strings.TrimSpace(name)
}
