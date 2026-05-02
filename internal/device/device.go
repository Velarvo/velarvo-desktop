package device

import (
	"os"
	"runtime"

	"github.com/google/uuid"
	"github.com/zalando/go-keyring"
)

const (
	service  = "dev.velarvo.desktop"
	deviceID = "device_id"
)

func GetOrCreateID() (string, error) {
	id, err := keyring.Get(service, deviceID)
	if err == nil {
		return id, nil
	}

	newID := uuid.New().String()
	if err := keyring.Set(service, deviceID, newID); err != nil {
		return "", err
	}
	return newID, nil
}

func GetOS() string {
	if runtime.GOOS == "darwin" {
		return "macOS"
	}

	return runtime.GOOS
}

func GetName() string {
	hostname, err := os.Hostname()
	if err != nil {
		return "Unknown Device"
	}
	return hostname
}
