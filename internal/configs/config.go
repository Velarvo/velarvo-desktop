package configs

import (
	"os"
	"strings"
)

type Config struct {
	BaseURL  string
	DevMode  bool
	LogLevel string
}

func LoadConfig() Config {
	devMode := os.Getenv("VELARVO_DEV") == "true"
	baseURL := strings.TrimSpace(os.Getenv("VELARVO_API_BASE_URL"))
	logLevel := strings.TrimSpace(os.Getenv("VELARVO_LOG_LEVEL"))
	if baseURL == "" {
		baseURL = "http://localhost:3000/api"
	}
	if logLevel == "" {
		if devMode {
			logLevel = "debug"
		} else {
			logLevel = "info"
		}
	}

	return Config{
		BaseURL:  baseURL,
		DevMode:  devMode,
		LogLevel: logLevel,
	}
}
