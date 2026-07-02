package crypto

import "github.com/awnumar/memguard"

type EnvelopeCipher interface {
	IsUnlocked() bool
	Encrypt(plaintext []byte) ([]byte, error)
	Decrypt(envelope []byte) (*memguard.LockedBuffer, error)
}
