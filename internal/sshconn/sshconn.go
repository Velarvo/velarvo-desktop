package sshconn

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/Velarvo/velarvo-desktop/internal/apperr"
	"github.com/Velarvo/velarvo-desktop/internal/db"
	"github.com/Velarvo/velarvo-desktop/internal/vault"
	"github.com/google/uuid"
	"golang.org/x/crypto/ssh"
)

var (
	ErrNotFound          = apperr.New("SSH_NOT_FOUND", "ssh connection not found")
	ErrWorkspaceNotFound = apperr.New("WORKSPACE_NOT_FOUND", "workspace not found")
	ErrHostRequired      = apperr.New("SSH_INVALID_INPUT", "host is required")
	ErrUsernameRequired  = apperr.New("SSH_INVALID_INPUT", "username is required")
	ErrNoAuthMethod      = apperr.New("SSH_NO_AUTH_METHOD", "connection needs a password or a key")
	ErrConnectFailed     = apperr.New("SSH_CONNECT_FAILED", "ssh connection failed")
	ErrNotConnected      = apperr.New("SSH_NOT_CONNECTED", "ssh connection is not connected")
)

const defaultPort = 22

type Connection struct {
	ID             string `json:"id"`
	WorkspaceID    string `json:"workspaceId"`
	Name           string `json:"name"`
	Host           string `json:"host"`
	Port           int    `json:"port"`
	Username       string `json:"username"`
	HasPassword    bool   `json:"hasPassword"`
	KeyID          string `json:"keyId,omitempty"`
	KeyLabel       string `json:"keyLabel,omitempty"`
	KeyFingerprint string `json:"keyFingerprint,omitempty"`
	OS             string `json:"os"`
	Connected      bool   `json:"connected"`
	LastUsedAt     int64  `json:"lastUsedAt,omitempty"`
	CreatedAt      int64  `json:"createdAt"`
	UpdatedAt      int64  `json:"updatedAt"`
}

type ConnectionInput struct {
	Name     string `json:"name"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`

	Password      string `json:"password"`
	ClearPassword bool   `json:"clearPassword"`
	KeyID         string `json:"keyId"`
}

type fields struct {
	Name     string `json:"name"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	OS       string `json:"os,omitempty"`
}

type row struct {
	Connection
	data     []byte
	password []byte
}

type Service struct {
	db      *sql.DB
	vault   *vault.Vault
	clients *Clients
}

func NewService(database *sql.DB, v *vault.Vault, clients *Clients) *Service {
	return &Service{db: database, vault: v, clients: clients}
}

func (s *Service) ServiceShutdown() error {
	s.clients.closeAll()
	return nil
}

const selectConnection = `
	SELECT c.id, c.workspace_id, c.data, c.password,
	       COALESCE(c.key_id, ''), COALESCE(k.label, ''), COALESCE(k.fingerprint, ''),
	       COALESCE(c.last_used_at, 0), c.created_at, c.updated_at
	FROM ssh_connections c
	LEFT JOIN ssh_keys k ON k.id = c.key_id`

func (s *Service) List(ctx context.Context, workspaceID string) ([]Connection, error) {
	if err := s.vault.Unlocked(); err != nil {
		return nil, err
	}

	rows, err := s.db.QueryContext(ctx, selectConnection+` WHERE c.workspace_id = ? ORDER BY c.created_at`, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("list ssh connections: %w", err)
	}
	defer func() { _ = rows.Close() }()

	connections := []Connection{}
	for rows.Next() {
		r, err := s.scan(rows)
		if err != nil {
			return nil, err
		}
		connections = append(connections, r.Connection)
	}
	return connections, rows.Err()
}

func (s *Service) Create(ctx context.Context, workspaceID string, input ConnectionInput) (Connection, error) {
	if err := s.vault.Unlocked(); err != nil {
		return Connection{}, err
	}
	var exists bool
	if err := s.db.QueryRowContext(ctx,
		`SELECT EXISTS (SELECT 1 FROM workspaces WHERE id = ?)`, workspaceID).Scan(&exists); err != nil {
		return Connection{}, fmt.Errorf("check workspace: %w", err)
	}
	if !exists {
		return Connection{}, ErrWorkspaceNotFound
	}

	f, err := s.validate(ctx, workspaceID, input, input.Password != "")
	if err != nil {
		return Connection{}, err
	}
	data, err := s.vault.Seal(f)
	if err != nil {
		return Connection{}, fmt.Errorf("encrypt ssh connection: %w", err)
	}
	password, err := s.sealPassword(input.Password)
	if err != nil {
		return Connection{}, err
	}

	id := uuid.Must(uuid.NewV7()).String()
	now := db.Now()
	if _, err := s.db.ExecContext(ctx, `
		INSERT INTO ssh_connections (id, workspace_id, key_id, data, password, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		id, workspaceID, nullable(input.KeyID), data, nullBlob(password), now, now,
	); err != nil {
		return Connection{}, fmt.Errorf("insert ssh connection: %w", err)
	}
	return s.get(ctx, id)
}

func (s *Service) Update(ctx context.Context, id string, input ConnectionInput) (Connection, error) {
	current, err := s.getRow(ctx, id)
	if err != nil {
		return Connection{}, err
	}

	password := current.password
	if input.ClearPassword {
		password = nil
	} else if input.Password != "" {
		if password, err = s.sealPassword(input.Password); err != nil {
			return Connection{}, err
		}
	}

	f, err := s.validate(ctx, current.WorkspaceID, input, password != nil)
	if err != nil {
		return Connection{}, err
	}
	f.OS = current.OS
	data, err := s.vault.Seal(f)
	if err != nil {
		return Connection{}, fmt.Errorf("encrypt ssh connection: %w", err)
	}

	result, err := s.db.ExecContext(ctx,
		`UPDATE ssh_connections SET key_id = ?, data = ?, password = ?, updated_at = ? WHERE id = ?`,
		nullable(input.KeyID), data, nullBlob(password), db.Now(), id)
	if err != nil {
		return Connection{}, fmt.Errorf("update ssh connection: %w", err)
	}
	if err := db.Affected(result, ErrNotFound); err != nil {
		return Connection{}, err
	}
	return s.get(ctx, id)
}

func (s *Service) Delete(ctx context.Context, id string) error {
	if err := s.vault.Unlocked(); err != nil {
		return err
	}
	s.clients.close(id)

	result, err := s.db.ExecContext(ctx, `DELETE FROM ssh_connections WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete ssh connection: %w", err)
	}
	return db.Affected(result, ErrNotFound)
}

func (s *Service) Connect(ctx context.Context, id string) (Connection, error) {
	r, err := s.getRow(ctx, id)
	if err != nil {
		return Connection{}, err
	}
	auth, err := s.authMethods(ctx, r)
	if err != nil {
		return Connection{}, err
	}
	client, err := s.clients.dial(id, r.Host, r.Port, r.Username, auth)
	if err != nil {
		return Connection{}, err
	}

	now := db.Now()
	if _, err := s.db.ExecContext(ctx, `UPDATE ssh_connections SET last_used_at = ? WHERE id = ?`, now, id); err != nil {
		slog.Warn("record ssh connection use", "id", id, "error", err)
	}
	if r.KeyID != "" {
		if _, err := s.db.ExecContext(ctx, `UPDATE ssh_keys SET last_used_at = ? WHERE id = ?`, now, r.KeyID); err != nil {
			slog.Warn("record ssh key use", "id", r.KeyID, "error", err)
		}
	}
	if detected := detectOS(client); detected != "" && detected != r.OS {
		if err := s.saveOS(ctx, r, detected); err != nil {
			slog.Warn("store detected os", "id", id, "error", err)
		}
	}

	return s.get(ctx, id)
}

func (s *Service) Disconnect(id string) {
	s.clients.close(id)
}

func (s *Service) authMethods(ctx context.Context, r row) ([]ssh.AuthMethod, error) {
	var methods []ssh.AuthMethod

	if r.KeyID != "" {
		signer, err := s.signer(ctx, r.KeyID)
		if err != nil {
			return nil, err
		}
		methods = append(methods, ssh.PublicKeys(signer))
	}

	if r.password != nil {
		password, err := s.vault.Decrypt(r.password)
		if err != nil {
			return nil, fmt.Errorf("decrypt ssh password: %w", err)
		}

		methods = append(methods, ssh.Password(string(password.Bytes())))
		password.Destroy()
	}

	if len(methods) == 0 {
		return nil, ErrNoAuthMethod
	}
	return methods, nil
}

func (s *Service) validate(ctx context.Context, workspaceID string, input ConnectionInput, hasPassword bool) (fields, error) {
	f := fields{
		Name:     strings.TrimSpace(input.Name),
		Host:     strings.TrimSpace(input.Host),
		Port:     input.Port,
		Username: strings.TrimSpace(input.Username),
	}
	switch {
	case f.Host == "":
		return fields{}, ErrHostRequired
	case f.Username == "":
		return fields{}, ErrUsernameRequired
	case !hasPassword && input.KeyID == "":
		return fields{}, ErrNoAuthMethod
	}
	if f.Port <= 0 || f.Port > 65535 {
		f.Port = defaultPort
	}
	if f.Name == "" {
		f.Name = f.Host
	}

	if input.KeyID != "" {
		var exists bool
		if err := s.db.QueryRowContext(ctx,
			`SELECT EXISTS (SELECT 1 FROM ssh_keys WHERE id = ? AND workspace_id = ?)`,
			input.KeyID, workspaceID).Scan(&exists); err != nil {
			return fields{}, fmt.Errorf("check ssh key: %w", err)
		}
		if !exists {
			return fields{}, ErrKeyNotFound
		}
	}
	return f, nil
}

func (s *Service) sealPassword(password string) ([]byte, error) {
	if password == "" {
		return nil, nil
	}
	sealed, err := s.vault.Encrypt([]byte(password))
	if err != nil {
		return nil, fmt.Errorf("encrypt ssh password: %w", err)
	}
	return sealed, nil
}

func (s *Service) saveOS(ctx context.Context, r row, os string) error {
	f := fields{Name: r.Name, Host: r.Host, Port: r.Port, Username: r.Username, OS: os}
	data, err := s.vault.Seal(f)
	if err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx, `UPDATE ssh_connections SET data = ? WHERE id = ?`, data, r.ID)
	return err
}

func (s *Service) get(ctx context.Context, id string) (Connection, error) {
	r, err := s.getRow(ctx, id)
	return r.Connection, err
}

func (s *Service) getRow(ctx context.Context, id string) (row, error) {
	if err := s.vault.Unlocked(); err != nil {
		return row{}, err
	}
	r, err := s.scan(s.db.QueryRowContext(ctx, selectConnection+` WHERE c.id = ?`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return row{}, ErrNotFound
	}
	return r, err
}

func (s *Service) scan(scanner interface{ Scan(...any) error }) (row, error) {
	var r row
	if err := scanner.Scan(&r.ID, &r.WorkspaceID, &r.data, &r.password,
		&r.KeyID, &r.KeyLabel, &r.KeyFingerprint, &r.LastUsedAt, &r.CreatedAt, &r.UpdatedAt); err != nil {
		return row{}, err
	}

	var f fields
	if err := s.vault.Open(r.data, &f); err != nil {
		return row{}, fmt.Errorf("decrypt ssh connection %s: %w", r.ID, err)
	}
	r.Name, r.Host, r.Port, r.Username, r.OS = f.Name, f.Host, f.Port, f.Username, f.OS
	r.HasPassword = r.password != nil
	r.Connected = s.clients.connected(r.ID)
	return r, nil
}

func nullable(id string) any {
	if id == "" {
		return nil
	}
	return id
}

func nullBlob(data []byte) any {
	if data == nil {
		return nil
	}
	return data
}
