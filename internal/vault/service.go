package vault

import (
	"context"
	"crypto/rand"
	"fmt"
	"sync"

	vaultcrypto "github.com/Velarvo/velarvo-desktop/internal/crypto"
	"github.com/Velarvo/velarvo-desktop/internal/storage"
	"github.com/awnumar/memguard"
	"github.com/google/uuid"
)

const (
	keySize             = 32
	saltSize            = 16
	defaultAutoLockSecs = 15 * 60
)

type Service struct {
	repo *Repository
	mu   sync.RWMutex
	dek  *memguard.Enclave
}

var _ vaultcrypto.EnvelopeCipher = (*Service)(nil)

func NewService(db *storage.DB) *Service {
	return &Service{
		repo: NewRepository(db),
	}
}

func (s *Service) State(ctx context.Context) (VaultState, error) {
	isSetup, err := s.repo.HasVault(ctx)
	if err != nil {
		return VaultState{}, err
	}

	state := VaultState{
		IsSetup:    isSetup,
		IsUnlocked: s.IsUnlocked(),
	}
	if !isSetup {
		return state, nil
	}

	meta, err := s.repo.GetMeta(ctx)
	if err != nil {
		return VaultState{}, err
	}
	state.AutoLockSeconds = meta.AutoLockSeconds
	return state, nil
}

func (s *Service) Setup(ctx context.Context, req SetupRequest) (VaultState, error) {
	password := req.MasterPassword
	if len(password) < masterPasswordMinLength {
		return VaultState{}, ErrWeakMasterPassword
	}

	exists, err := s.repo.HasVault(ctx)
	if err != nil {
		return VaultState{}, err
	}
	if exists {
		return VaultState{}, ErrAlreadySetup
	}

	saltKEK, err := randomBytes(saltSize)
	if err != nil {
		return VaultState{}, err
	}
	saltAuth, err := randomBytes(saltSize)
	if err != nil {
		return VaultState{}, err
	}

	dekBuf := memguard.NewBufferRandom(keySize)
	defer dekBuf.Destroy()

	kdfThreads := defaultKDFThreads()
	kek, err := deriveKEK(password, saltKEK, defaultKDFTime, defaultKDFMemoryKiB, kdfThreads, defaultKDFKeyLen)
	if err != nil {
		return VaultState{}, err
	}
	kekBuf := memguard.NewBufferFromBytes(kek)
	defer kekBuf.Destroy()

	passwordEnvelope, err := vaultcrypto.Encrypt(kekBuf.Bytes(), dekBuf.Bytes())
	if err != nil {
		return VaultState{}, fmt.Errorf("encrypt dek envelope: %w", err)
	}

	deviceID, err := uuid.NewV7()
	if err != nil {
		return VaultState{}, fmt.Errorf("generate vault device id: %w", err)
	}

	now := storage.Now()
	meta := storage.VaultMeta{
		SchemaVersion:   storage.CurrentSchemaVersion,
		CryptoVersion:   storage.CurrentCryptoVersion,
		CreatedAt:       now,
		KDFID:           storage.KDFArgon2id,
		KDFTime:         defaultKDFTime,
		KDFMemory:       defaultKDFMemoryKiB,
		KDFThreads:      kdfThreads,
		KDFKeyLen:       defaultKDFKeyLen,
		SaltKEK:         saltKEK,
		SaltAuth:        saltAuth,
		DeviceID:        deviceID.String(),
		AutoLockSeconds: defaultAutoLockSecs,
	}
	envelope := storage.DEKEnvelope{
		Method:    storage.DEKEnvelopePassword,
		Envelope:  passwordEnvelope,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.repo.InsertInitialVault(ctx, meta, envelope); err != nil {
		return VaultState{}, err
	}

	s.storeDEK(dekBuf)
	return s.State(ctx)
}

func (s *Service) Unlock(ctx context.Context, req UnlockRequest) (VaultState, error) {
	password := req.MasterPassword
	if password == "" {
		return VaultState{}, ErrInvalidMasterPassword
	}

	meta, err := s.repo.GetMeta(ctx)
	if err != nil {
		return VaultState{}, err
	}
	if !meta.SchemaVersion.Supported() {
		return VaultState{}, fmt.Errorf("unsupported vault schema version %d", meta.SchemaVersion)
	}

	envelope, err := s.repo.GetEnvelope(ctx, storage.DEKEnvelopePassword)
	if err != nil {
		return VaultState{}, err
	}

	kek, err := deriveKEK(password, meta.SaltKEK, meta.KDFTime, meta.KDFMemory, meta.KDFThreads, meta.KDFKeyLen)
	if err != nil {
		return VaultState{}, err
	}
	kekBuf := memguard.NewBufferFromBytes(kek)
	defer kekBuf.Destroy()

	dekBuf, err := vaultcrypto.Decrypt(kekBuf.Bytes(), envelope.Envelope)
	if err != nil {
		return VaultState{}, ErrInvalidMasterPassword
	}
	defer dekBuf.Destroy()
	if dekBuf.Size() != keySize {
		return VaultState{}, ErrInvalidMasterPassword
	}

	s.storeDEK(dekBuf)
	return s.State(ctx)
}

func (s *Service) Lock() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.dek = nil
}

func (s *Service) IsUnlocked() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.dek != nil
}

func (s *Service) Encrypt(plaintext []byte) ([]byte, error) {
	dek, err := s.openDEK()
	if err != nil {
		return nil, err
	}
	defer dek.Destroy()
	return vaultcrypto.Encrypt(dek.Bytes(), plaintext)
}

func (s *Service) Decrypt(envelope []byte) (*memguard.LockedBuffer, error) {
	dek, err := s.openDEK()
	if err != nil {
		return nil, err
	}
	defer dek.Destroy()
	return vaultcrypto.Decrypt(dek.Bytes(), envelope)
}

func (s *Service) storeDEK(buf *memguard.LockedBuffer) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.dek = buf.Seal()
}

func (s *Service) openDEK() (*memguard.LockedBuffer, error) {
	s.mu.RLock()
	enc := s.dek
	s.mu.RUnlock()

	if enc == nil {
		return nil, ErrLocked
	}
	dek, err := enc.Open()
	if err != nil {
		return nil, fmt.Errorf("open vault key: %w", err)
	}
	return dek, nil
}

func randomBytes(size int) ([]byte, error) {
	out := make([]byte, size)
	if _, err := rand.Read(out); err != nil {
		return nil, fmt.Errorf("generate random bytes: %w", err)
	}
	return out, nil
}
