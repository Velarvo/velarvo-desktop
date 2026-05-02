package logger_test

import (
	"testing"

	"github.com/Velarvo/velarvo-desktop/internal/logger"
	"github.com/stretchr/testify/require"
)

func TestInitCreatesLogger(t *testing.T) {
	require.NoError(t, logger.Init(logger.Options{
		Level:       "debug",
		Development: true,
	}))
	t.Cleanup(func() {
		require.NoError(t, logger.Close())
	})

	require.NotNil(t, logger.Default())
	require.NotNil(t, logger.Named("test"))
}

func TestCloseWithoutInitDoesNotPanic(t *testing.T) {
	originalLogger := logger.Log
	logger.Log = nil
	t.Cleanup(func() {
		logger.Log = originalLogger
	})

	require.NotPanics(t, func() {
		_ = logger.Close()
	})
}
