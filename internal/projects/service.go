package projects

import (
	"context"
	"database/sql"
	"encoding/base64"
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

func (s *Service) List(ctx context.Context) ([]ProjectDTO, error) {
	if err := s.requireUnlocked(); err != nil {
		return nil, err
	}

	rows, err := s.repo.List(ctx)
	if err != nil {
		return nil, err
	}

	out := make([]ProjectDTO, 0, len(rows))
	for _, row := range rows {
		dto, err := s.toDTO(ctx, row)
		if err != nil {
			return nil, err
		}
		out = append(out, dto)
	}
	return out, nil
}

func (s *Service) Get(ctx context.Context, id string) (*ProjectDTO, error) {
	if err := s.requireUnlocked(); err != nil {
		return nil, err
	}

	row, err := s.repo.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	dto, err := s.toDTO(ctx, *row)
	if err != nil {
		return nil, err
	}
	return &dto, nil
}

func (s *Service) Create(ctx context.Context, req CreateProjectRequest) (*ProjectDTO, error) {
	if err := s.requireUnlocked(); err != nil {
		return nil, err
	}

	data := ProjectData{
		Name:  normalizeName(req.Name),
		Color: strings.TrimSpace(req.Color),
	}

	if err := s.ensureUniqueName(ctx, "", data.Name); err != nil {
		return nil, err
	}

	plaintext, dataVersion, err := MarshalData(data)
	if err != nil {
		return nil, err
	}
	defer plaintext.Destroy()

	encrypted, err := s.cipher.Encrypt(plaintext.Bytes())
	if err != nil {
		return nil, fmt.Errorf("encrypt project: %w", err)
	}

	id, err := newUUIDV7()
	if err != nil {
		return nil, err
	}

	maxSort, err := s.repo.MaxSortOrder(ctx)
	if err != nil {
		return nil, err
	}

	now := storage.Now()
	revision := uuid.NewString()
	project := storage.Project{
		ID:            id,
		EncryptedData: encrypted,
		DataVersion:   dataVersion,
		SortOrder:     maxSort + sortOrderStep,
		CreatedAt:     now,
		UpdatedAt:     now,
		Revision:      revision,
	}

	if err := s.db.Tx(ctx, func(tx *sql.Tx) error {
		if err := s.repo.Insert(ctx, tx, project); err != nil {
			return err
		}
		return s.repo.InsertOutbox(ctx, tx, project.ID, storage.SyncOperationCreate, revision, now)
	}); err != nil {
		return nil, err
	}

	dto, err := s.toDTO(ctx, project)
	if err != nil {
		return nil, err
	}
	return &dto, nil
}

func (s *Service) Update(ctx context.Context, req UpdateProjectRequest) (*ProjectDTO, error) {
	if err := s.requireUnlocked(); err != nil {
		return nil, err
	}

	row, err := s.repo.Get(ctx, req.ID)
	if err != nil {
		return nil, err
	}

	data := ProjectData{
		Name:  normalizeName(req.Name),
		Color: strings.TrimSpace(req.Color),
	}

	if err := s.ensureUniqueName(ctx, req.ID, data.Name); err != nil {
		return nil, err
	}

	plaintext, dataVersion, err := MarshalData(data)
	if err != nil {
		return nil, err
	}
	defer plaintext.Destroy()

	encrypted, err := s.cipher.Encrypt(plaintext.Bytes())
	if err != nil {
		return nil, fmt.Errorf("encrypt project: %w", err)
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

	dto, err := s.toDTO(ctx, *row)
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

func (s *Service) SetIcon(ctx context.Context, req SetProjectIconRequest) (*ProjectIconDTO, error) {
	if err := s.requireUnlocked(); err != nil {
		return nil, err
	}
	if !req.MIME.Valid() {
		return nil, fmt.Errorf("%w: unsupported project icon mime", ErrInvalidInput)
	}
	if len(req.Data) == 0 || len(req.Data) > 256*1024 {
		return nil, fmt.Errorf("%w: project icon must be 1-256 KiB", ErrInvalidInput)
	}

	if _, err := s.repo.Get(ctx, req.ProjectID); err != nil {
		return nil, err
	}

	encrypted, err := s.cipher.Encrypt(req.Data)
	if err != nil {
		return nil, fmt.Errorf("encrypt project icon: %w", err)
	}

	now := storage.Now()
	revision := uuid.NewString()
	icon := storage.ProjectIcon{
		ProjectID:     req.ProjectID,
		EncryptedIcon: encrypted,
		IconMIME:      req.MIME,
		LocalStatus:   storage.LocalStatusPresent,
		SyncState:     storage.SyncItemLocalOnly,
		CreatedAt:     now,
		UpdatedAt:     now,
		Revision:      revision,
	}

	if err := s.db.Tx(ctx, func(tx *sql.Tx) error {
		if err := s.repo.UpsertIcon(ctx, tx, icon); err != nil {
			return err
		}
		return s.repo.InsertIconOutbox(ctx, tx, req.ProjectID, storage.SyncOperationUpdate, revision, now)
	}); err != nil {
		return nil, err
	}

	return s.iconToDTO(icon, req.Data), nil
}

func (s *Service) GetIcon(ctx context.Context, projectID string) (*ProjectIconDTO, error) {
	if err := s.requireUnlocked(); err != nil {
		return nil, err
	}

	icon, err := s.repo.GetIcon(ctx, projectID)
	if err != nil {
		return nil, err
	}
	if icon.LocalStatus == storage.LocalStatusMissing {
		return &ProjectIconDTO{
			ProjectID: projectID,
			Status:    icon.LocalStatus,
			SyncState: icon.SyncState,
			UpdatedAt: icon.UpdatedAt,
		}, nil
	}

	plaintext, err := s.cipher.Decrypt(icon.EncryptedIcon)
	if err != nil {
		return nil, fmt.Errorf("decrypt project icon: %w", err)
	}
	defer plaintext.Destroy()

	return s.iconToDTO(*icon, plaintext.Bytes()), nil
}

func (s *Service) DeleteIcon(ctx context.Context, projectID string) error {
	if err := s.requireUnlocked(); err != nil {
		return err
	}

	now := storage.Now()
	revision := uuid.NewString()
	return s.db.Tx(ctx, func(tx *sql.Tx) error {
		if err := s.repo.DeleteIcon(ctx, tx, projectID); err != nil {
			return err
		}
		return s.repo.InsertIconOutbox(ctx, tx, projectID, storage.SyncOperationDelete, revision, now)
	})
}

func (s *Service) toDTO(ctx context.Context, row storage.Project) (ProjectDTO, error) {
	plaintext, err := s.cipher.Decrypt(row.EncryptedData)
	if err != nil {
		return ProjectDTO{}, fmt.Errorf("decrypt project: %w", err)
	}
	defer plaintext.Destroy()

	data, err := UnmarshalData(plaintext.Bytes(), row.DataVersion)
	if err != nil {
		return ProjectDTO{}, err
	}

	hasIcon, err := s.repo.HasIcon(ctx, row.ID)
	if err != nil {
		return ProjectDTO{}, err
	}

	return ProjectDTO{
		ID:        row.ID,
		Name:      data.Name,
		Color:     data.Color,
		SortOrder: row.SortOrder,
		HasIcon:   hasIcon,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
		Revision:  row.Revision,
	}, nil
}

func (s *Service) iconToDTO(icon storage.ProjectIcon, plaintext []byte) *ProjectIconDTO {
	return &ProjectIconDTO{
		ProjectID:  icon.ProjectID,
		Status:     icon.LocalStatus,
		SyncState:  icon.SyncState,
		MIME:       icon.IconMIME,
		DataBase64: base64.StdEncoding.EncodeToString(plaintext),
		UpdatedAt:  icon.UpdatedAt,
	}
}

func (s *Service) ensureUniqueName(ctx context.Context, exceptID string, name string) error {
	if name == "" {
		return fmt.Errorf("%w: project name is required", ErrInvalidInput)
	}

	projects, err := s.List(ctx)
	if err != nil {
		return err
	}
	for _, project := range projects {
		if project.ID != exceptID && strings.EqualFold(project.Name, name) {
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
