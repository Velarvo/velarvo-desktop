package workspaces

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/Velarvo/velarvo-desktop/internal/storage"
)

type Repository struct {
	db *storage.DB
}

func NewRepository(db *storage.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) ListByProject(ctx context.Context, projectID string) ([]storage.Workspace, error) {
	rows, err := r.db.Read().QueryContext(ctx, `
		SELECT id, project_id, encrypted_data, data_version, sort_order,
		       created_at, updated_at, server_updated_at, deleted_at, revision
		FROM workspaces
		WHERE project_id = ? AND deleted_at IS NULL
		ORDER BY sort_order ASC, created_at ASC`, projectID)
	if err != nil {
		return nil, fmt.Errorf("list workspaces: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var out []storage.Workspace
	for rows.Next() {
		workspace, err := scanWorkspace(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, workspace)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate workspaces: %w", err)
	}
	return out, nil
}

func (r *Repository) Get(ctx context.Context, id string) (*storage.Workspace, error) {
	row := r.db.Read().QueryRowContext(ctx, `
		SELECT id, project_id, encrypted_data, data_version, sort_order,
		       created_at, updated_at, server_updated_at, deleted_at, revision
		FROM workspaces
		WHERE id = ? AND deleted_at IS NULL`, id)

	workspace, err := scanWorkspace(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &workspace, nil
}

func (r *Repository) MaxSortOrder(ctx context.Context, projectID string) (int, error) {
	var max sql.NullInt64
	err := r.db.Read().QueryRowContext(ctx, `
		SELECT MAX(sort_order)
		FROM workspaces
		WHERE project_id = ? AND deleted_at IS NULL`, projectID).Scan(&max)
	if err != nil {
		return 0, fmt.Errorf("read max workspace sort order: %w", err)
	}
	if !max.Valid {
		return 0, nil
	}
	return int(max.Int64), nil
}

func (r *Repository) ProjectExists(ctx context.Context, projectID string) (bool, error) {
	var count int
	err := r.db.Read().QueryRowContext(ctx, `
		SELECT COUNT(1)
		FROM projects
		WHERE id = ? AND deleted_at IS NULL`, projectID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("check project existence: %w", err)
	}
	return count > 0, nil
}

func (r *Repository) Insert(ctx context.Context, tx *sql.Tx, workspace storage.Workspace) error {
	_, err := tx.ExecContext(ctx, `
		INSERT INTO workspaces (
			id, project_id, encrypted_data, data_version, sort_order,
			created_at, updated_at, server_updated_at, deleted_at, revision
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		workspace.ID, workspace.ProjectID, workspace.EncryptedData, workspace.DataVersion, workspace.SortOrder,
		workspace.CreatedAt, workspace.UpdatedAt, storage.NullableTimestamp(workspace.ServerUpdatedAt),
		storage.NullableTimestamp(workspace.DeletedAt), workspace.Revision,
	)
	if err != nil {
		return fmt.Errorf("insert workspace: %w", err)
	}
	return nil
}

func (r *Repository) Update(ctx context.Context, tx *sql.Tx, id string, encryptedData []byte, dataVersion storage.DataVersion, updatedAt storage.Timestamp, revision string) error {
	result, err := tx.ExecContext(ctx, `
		UPDATE workspaces
		SET encrypted_data = ?, data_version = ?, updated_at = ?, revision = ?
		WHERE id = ? AND deleted_at IS NULL`,
		encryptedData, dataVersion, updatedAt, revision, id,
	)
	if err != nil {
		return fmt.Errorf("update workspace: %w", err)
	}
	return storage.RequireAffected(result, ErrNotFound)
}

func (r *Repository) SoftDelete(ctx context.Context, tx *sql.Tx, id string, deletedAt storage.Timestamp, revision string) error {
	result, err := tx.ExecContext(ctx, `
		UPDATE workspaces
		SET deleted_at = ?, updated_at = ?, revision = ?
		WHERE id = ? AND deleted_at IS NULL`,
		deletedAt, deletedAt, revision, id,
	)
	if err != nil {
		return fmt.Errorf("soft delete workspace: %w", err)
	}
	return storage.RequireAffected(result, ErrNotFound)
}

func (r *Repository) InsertOutbox(ctx context.Context, tx *sql.Tx, entityID string, operation storage.SyncOperation, revision string, createdAt storage.Timestamp) error {
	_, err := tx.ExecContext(ctx, `
		INSERT INTO sync_outbox (entity_type, entity_id, operation, revision, created_at)
		VALUES (?, ?, ?, ?, ?)`,
		storage.SyncEntityWorkspace, entityID, operation, revision, createdAt,
	)
	if err != nil {
		return fmt.Errorf("insert workspace outbox entry: %w", err)
	}
	return nil
}

type workspaceScanner interface {
	Scan(dest ...any) error
}

func scanWorkspace(scanner workspaceScanner) (storage.Workspace, error) {
	var workspace storage.Workspace
	var dataVersion int
	var createdAt, updatedAt int64
	var serverUpdatedAt, deletedAt sql.NullInt64

	err := scanner.Scan(
		&workspace.ID, &workspace.ProjectID, &workspace.EncryptedData, &dataVersion, &workspace.SortOrder,
		&createdAt, &updatedAt, &serverUpdatedAt, &deletedAt, &workspace.Revision,
	)
	if err != nil {
		return storage.Workspace{}, err
	}

	workspace.DataVersion = storage.DataVersion(dataVersion)
	workspace.CreatedAt = storage.Timestamp(createdAt)
	workspace.UpdatedAt = storage.Timestamp(updatedAt)
	workspace.ServerUpdatedAt = storage.TimestampPtr(serverUpdatedAt)
	workspace.DeletedAt = storage.TimestampPtr(deletedAt)
	return workspace, nil
}
