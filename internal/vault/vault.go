package vault

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"sync"

	"github.com/Velarvo/velarvo-desktop/internal/apperr"
	"github.com/Velarvo/velarvo-desktop/internal/db"
	"github.com/awnumar/memguard"
	"golang.org/x/crypto/argon2"
)

var (
	ErrNotSetup      = apperr.New("VAULT_NOT_SETUP", "vault is not set up")
	ErrAlreadySetup  = apperr.New("VAULT_ALREADY_SETUP", "vault is already set up")
	ErrLocked        = apperr.New("VAULT_LOCKED", "vault is locked")
	ErrWrongPassword = apperr.New("INVALID_MASTER_PASSWORD", "invalid master password")
	ErrWeakPassword  = apperr.New("WEAK_MASTER_PASSWORD", "master password is too weak")
)

const (
	minPasswordLength = 12
	saltSize          = 16
	autoLockSeconds   = 15 * 60

	kdfTime      = 3
	kdfMemoryKiB = 64 * 1024
	kdfThreads   = 4
)

type kdfParams struct {
	time      int
	memoryKiB int
	threads   int
}

var currentKDF = kdfParams{time: kdfTime, memoryKiB: kdfMemoryKiB, threads: kdfThreads}

func (p kdfParams) valid() bool {
	return p.time >= 1 && p.time <= 16 &&
		p.memoryKiB >= 8*1024 && p.memoryKiB <= 512*1024 &&
		p.threads >= 1 && p.threads <= 64
}

func (p kdfParams) weakerThan(q kdfParams) bool {
	return p.time < q.time || p.memoryKiB < q.memoryKiB || p.threads < q.threads
}

type Vault struct {
	mu  sync.RWMutex
	key *memguard.Enclave
}

func (v *Vault) Unlocked() error {
	v.mu.RLock()
	defer v.mu.RUnlock()
	if v.key == nil {
		return ErrLocked
	}
	return nil
}

func (v *Vault) Encrypt(plaintext []byte) ([]byte, error) {
	key, err := v.openKey()
	if err != nil {
		return nil, err
	}
	defer key.Destroy()
	return seal(key.Bytes(), plaintext)
}

func (v *Vault) Decrypt(envelope []byte) (*memguard.LockedBuffer, error) {
	key, err := v.openKey()
	if err != nil {
		return nil, err
	}
	defer key.Destroy()
	return open(key.Bytes(), envelope)
}

func (v *Vault) Seal(value any) ([]byte, error) {
	plaintext, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	defer clear(plaintext)
	return v.Encrypt(plaintext)
}

func (v *Vault) Open(envelope []byte, value any) error {
	plaintext, err := v.Decrypt(envelope)
	if err != nil {
		return err
	}
	defer plaintext.Destroy()
	return json.Unmarshal(plaintext.Bytes(), value)
}

func (v *Vault) setKey(key *memguard.LockedBuffer) {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.key = key.Seal()
}

func (v *Vault) lock() {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.key = nil
}

func (v *Vault) openKey() (*memguard.LockedBuffer, error) {
	v.mu.RLock()
	enclave := v.key
	v.mu.RUnlock()
	if enclave == nil {
		return nil, ErrLocked
	}
	return enclave.Open()
}

type State struct {
	IsSetup         bool `json:"isSetup"`
	IsUnlocked      bool `json:"isUnlocked"`
	AutoLockSeconds int  `json:"autoLockSeconds"`
}

type Service struct {
	db    *sql.DB
	vault *Vault
}

func NewService(database *sql.DB, v *Vault) *Service {
	return &Service{db: database, vault: v}
}

func (s *Service) ServiceShutdown() error {
	s.vault.lock()
	memguard.Purge()
	return nil
}

func (s *Service) State(ctx context.Context) (State, error) {
	state := State{IsUnlocked: s.vault.Unlocked() == nil}
	err := s.db.QueryRowContext(ctx,
		`SELECT auto_lock_seconds FROM vault WHERE id = 1`).Scan(&state.AutoLockSeconds)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return state, nil
	case err != nil:
		return State{}, fmt.Errorf("read vault: %w", err)
	}
	state.IsSetup = true
	return state, nil
}

func (s *Service) Setup(ctx context.Context, password string) (State, error) {
	if len(password) < minPasswordLength {
		return State{}, ErrWeakPassword
	}
	if state, err := s.State(ctx); err != nil {
		return State{}, err
	} else if state.IsSetup {
		return State{}, ErrAlreadySetup
	}

	dataKey := memguard.NewBufferRandom(keySize)
	defer dataKey.Destroy()

	salt, wrapped, err := wrapKey(password, dataKey, currentKDF)
	if err != nil {
		return State{}, err
	}

	if _, err := s.db.ExecContext(ctx, `
		INSERT INTO vault (id, kdf_time, kdf_memory, kdf_threads, salt, wrapped_key, auto_lock_seconds, created_at)
		VALUES (1, ?, ?, ?, ?, ?, ?, ?)`,
		currentKDF.time, currentKDF.memoryKiB, currentKDF.threads, salt, wrapped, autoLockSeconds, db.Now(),
	); err != nil {
		return State{}, fmt.Errorf("save vault: %w", err)
	}

	s.vault.setKey(dataKey)
	return s.State(ctx)
}

func (s *Service) Unlock(ctx context.Context, password string) (State, error) {
	if password == "" {
		return State{}, ErrWrongPassword
	}

	var (
		params        kdfParams
		salt, wrapped []byte
	)
	err := s.db.QueryRowContext(ctx,
		`SELECT kdf_time, kdf_memory, kdf_threads, salt, wrapped_key FROM vault WHERE id = 1`,
	).Scan(&params.time, &params.memoryKiB, &params.threads, &salt, &wrapped)
	if errors.Is(err, sql.ErrNoRows) {
		return State{}, ErrNotSetup
	}
	if err != nil {
		return State{}, fmt.Errorf("read vault: %w", err)
	}

	if !params.valid() {
		return State{}, errors.New("unsupported vault parameters")
	}

	kek := deriveKey(password, salt, params)
	defer kek.Destroy()

	dataKey, err := open(kek.Bytes(), wrapped)
	if err != nil {
		return State{}, ErrWrongPassword
	}
	defer dataKey.Destroy()
	if dataKey.Size() != keySize {
		return State{}, ErrWrongPassword
	}

	if params.weakerThan(currentKDF) {
		if err := s.rewrapKey(ctx, password, dataKey); err != nil {
			slog.Warn("upgrade vault key derivation", "error", err)
		}
	}

	s.vault.setKey(dataKey)
	return s.State(ctx)
}

func (s *Service) Lock(ctx context.Context) (State, error) {
	s.vault.lock()
	return s.State(ctx)
}

func (s *Service) rewrapKey(ctx context.Context, password string, dataKey *memguard.LockedBuffer) error {
	salt, wrapped, err := wrapKey(password, dataKey, currentKDF)
	if err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx, `
		UPDATE vault SET kdf_time = ?, kdf_memory = ?, kdf_threads = ?, salt = ?, wrapped_key = ?
		WHERE id = 1`,
		currentKDF.time, currentKDF.memoryKiB, currentKDF.threads, salt, wrapped,
	)
	if err != nil {
		return fmt.Errorf("save vault: %w", err)
	}
	return nil
}

func wrapKey(password string, dataKey *memguard.LockedBuffer, params kdfParams) (salt, wrapped []byte, err error) {
	salt = make([]byte, saltSize)
	if _, err := rand.Read(salt); err != nil {
		return nil, nil, err
	}

	kek := deriveKey(password, salt, params)
	defer kek.Destroy()

	wrapped, err = seal(kek.Bytes(), dataKey.Bytes())
	if err != nil {
		return nil, nil, fmt.Errorf("wrap data key: %w", err)
	}
	return salt, wrapped, nil
}

func deriveKey(password string, salt []byte, params kdfParams) *memguard.LockedBuffer {
	//nolint:gosec // the parameters are bounded by kdfParams.valid
	return memguard.NewBufferFromBytes(argon2.IDKey(
		[]byte(password), salt, uint32(params.time), uint32(params.memoryKiB), uint8(params.threads), keySize))
}
