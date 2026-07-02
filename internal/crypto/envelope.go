package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"fmt"

	"github.com/Velarvo/velarvo-desktop/internal/storage"
	"github.com/awnumar/memguard"
)

const (
	envelopeHeaderSize = 1
	envelopeNonceSize  = 12
	envelopeTagSize    = 16
	minEnvelopeSize    = envelopeHeaderSize + envelopeNonceSize + envelopeTagSize
)

func Encrypt(key, plaintext []byte) ([]byte, error) {
	if len(key) != 32 {
		return nil, fmt.Errorf("encrypt envelope: expected 32-byte key, got %d", len(key))
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("encrypt envelope: init aes cipher: %w", err)
	}

	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("encrypt envelope: init gcm: %w", err)
	}

	out := make([]byte, envelopeHeaderSize+envelopeNonceSize, envelopeHeaderSize+envelopeNonceSize+len(plaintext)+envelopeTagSize)
	out[0] = byte(storage.CurrentCryptoVersion)

	nonce := out[envelopeHeaderSize : envelopeHeaderSize+envelopeNonceSize]
	if _, err := rand.Read(nonce); err != nil {
		return nil, fmt.Errorf("encrypt envelope: generate nonce: %w", err)
	}

	return aead.Seal(out, nonce, plaintext, nil), nil
}

func Decrypt(key, envelope []byte) (*memguard.LockedBuffer, error) {
	if len(key) != 32 {
		return nil, fmt.Errorf("decrypt envelope: expected 32-byte key, got %d", len(key))
	}
	if len(envelope) < minEnvelopeSize {
		return nil, errors.New("decrypt envelope: payload too short")
	}

	version := storage.CryptoVersion(envelope[0])
	if !version.Valid() {
		return nil, fmt.Errorf("decrypt envelope: unsupported crypto version %d", version)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("decrypt envelope: init aes cipher: %w", err)
	}

	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("decrypt envelope: init gcm: %w", err)
	}

	nonce := envelope[envelopeHeaderSize : envelopeHeaderSize+envelopeNonceSize]
	ciphertext := envelope[envelopeHeaderSize+envelopeNonceSize:]

	plainLen := len(ciphertext) - envelopeTagSize
	if plainLen <= 0 {
		if _, err := aead.Open(nil, nonce, ciphertext, nil); err != nil {
			return nil, errors.New("decrypt envelope: authentication failed")
		}
		return memguard.NewBuffer(0), nil
	}

	buf := memguard.NewBuffer(plainLen)
	if _, err := aead.Open(buf.Bytes()[:0], nonce, ciphertext, nil); err != nil {
		buf.Destroy()
		return nil, errors.New("decrypt envelope: authentication failed")
	}
	buf.Freeze()

	return buf, nil
}

func Zero(data []byte) {
	for i := range data {
		data[i] = 0
	}
}
