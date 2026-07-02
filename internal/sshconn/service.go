package sshconn

import (
	"context"
	"database/sql"
	"fmt"

	appcrypto "github.com/Velarvo/velarvo-desktop/internal/crypto"
	"github.com/Velarvo/velarvo-desktop/internal/storage"
	"github.com/google/uuid"
)

const sortOrderStep = 100

type Service struct {
	db        *storage.DB
	repo      *Repository
	cipher    appcrypto.EnvelopeCipher
	connector *Connector
	terminal  *Terminal
}

func NewService(db *storage.DB, cipher appcrypto.EnvelopeCipher, dataDir string) *Service {
	connector := NewConnector(dataDir)
	return &Service{
		db:        db,
		repo:      NewRepository(db),
		cipher:    cipher,
		connector: connector,
		terminal:  NewTerminal(connector),
	}
}

func (s *Service) List(ctx context.Context, workspaceID string) ([]ConnectionDTO, error) {
	if err := s.requireUnlocked(); err != nil {
		return nil, err
	}
	if workspaceID == "" {
		return nil, ErrWorkspaceRequired
	}

	rows, err := s.repo.List(ctx, workspaceID)
	if err != nil {
		return nil, err
	}

	out := make([]ConnectionDTO, 0, len(rows))
	for _, row := range rows {
		dto, err := s.toDTO(row)
		if err != nil {
			return nil, err
		}
		out = append(out, dto)
	}
	return out, nil
}

func (s *Service) Get(ctx context.Context, id string) (*ConnectionDTO, error) {
	if err := s.requireUnlocked(); err != nil {
		return nil, err
	}

	row, err := s.repo.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	dto, err := s.toDTO(*row)
	if err != nil {
		return nil, err
	}
	return &dto, nil
}

func (s *Service) Create(ctx context.Context, req CreateConnectionRequest) (*ConnectionDTO, error) {
	if err := s.requireUnlocked(); err != nil {
		return nil, err
	}
	if req.WorkspaceID == "" {
		return nil, ErrWorkspaceRequired
	}

	exists, err := s.repo.WorkspaceExists(ctx, req.WorkspaceID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrWorkspaceNotFound
	}

	listing := ListingData{
		Name:        req.Name,
		Host:        req.Host,
		Port:        req.Port,
		Username:    req.Username,
		HasPassword: req.Password != "",
	}

	listingEnc, listingVer, err := s.encryptListing(listing)
	if err != nil {
		return nil, err
	}

	secretEnc, secretVer, err := s.encryptSecret(SecretData{Password: req.Password})
	if err != nil {
		return nil, err
	}

	id, err := newUUIDV7()
	if err != nil {
		return nil, err
	}

	maxSort, err := s.repo.MaxSortOrder(ctx, req.WorkspaceID)
	if err != nil {
		return nil, err
	}

	now := storage.Now()
	conn := storage.SSHConnection{
		ID:               id,
		WorkspaceID:      req.WorkspaceID,
		ListingEncrypted: listingEnc,
		ListingVersion:   listingVer,
		SecretEncrypted:  secretEnc,
		SecretVersion:    secretVer,
		SortOrder:        maxSort + sortOrderStep,
		CreatedAt:        now,
		UpdatedAt:        now,
		Revision:         uuid.NewString(),
	}

	if err := s.db.Tx(ctx, func(tx *sql.Tx) error {
		return s.repo.Insert(ctx, tx, conn)
	}); err != nil {
		return nil, err
	}

	dto, err := s.toDTO(conn)
	if err != nil {
		return nil, err
	}
	return &dto, nil
}

func (s *Service) Update(ctx context.Context, req UpdateConnectionRequest) (*ConnectionDTO, error) {
	if err := s.requireUnlocked(); err != nil {
		return nil, err
	}

	row, err := s.repo.Get(ctx, req.ID)
	if err != nil {
		return nil, err
	}

	prev, err := s.decryptListing(*row)
	if err != nil {
		return nil, err
	}

	keepSecret := !req.ClearPassword && req.Password == ""

	hasPassword := prev.HasPassword
	secretEnc := row.SecretEncrypted
	secretVer := row.SecretVersion

	if !keepSecret {
		password := ""
		if !req.ClearPassword {
			password = req.Password
		}
		hasPassword = password != ""
		secretEnc, secretVer, err = s.encryptSecret(SecretData{Password: password})
		if err != nil {
			return nil, err
		}
	}

	listing := ListingData{
		Name:        req.Name,
		Host:        req.Host,
		Port:        req.Port,
		Username:    req.Username,
		HasPassword: hasPassword,
		OS:          prev.OS,
	}

	listingEnc, listingVer, err := s.encryptListing(listing)
	if err != nil {
		return nil, err
	}

	now := storage.Now()
	row.ListingEncrypted = listingEnc
	row.ListingVersion = listingVer
	row.SecretEncrypted = secretEnc
	row.SecretVersion = secretVer
	row.UpdatedAt = now
	row.Revision = uuid.NewString()

	if err := s.db.Tx(ctx, func(tx *sql.Tx) error {
		return s.repo.Update(ctx, tx, *row)
	}); err != nil {
		return nil, err
	}

	dto, err := s.toDTO(*row)
	if err != nil {
		return nil, err
	}
	return &dto, nil
}

func (s *Service) Delete(ctx context.Context, id string) error {
	if err := s.requireUnlocked(); err != nil {
		return err
	}

	s.terminal.CloseForConnection(id)
	s.connector.Disconnect(id)

	now := storage.Now()
	return s.db.Tx(ctx, func(tx *sql.Tx) error {
		return s.repo.SoftDelete(ctx, tx, id, now, uuid.NewString())
	})
}

func (s *Service) Connect(ctx context.Context, id string) (*ConnectionState, error) {
	if err := s.requireUnlocked(); err != nil {
		return nil, err
	}

	row, err := s.repo.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	listing, err := s.decryptListing(*row)
	if err != nil {
		return nil, err
	}

	password, err := s.decryptPassword(*row)
	if err != nil {
		return nil, err
	}
	if password == "" {
		return nil, ErrNoPassword
	}

	if err := s.connector.Connect(id, *listing, password); err != nil {
		return nil, err
	}

	if err := s.repo.TouchLastUsed(ctx, id, storage.Now()); err != nil {
		_ = err
	}

	if client, ok := s.connector.Client(id); ok {
		if detected := DetectOS(client); detected != "" && detected != listing.OS {
			listing.OS = detected
			if err := s.persistListing(ctx, row, *listing); err != nil {
				_ = err
			}
		}
	}

	return &ConnectionState{ID: id, Connected: true, OS: listing.OS}, nil
}

func (s *Service) persistListing(ctx context.Context, row *storage.SSHConnection, listing ListingData) error {
	listingEnc, listingVer, err := s.encryptListing(listing)
	if err != nil {
		return err
	}

	row.ListingEncrypted = listingEnc
	row.ListingVersion = listingVer
	row.UpdatedAt = storage.Now()
	row.Revision = uuid.NewString()

	return s.db.Tx(ctx, func(tx *sql.Tx) error {
		return s.repo.Update(ctx, tx, *row)
	})
}

func (s *Service) Disconnect(_ context.Context, id string) (*ConnectionState, error) {
	s.terminal.CloseForConnection(id)
	s.connector.Disconnect(id)
	return &ConnectionState{ID: id, Connected: false}, nil
}

func (s *Service) ConnectedIDs() []string {
	return s.connector.ConnectedIDs()
}

func (s *Service) OpenTerminal(connID string, cols, rows int, cb TerminalCallbacks) (string, error) {
	return s.terminal.Open(connID, cols, rows, cb)
}

func (s *Service) WriteTerminal(sessionID string, data []byte) error {
	return s.terminal.Write(sessionID, data)
}

func (s *Service) ResizeTerminal(sessionID string, cols, rows int) error {
	return s.terminal.Resize(sessionID, cols, rows)
}

func (s *Service) CloseTerminal(sessionID string) {
	s.terminal.Close(sessionID)
}

func (s *Service) Shutdown() {
	if s.terminal != nil {
		s.terminal.CloseAll()
	}
	if s.connector != nil {
		s.connector.CloseAll()
	}
}

func (s *Service) encryptListing(listing ListingData) ([]byte, storage.DataVersion, error) {
	plaintext, version, err := MarshalListing(listing)
	if err != nil {
		return nil, 0, err
	}
	defer plaintext.Destroy()

	enc, err := s.cipher.Encrypt(plaintext.Bytes())
	if err != nil {
		return nil, 0, fmt.Errorf("encrypt ssh listing: %w", err)
	}
	return enc, version, nil
}

func (s *Service) encryptSecret(secret SecretData) ([]byte, storage.DataVersion, error) {
	plaintext, version, err := MarshalSecret(secret)
	if err != nil {
		return nil, 0, err
	}
	defer plaintext.Destroy()

	enc, err := s.cipher.Encrypt(plaintext.Bytes())
	if err != nil {
		return nil, 0, fmt.Errorf("encrypt ssh secret: %w", err)
	}
	return enc, version, nil
}

func (s *Service) decryptListing(row storage.SSHConnection) (*ListingData, error) {
	plaintext, err := s.cipher.Decrypt(row.ListingEncrypted)
	if err != nil {
		return nil, fmt.Errorf("decrypt ssh listing: %w", err)
	}
	defer plaintext.Destroy()

	return UnmarshalListing(plaintext.Bytes(), row.ListingVersion)
}

func (s *Service) decryptPassword(row storage.SSHConnection) (string, error) {
	plaintext, err := s.cipher.Decrypt(row.SecretEncrypted)
	if err != nil {
		return "", fmt.Errorf("decrypt ssh secret: %w", err)
	}
	defer plaintext.Destroy()

	secret, err := UnmarshalSecret(plaintext.Bytes(), row.SecretVersion)
	if err != nil {
		return "", err
	}
	return secret.Password, nil
}

func (s *Service) toDTO(row storage.SSHConnection) (ConnectionDTO, error) {
	listing, err := s.decryptListing(row)
	if err != nil {
		return ConnectionDTO{}, err
	}

	return ConnectionDTO{
		ID:          row.ID,
		WorkspaceID: row.WorkspaceID,
		Name:        listing.Name,
		Host:        listing.Host,
		Port:        listing.Port,
		Username:    listing.Username,
		HasPassword: listing.HasPassword,
		OS:          listing.OS,
		SortOrder:   row.SortOrder,
		LastUsedAt:  row.LastUsedAt,
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
		Revision:    row.Revision,
		Connected:   s.connector.IsConnected(row.ID),
	}, nil
}

func (s *Service) requireUnlocked() error {
	if s.cipher == nil {
		return ErrCipherNotReady
	}
	if !s.cipher.IsUnlocked() {
		return ErrLocked
	}
	return nil
}

func newUUIDV7() (string, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return "", fmt.Errorf("generate uuid v7: %w", err)
	}
	return id.String(), nil
}
