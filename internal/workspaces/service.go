package workspaces

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	appcrypto "github.com/Velarvo/velarvo-desktop/internal/crypto"
	"github.com/Velarvo/velarvo-desktop/internal/storage"
	"github.com/google/uuid"
)

const sortOrderStep = 100

type Service struct {
	db     *storage.DB
	repo   *Repository
	cipher appcrypto.EnvelopeCipher
}

func NewService(db *storage.DB, cipher appcrypto.EnvelopeCipher) *Service {
	return &Service{
		db:     db,
		repo:   NewRepository(db),
		cipher: cipher,
	}
}

func (s *Service) ListByProject(ctx context.Context, projectID string) ([]WorkspaceDTO, error) {
	if err := s.requireUnlocked(); err != nil {
		return nil, err
	}

	projectID = strings.TrimSpace(projectID)
	if projectID == "" {
		return nil, fmt.Errorf("%w: workspace projectId is required", ErrInvalidInput)
	}

	rows, err := s.repo.ListByProject(ctx, projectID)
	if err != nil {
		return nil, err
	}

	out := make([]WorkspaceDTO, 0, len(rows))
	for _, row := range rows {
		dto, err := s.toDTO(row)
		if err != nil {
			return nil, err
		}
		out = append(out, dto)
	}
	return out, nil
}

func (s *Service) Get(ctx context.Context, id string) (*WorkspaceDTO, error) {
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

func (s *Service) Create(ctx context.Context, req CreateWorkspaceRequest) (*WorkspaceDTO, error) {
	if err := s.requireUnlocked(); err != nil {
		return nil, err
	}

	projectID := strings.TrimSpace(req.ProjectID)
	if projectID == "" {
		return nil, fmt.Errorf("%w: workspace projectId is required", ErrInvalidInput)
	}

	exists, err := s.repo.ProjectExists(ctx, projectID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrProjectMissing
	}

	data := WorkspaceData{
		Name:  normalizeName(req.Name),
		Color: strings.TrimSpace(req.Color),
	}

	if err := s.ensureUniqueName(ctx, projectID, "", data.Name); err != nil {
		return nil, err
	}

	plaintext, dataVersion, err := MarshalData(data)
	if err != nil {
		return nil, err
	}
	defer plaintext.Destroy()

	encrypted, err := s.cipher.Encrypt(plaintext.Bytes())
	if err != nil {
		return nil, fmt.Errorf("encrypt workspace: %w", err)
	}

	id, err := newUUIDV7()
	if err != nil {
		return nil, err
	}

	maxSort, err := s.repo.MaxSortOrder(ctx, projectID)
	if err != nil {
		return nil, err
	}

	now := storage.Now()
	revision := uuid.NewString()
	workspace := storage.Workspace{
		ID:            id,
		ProjectID:     projectID,
		EncryptedData: encrypted,
		DataVersion:   dataVersion,
		SortOrder:     maxSort + sortOrderStep,
		CreatedAt:     now,
		UpdatedAt:     now,
		Revision:      revision,
	}

	if err := s.db.Tx(ctx, func(tx *sql.Tx) error {
		if err := s.repo.Insert(ctx, tx, workspace); err != nil {
			return err
		}
		return s.repo.InsertOutbox(ctx, tx, workspace.ID, storage.SyncOperationCreate, revision, now)
	}); err != nil {
		return nil, err
	}

	dto, err := s.toDTO(workspace)
	if err != nil {
		return nil, err
	}
	return &dto, nil
}

func (s *Service) Update(ctx context.Context, req UpdateWorkspaceRequest) (*WorkspaceDTO, error) {
	if err := s.requireUnlocked(); err != nil {
		return nil, err
	}

	row, err := s.repo.Get(ctx, req.ID)
	if err != nil {
		return nil, err
	}

	data := WorkspaceData{
		Name:  normalizeName(req.Name),
		Color: strings.TrimSpace(req.Color),
	}

	if err := s.ensureUniqueName(ctx, row.ProjectID, req.ID, data.Name); err != nil {
		return nil, err
	}

	plaintext, dataVersion, err := MarshalData(data)
	if err != nil {
		return nil, err
	}
	defer plaintext.Destroy()

	encrypted, err := s.cipher.Encrypt(plaintext.Bytes())
	if err != nil {
		return nil, fmt.Errorf("encrypt workspace: %w", err)
	}

	now := storage.Now()
	revision := uuid.NewString()
	if err := s.db.Tx(ctx, func(tx *sql.Tx) error {
		if err := s.repo.Update(ctx, tx, req.ID, encrypted, dataVersion, now, revision); err != nil {
			return err
		}
		return s.repo.InsertOutbox(ctx, tx, req.ID, storage.SyncOperationUpdate, revision, now)
	}); err != nil {
		return nil, err
	}

	row.EncryptedData = encrypted
	row.DataVersion = dataVersion
	row.UpdatedAt = now
	row.Revision = revision

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

	now := storage.Now()
	revision := uuid.NewString()
	return s.db.Tx(ctx, func(tx *sql.Tx) error {
		if err := s.repo.SoftDelete(ctx, tx, id, now, revision); err != nil {
			return err
		}
		return s.repo.InsertOutbox(ctx, tx, id, storage.SyncOperationDelete, revision, now)
	})
}

func (s *Service) toDTO(row storage.Workspace) (WorkspaceDTO, error) {
	plaintext, err := s.cipher.Decrypt(row.EncryptedData)
	if err != nil {
		return WorkspaceDTO{}, fmt.Errorf("decrypt workspace: %w", err)
	}
	defer plaintext.Destroy()

	data, err := UnmarshalData(plaintext.Bytes(), row.DataVersion)
	if err != nil {
		return WorkspaceDTO{}, err
	}

	return WorkspaceDTO{
		ID:        row.ID,
		ProjectID: row.ProjectID,
		Name:      data.Name,
		Color:     data.Color,
		SortOrder: row.SortOrder,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
		Revision:  row.Revision,
	}, nil
}

func (s *Service) ensureUniqueName(ctx context.Context, projectID string, exceptID string, name string) error {
	if name == "" {
		return fmt.Errorf("%w: workspace name is required", ErrInvalidInput)
	}

	workspaces, err := s.ListByProject(ctx, projectID)
	if err != nil {
		return err
	}
	for _, workspace := range workspaces {
		if workspace.ID != exceptID && strings.EqualFold(workspace.Name, name) {
			return ErrDuplicateName
		}
	}
	return nil
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
