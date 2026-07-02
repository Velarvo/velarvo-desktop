package vault

import (
	"runtime"

	"golang.org/x/crypto/argon2"
)

const (
	masterPasswordMinLength = 12
	defaultKDFTime          = 3
	defaultKDFMemoryKiB     = 64 * 1024
	defaultKDFKeyLen        = 32
	maxKDFTime              = 16
	maxKDFMemoryKiB         = 512 * 1024
	maxKDFThreads           = 64
	maxKDFKeyLen            = 64
)

func defaultKDFThreads() int {
	threads := runtime.NumCPU()
	if threads > 4 {
		return 4
	}
	if threads < 1 {
		return 1
	}
	return threads
}

func deriveKEK(password string, salt []byte, timeCost, memoryKiB, threads, keyLen int) ([]byte, error) {
	if timeCost < 1 || timeCost > maxKDFTime ||
		memoryKiB < 1 || memoryKiB > maxKDFMemoryKiB ||
		threads < 1 || threads > maxKDFThreads ||
		keyLen < 1 || keyLen > maxKDFKeyLen {
		return nil, ErrInvalidKDFParameters
	}

	return argon2.IDKey(
		[]byte(password),
		salt,
		uint32(timeCost),
		uint32(memoryKiB),
		uint8(threads),
		uint32(keyLen),
	), nil
}
