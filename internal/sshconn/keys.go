package sshconn

import (
	"context"
	"crypto"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"database/sql"
	"encoding/pem"
	"errors"
	"fmt"
	"strings"

	"github.com/Velarvo/velarvo-desktop/internal/apperr"
	"github.com/Velarvo/velarvo-desktop/internal/db"
	"github.com/google/uuid"
	"golang.org/x/crypto/ssh"
)

var (
	ErrKeyNotFound  = apperr.New("SSH_KEY_NOT_FOUND", "ssh key not found")
	ErrInvalidKey   = apperr.New("SSH_INVALID_KEY", "invalid ssh private key")
	ErrPublicKey    = apperr.New("SSH_INVALID_KEY", "this is a public key; paste the private key")
	ErrKeyType      = apperr.New("SSH_INVALID_KEY", "unsupported ssh key type")
	ErrDuplicateKey = apperr.New("SSH_DUPLICATE_KEY", "this key is already in the workspace")
)

const (
	KeyEd25519  = "ed25519"
	KeyRSA4096  = "rsa4096"
	KeyECDSA384 = "ecdsa384"
)

type Key struct {
	ID                string `json:"id"`
	WorkspaceID       string `json:"workspaceId"`
	Label             string `json:"label"`
	KeyType           string `json:"keyType"`
	PublicKey         string `json:"publicKey"`
	FingerprintSHA256 string `json:"fingerprintSha256"`
	HasPassphrase     bool   `json:"hasPassphrase"`
	LastUsedAt        int64  `json:"lastUsedAt,omitempty"`
	CreatedAt         int64  `json:"createdAt"`
}

type NewKey struct {
	Label      string `json:"label"`
	KeyType    string `json:"keyType"`
	Passphrase string `json:"passphrase"`
}

type ImportedKey struct {
	Label      string `json:"label"`
	PrivateKey string `json:"privateKey"`
	Passphrase string `json:"passphrase"`
}

type privateKey struct {
	PEM        string `json:"privateKey"`
	Passphrase string `json:"passphrase,omitempty"`
}

func (p privateKey) signer() (ssh.Signer, error) {
	var (
		signer ssh.Signer
		err    error
	)
	if p.Passphrase != "" {
		signer, err = ssh.ParsePrivateKeyWithPassphrase([]byte(p.PEM), []byte(p.Passphrase))
	} else {
		signer, err = ssh.ParsePrivateKey([]byte(p.PEM))
	}
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidKey, err)
	}
	return signer, nil
}

func (s *Service) ListKeys(ctx context.Context, workspaceID string) ([]Key, error) {
	if err := s.vault.Unlocked(); err != nil {
		return nil, err
	}

	rows, err := s.db.QueryContext(ctx, `
		SELECT id, workspace_id, label, key_type, public_key, fingerprint, has_passphrase,
		       COALESCE(last_used_at, 0), created_at
		FROM ssh_keys
		WHERE workspace_id = ?
		ORDER BY label COLLATE NOCASE, created_at`, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("list ssh keys: %w", err)
	}
	defer func() { _ = rows.Close() }()

	keys := []Key{}
	for rows.Next() {
		var k Key
		if err := rows.Scan(&k.ID, &k.WorkspaceID, &k.Label, &k.KeyType, &k.PublicKey,
			&k.FingerprintSHA256, &k.HasPassphrase, &k.LastUsedAt, &k.CreatedAt); err != nil {
			return nil, err
		}
		keys = append(keys, k)
	}
	return keys, rows.Err()
}

func (s *Service) GenerateKey(ctx context.Context, workspaceID string, input NewKey) (Key, error) {
	var (
		key          crypto.Signer
		err          error
		keyType      = input.KeyType
		defaultLabel string
	)
	switch keyType {
	case "", KeyEd25519:
		keyType, defaultLabel = KeyEd25519, "Ed25519 key"
		_, key, err = ed25519.GenerateKey(rand.Reader)
	case KeyRSA4096:
		defaultLabel = "RSA 4096 key"
		key, err = rsa.GenerateKey(rand.Reader, 4096)
	case KeyECDSA384:
		defaultLabel = "ECDSA P-384 key"
		key, err = ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	default:
		return Key{}, ErrKeyType
	}
	if err != nil {
		return Key{}, fmt.Errorf("generate %s key: %w", keyType, err)
	}

	label := labelOr(input.Label, defaultLabel)
	var block *pem.Block
	if input.Passphrase != "" {
		block, err = ssh.MarshalPrivateKeyWithPassphrase(key, label, []byte(input.Passphrase))
	} else {
		block, err = ssh.MarshalPrivateKey(key, label)
	}
	if err != nil {
		return Key{}, fmt.Errorf("encode private key: %w", err)
	}

	signer, err := ssh.NewSignerFromKey(key)
	if err != nil {
		return Key{}, fmt.Errorf("create signer: %w", err)
	}
	private := privateKey{PEM: string(pem.EncodeToMemory(block)), Passphrase: input.Passphrase}
	return s.saveKey(ctx, workspaceID, label, keyType, signer, private)
}

func (s *Service) ImportKey(ctx context.Context, workspaceID string, input ImportedKey) (Key, error) {
	text := strings.TrimSpace(input.PrivateKey)
	if _, _, _, _, err := ssh.ParseAuthorizedKey([]byte(text)); err == nil {
		return Key{}, ErrPublicKey
	}

	private := privateKey{PEM: text, Passphrase: input.Passphrase}
	signer, err := private.signer()
	if err != nil {
		return Key{}, err
	}

	var keyType, defaultLabel string
	switch signer.PublicKey().Type() {
	case ssh.KeyAlgoED25519:
		keyType, defaultLabel = KeyEd25519, "Imported Ed25519 key"
	case ssh.KeyAlgoRSA:
		keyType, defaultLabel = KeyRSA4096, "Imported RSA key"
	case ssh.KeyAlgoECDSA384:
		keyType, defaultLabel = KeyECDSA384, "Imported ECDSA key"
	default:
		return Key{}, ErrKeyType
	}
	return s.saveKey(ctx, workspaceID, labelOr(input.Label, defaultLabel), keyType, signer, private)
}

func (s *Service) saveKey(ctx context.Context, workspaceID, label, keyType string, signer ssh.Signer, private privateKey) (Key, error) {
	sealed, err := s.vault.Seal(private)
	if err != nil {
		return Key{}, fmt.Errorf("encrypt private key: %w", err)
	}

	key := Key{
		ID:                uuid.Must(uuid.NewV7()).String(),
		WorkspaceID:       workspaceID,
		Label:             label,
		KeyType:           keyType,
		PublicKey:         strings.TrimSpace(string(ssh.MarshalAuthorizedKey(signer.PublicKey()))),
		FingerprintSHA256: ssh.FingerprintSHA256(signer.PublicKey()),
		HasPassphrase:     private.Passphrase != "",
		CreatedAt:         db.Now(),
	}

	var duplicate bool
	if err := s.db.QueryRowContext(ctx,
		`SELECT EXISTS (SELECT 1 FROM ssh_keys WHERE workspace_id = ? AND fingerprint = ?)`,
		workspaceID, key.FingerprintSHA256).Scan(&duplicate); err != nil {
		return Key{}, fmt.Errorf("check ssh key: %w", err)
	}
	if duplicate {
		return Key{}, ErrDuplicateKey
	}

	if _, err := s.db.ExecContext(ctx, `
		INSERT INTO ssh_keys (id, workspace_id, label, key_type, public_key, fingerprint, has_passphrase, private_key, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		key.ID, workspaceID, key.Label, key.KeyType, key.PublicKey, key.FingerprintSHA256,
		key.HasPassphrase, sealed, key.CreatedAt,
	); err != nil {
		return Key{}, fmt.Errorf("insert ssh key: %w", err)
	}
	return key, nil
}

func (s *Service) signer(ctx context.Context, keyID string) (ssh.Signer, error) {
	var sealed []byte
	err := s.db.QueryRowContext(ctx, `SELECT private_key FROM ssh_keys WHERE id = ?`, keyID).Scan(&sealed)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrKeyNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("read ssh key: %w", err)
	}

	var private privateKey
	if err := s.vault.Open(sealed, &private); err != nil {
		return nil, fmt.Errorf("decrypt private key: %w", err)
	}
	return private.signer()
}

func labelOr(label, fallback string) string {
	if label = strings.TrimSpace(label); label != "" {
		return label
	}
	return fallback
}
