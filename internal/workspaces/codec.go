package workspaces

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Velarvo/velarvo-desktop/internal/storage"
	"github.com/awnumar/memguard"
)

func MarshalData(data WorkspaceData) (*memguard.LockedBuffer, storage.DataVersion, error) {
	data.Name = normalizeName(data.Name)
	if data.Name == "" {
		return nil, 0, fmt.Errorf("%w: workspace name is required", ErrInvalidInput)
	}

	plaintext, err := json.Marshal(data)
	if err != nil {
		return nil, 0, fmt.Errorf("marshal workspace data: %w", err)
	}
	return memguard.NewBufferFromBytes(plaintext), storage.CurrentWorkspaceDataVersion, nil
}

func UnmarshalData(plaintext []byte, dataVersion storage.DataVersion) (*WorkspaceData, error) {
	switch dataVersion {
	case storage.DataVersionV1:
		var data WorkspaceDataV1
		if err := json.Unmarshal(plaintext, &data); err != nil {
			return nil, fmt.Errorf("unmarshal workspace data v1: %w", err)
		}
		data.Name = normalizeName(data.Name)
		return &data, nil
	default:
		return nil, fmt.Errorf("unsupported workspace data version: %d", dataVersion)
	}
}

func normalizeName(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}
