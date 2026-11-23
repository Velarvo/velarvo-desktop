package vault

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"fmt"

	"github.com/awnumar/memguard"
)

const (
	envelopeVersion = 1
	keySize         = 32
	nonceSize       = 12
	tagSize         = 16
	headerSize      = 1 + nonceSize
)

var errDecrypt = errors.New("decrypt: authentication failed")

func newGCM(key []byte) (cipher.AEAD, error) {
	if len(key) != keySize {
		return nil, fmt.Errorf("expected %d-byte key, got %d", keySize, len(key))
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	return cipher.NewGCM(block)
}

func seal(key, plaintext []byte) ([]byte, error) {
	aead, err := newGCM(key)
	if err != nil {
		return nil, fmt.Errorf("encrypt: %w", err)
	}

	out := make([]byte, headerSize, headerSize+len(plaintext)+tagSize)
	out[0] = envelopeVersion
	nonce := out[1:headerSize]
	if _, err := rand.Read(nonce); err != nil {
		return nil, fmt.Errorf("encrypt: generate nonce: %w", err)
	}
	return aead.Seal(out, nonce, plaintext, nil), nil
}

func open(key, envelope []byte) (*memguard.LockedBuffer, error) {
	aead, err := newGCM(key)
	if err != nil {
		return nil, fmt.Errorf("decrypt: %w", err)
	}
	if len(envelope) < headerSize+tagSize {
		return nil, errors.New("decrypt: envelope too short")
	}
	if envelope[0] != envelopeVersion {
		return nil, fmt.Errorf("decrypt: unsupported envelope version %d", envelope[0])
	}

	nonce, ciphertext := envelope[1:headerSize], envelope[headerSize:]
	size := len(ciphertext) - tagSize
	if size == 0 {
		if _, err := aead.Open(nil, nonce, ciphertext, nil); err != nil {
			return nil, errDecrypt
		}
		return memguard.NewBuffer(0), nil
	}

	buf := memguard.NewBuffer(size)
	if _, err := aead.Open(buf.Bytes()[:0], nonce, ciphertext, nil); err != nil {
		buf.Destroy()
		return nil, errDecrypt
	}
	buf.Freeze()
	return buf, nil
}
