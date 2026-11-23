package projects

import (
	"context"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"github.com/Velarvo/velarvo-desktop/internal/apperr"
	"github.com/Velarvo/velarvo-desktop/internal/db"
	"github.com/Velarvo/velarvo-desktop/internal/vault"
	"github.com/google/uuid"
)

var (
	ErrNotFound      = apperr.New("PROJECT_NOT_FOUND", "project not found")
	ErrDuplicateName = apperr.New("PROJECT_DUPLICATE_NAME", "project name already exists")
	ErrNameRequired  = apperr.New("PROJECT_INVALID_INPUT", "project name is required")
	ErrInvalidIcon   = apperr.New("PROJECT_INVALID_INPUT", "project icon must be a PNG, JPEG or WebP image of at most 256 KiB")
	ErrIconNotFound  = apperr.New("PROJECT_ICON_NOT_FOUND", "project icon not found")
)

const maxIconBytes = 256 << 10

var iconTypes = map[string]bool{"image/png": true, "image/jpeg": true, "image/webp": true}

type Project struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Color     string `json:"color"`
	HasIcon   bool   `json:"hasIcon"`
	CreatedAt int64  `json:"createdAt"`
	UpdatedAt int64  `json:"updatedAt"`
}

type ProjectInput struct {
	Name  string `json:"name"`
	Color string `json:"color"`
}

type fields struct {
	Name  string `json:"name"`
	Color string `json:"color,omitempty"`
}

type Service struct {
	db    *sql.DB
	vault *vault.Vault
}

func NewService(database *sql.DB, v *vault.Vault) *Service {
	return &Service{db: database, vault: v}
}

const selectProject = `
	SELECT p.id, p.data, p.created_at, p.updated_at,
	       EXISTS (SELECT 1 FROM project_icons i WHERE i.project_id = p.id)
	FROM projects p`

func (s *Service) List(ctx context.Context) ([]Project, error) {
	if err := s.vault.Unlocked(); err != nil {
		return nil, err
	}

	rows, err := s.db.QueryContext(ctx, selectProject+` ORDER BY p.created_at`)
	if err != nil {
		return nil, fmt.Errorf("list projects: %w", err)
	}
	defer func() { _ = rows.Close() }()

	projects := []Project{}
	for rows.Next() {
		project, err := s.scan(rows)
		if err != nil {
			return nil, err
		}
		projects = append(projects, project)
	}
	return projects, rows.Err()
}

func (s *Service) Get(ctx context.Context, id string) (Project, error) {
	if err := s.vault.Unlocked(); err != nil {
		return Project{}, err
	}
	project, err := s.scan(s.db.QueryRowContext(ctx, selectProject+` WHERE p.id = ?`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return Project{}, ErrNotFound
	}
	return project, err
}

func (s *Service) Create(ctx context.Context, input ProjectInput) (Project, error) {
	f, err := s.validate(ctx, "", input)
	if err != nil {
		return Project{}, err
	}
	data, err := s.vault.Seal(f)
	if err != nil {
		return Project{}, fmt.Errorf("encrypt project: %w", err)
	}

	now := db.Now()
	project := Project{ID: uuid.Must(uuid.NewV7()).String(), Name: f.Name, Color: f.Color, CreatedAt: now, UpdatedAt: now}
	if _, err := s.db.ExecContext(ctx,
		`INSERT INTO projects (id, data, created_at, updated_at) VALUES (?, ?, ?, ?)`,
		project.ID, data, now, now,
	); err != nil {
		return Project{}, fmt.Errorf("insert project: %w", err)
	}
	return project, nil
}

func (s *Service) Update(ctx context.Context, id string, input ProjectInput) (Project, error) {
	project, err := s.Get(ctx, id)
	if err != nil {
		return Project{}, err
	}
	f, err := s.validate(ctx, id, input)
	if err != nil {
		return Project{}, err
	}
	data, err := s.vault.Seal(f)
	if err != nil {
		return Project{}, fmt.Errorf("encrypt project: %w", err)
	}

	project.Name, project.Color, project.UpdatedAt = f.Name, f.Color, db.Now()
	result, err := s.db.ExecContext(ctx,
		`UPDATE projects SET data = ?, updated_at = ? WHERE id = ?`, data, project.UpdatedAt, id)
	if err != nil {
		return Project{}, fmt.Errorf("update project: %w", err)
	}
	return project, db.Affected(result, ErrNotFound)
}

func (s *Service) Delete(ctx context.Context, id string) error {
	if err := s.vault.Unlocked(); err != nil {
		return err
	}
	result, err := s.db.ExecContext(ctx, `DELETE FROM projects WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete project: %w", err)
	}
	return db.Affected(result, ErrNotFound)
}

func (s *Service) SetIcon(ctx context.Context, projectID, mime string, image []byte) (string, error) {
	if !iconTypes[mime] || len(image) == 0 || len(image) > maxIconBytes {
		return "", ErrInvalidIcon
	}
	if _, err := s.Get(ctx, projectID); err != nil {
		return "", err
	}
	data, err := s.vault.Encrypt(image)
	if err != nil {
		return "", fmt.Errorf("encrypt project icon: %w", err)
	}

	if _, err := s.db.ExecContext(ctx, `
		INSERT INTO project_icons (project_id, mime, data, updated_at) VALUES (?, ?, ?, ?)
		ON CONFLICT (project_id) DO UPDATE SET mime = excluded.mime, data = excluded.data, updated_at = excluded.updated_at`,
		projectID, mime, data, db.Now(),
	); err != nil {
		return "", fmt.Errorf("save project icon: %w", err)
	}
	return dataURL(mime, image), nil
}

func (s *Service) Icon(ctx context.Context, projectID string) (string, error) {
	if err := s.vault.Unlocked(); err != nil {
		return "", err
	}

	var mime string
	var data []byte
	err := s.db.QueryRowContext(ctx,
		`SELECT mime, data FROM project_icons WHERE project_id = ?`, projectID).Scan(&mime, &data)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrIconNotFound
	}
	if err != nil {
		return "", fmt.Errorf("read project icon: %w", err)
	}

	image, err := s.vault.Decrypt(data)
	if err != nil {
		return "", fmt.Errorf("decrypt project icon: %w", err)
	}
	defer image.Destroy()
	return dataURL(mime, image.Bytes()), nil
}

func (s *Service) DeleteIcon(ctx context.Context, projectID string) error {
	if err := s.vault.Unlocked(); err != nil {
		return err
	}
	if _, err := s.db.ExecContext(ctx, `DELETE FROM project_icons WHERE project_id = ?`, projectID); err != nil {
		return fmt.Errorf("delete project icon: %w", err)
	}
	return nil
}

func (s *Service) validate(ctx context.Context, exceptID string, input ProjectInput) (fields, error) {
	f := fields{Name: strings.TrimSpace(input.Name), Color: strings.TrimSpace(input.Color)}
	if f.Name == "" {
		return fields{}, ErrNameRequired
	}

	projects, err := s.List(ctx)
	if err != nil {
		return fields{}, err
	}
	for _, project := range projects {
		if project.ID != exceptID && strings.EqualFold(project.Name, f.Name) {
			return fields{}, ErrDuplicateName
		}
	}
	return f, nil
}

func (s *Service) scan(row interface{ Scan(...any) error }) (Project, error) {
	var project Project
	var data []byte
	if err := row.Scan(&project.ID, &data, &project.CreatedAt, &project.UpdatedAt, &project.HasIcon); err != nil {
		return Project{}, err
	}

	var f fields
	if err := s.vault.Open(data, &f); err != nil {
		return Project{}, fmt.Errorf("decrypt project %s: %w", project.ID, err)
	}
	project.Name, project.Color = f.Name, f.Color
	return project, nil
}

func dataURL(mime string, data []byte) string {
	return "data:" + mime + ";base64," + base64.StdEncoding.EncodeToString(data)
}
