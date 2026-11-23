package workspaces

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/Velarvo/velarvo-desktop/internal/apperr"
	"github.com/Velarvo/velarvo-desktop/internal/db"
	"github.com/Velarvo/velarvo-desktop/internal/vault"
	"github.com/google/uuid"
)

var (
	ErrNotFound        = apperr.New("WORKSPACE_NOT_FOUND", "workspace not found")
	ErrProjectNotFound = apperr.New("WORKSPACE_PROJECT_MISSING", "workspace project not found")
	ErrDuplicateName   = apperr.New("WORKSPACE_DUPLICATE_NAME", "workspace name already exists in the project")
	ErrNameRequired    = apperr.New("WORKSPACE_INVALID_INPUT", "workspace name is required")
)

type Workspace struct {
	ID        string `json:"id"`
	ProjectID string `json:"projectId"`
	Name      string `json:"name"`
	Color     string `json:"color"`
	CreatedAt int64  `json:"createdAt"`
	UpdatedAt int64  `json:"updatedAt"`
}

type WorkspaceInput struct {
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

const selectWorkspace = `SELECT id, project_id, data, created_at, updated_at FROM workspaces`

func (s *Service) List(ctx context.Context, projectID string) ([]Workspace, error) {
	if err := s.vault.Unlocked(); err != nil {
		return nil, err
	}

	rows, err := s.db.QueryContext(ctx, selectWorkspace+` WHERE project_id = ? ORDER BY created_at`, projectID)
	if err != nil {
		return nil, fmt.Errorf("list workspaces: %w", err)
	}
	defer func() { _ = rows.Close() }()

	workspaces := []Workspace{}
	for rows.Next() {
		workspace, err := s.scan(rows)
		if err != nil {
			return nil, err
		}
		workspaces = append(workspaces, workspace)
	}
	return workspaces, rows.Err()
}

func (s *Service) Get(ctx context.Context, id string) (Workspace, error) {
	if err := s.vault.Unlocked(); err != nil {
		return Workspace{}, err
	}
	workspace, err := s.scan(s.db.QueryRowContext(ctx, selectWorkspace+` WHERE id = ?`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return Workspace{}, ErrNotFound
	}
	return workspace, err
}

func (s *Service) Create(ctx context.Context, projectID string, input WorkspaceInput) (Workspace, error) {
	var exists bool
	if err := s.db.QueryRowContext(ctx,
		`SELECT EXISTS (SELECT 1 FROM projects WHERE id = ?)`, projectID).Scan(&exists); err != nil {
		return Workspace{}, fmt.Errorf("check project: %w", err)
	}
	if !exists {
		return Workspace{}, ErrProjectNotFound
	}

	f, err := s.validate(ctx, projectID, "", input)
	if err != nil {
		return Workspace{}, err
	}
	data, err := s.vault.Seal(f)
	if err != nil {
		return Workspace{}, fmt.Errorf("encrypt workspace: %w", err)
	}

	now := db.Now()
	workspace := Workspace{
		ID:        uuid.Must(uuid.NewV7()).String(),
		ProjectID: projectID,
		Name:      f.Name,
		Color:     f.Color,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if _, err := s.db.ExecContext(ctx,
		`INSERT INTO workspaces (id, project_id, data, created_at, updated_at) VALUES (?, ?, ?, ?, ?)`,
		workspace.ID, projectID, data, now, now,
	); err != nil {
		return Workspace{}, fmt.Errorf("insert workspace: %w", err)
	}
	return workspace, nil
}

func (s *Service) Update(ctx context.Context, id string, input WorkspaceInput) (Workspace, error) {
	workspace, err := s.Get(ctx, id)
	if err != nil {
		return Workspace{}, err
	}
	f, err := s.validate(ctx, workspace.ProjectID, id, input)
	if err != nil {
		return Workspace{}, err
	}
	data, err := s.vault.Seal(f)
	if err != nil {
		return Workspace{}, fmt.Errorf("encrypt workspace: %w", err)
	}

	workspace.Name, workspace.Color, workspace.UpdatedAt = f.Name, f.Color, db.Now()
	result, err := s.db.ExecContext(ctx,
		`UPDATE workspaces SET data = ?, updated_at = ? WHERE id = ?`, data, workspace.UpdatedAt, id)
	if err != nil {
		return Workspace{}, fmt.Errorf("update workspace: %w", err)
	}
	return workspace, db.Affected(result, ErrNotFound)
}

func (s *Service) Delete(ctx context.Context, id string) error {
	if err := s.vault.Unlocked(); err != nil {
		return err
	}
	result, err := s.db.ExecContext(ctx, `DELETE FROM workspaces WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete workspace: %w", err)
	}
	return db.Affected(result, ErrNotFound)
}

func (s *Service) validate(ctx context.Context, projectID, exceptID string, input WorkspaceInput) (fields, error) {
	f := fields{Name: strings.ToLower(strings.TrimSpace(input.Name)), Color: strings.TrimSpace(input.Color)}
	if f.Name == "" {
		return fields{}, ErrNameRequired
	}

	workspaces, err := s.List(ctx, projectID)
	if err != nil {
		return fields{}, err
	}
	for _, workspace := range workspaces {
		if workspace.ID != exceptID && workspace.Name == f.Name {
			return fields{}, ErrDuplicateName
		}
	}
	return f, nil
}

func (s *Service) scan(row interface{ Scan(...any) error }) (Workspace, error) {
	var workspace Workspace
	var data []byte
	if err := row.Scan(&workspace.ID, &workspace.ProjectID, &data, &workspace.CreatedAt, &workspace.UpdatedAt); err != nil {
		return Workspace{}, err
	}

	var f fields
	if err := s.vault.Open(data, &f); err != nil {
		return Workspace{}, fmt.Errorf("decrypt workspace %s: %w", workspace.ID, err)
	}
	workspace.Name, workspace.Color = f.Name, f.Color
	return workspace, nil
}
