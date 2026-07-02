package projects

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

func (r *Repository) List(ctx context.Context) ([]storage.Project, error) {
	rows, err := r.db.Read().QueryContext(ctx, `
		SELECT p.id, p.encrypted_data, p.data_version, p.sort_order,
		       p.created_at, p.updated_at, p.server_updated_at, p.deleted_at, p.revision
		FROM projects p
		WHERE p.deleted_at IS NULL
		ORDER BY p.sort_order ASC, p.created_at ASC`)
	if err != nil {
		return nil, fmt.Errorf("list projects: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var out []storage.Project
	for rows.Next() {
		project, err := scanProject(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, project)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate projects: %w", err)
	}
	return out, nil
}

func (r *Repository) Get(ctx context.Context, id string) (*storage.Project, error) {
	row := r.db.Read().QueryRowContext(ctx, `
		SELECT id, encrypted_data, data_version, sort_order,
		       created_at, updated_at, server_updated_at, deleted_at, revision
		FROM projects
		WHERE id = ? AND deleted_at IS NULL`, id)

	project, err := scanProject(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &project, nil
}

func (r *Repository) HasIcon(ctx context.Context, projectID string) (bool, error) {
	var count int
	err := r.db.Read().QueryRowContext(ctx, `
		SELECT COUNT(1)
		FROM project_icons
		WHERE project_id = ? AND local_status = 'present'`, projectID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("check project icon: %w", err)
	}
	return count > 0, nil
}

func (r *Repository) MaxSortOrder(ctx context.Context) (int, error) {
	var max sql.NullInt64
	err := r.db.Read().QueryRowContext(ctx, `
		SELECT MAX(sort_order)
		FROM projects
		WHERE deleted_at IS NULL`).Scan(&max)
	if err != nil {
		return 0, fmt.Errorf("read max project sort order: %w", err)
	}
	if !max.Valid {
		return 0, nil
	}
	return int(max.Int64), nil
}

func (r *Repository) Insert(ctx context.Context, tx *sql.Tx, project storage.Project) error {
	_, err := tx.ExecContext(ctx, `
		INSERT INTO projects (
			id, encrypted_data, data_version, sort_order,
			created_at, updated_at, server_updated_at, deleted_at, revision
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		project.ID, project.EncryptedData, project.DataVersion, project.SortOrder,
		project.CreatedAt, project.UpdatedAt, storage.NullableTimestamp(project.ServerUpdatedAt),
		storage.NullableTimestamp(project.DeletedAt), project.Revision,
	)
	if err != nil {
		return fmt.Errorf("insert project: %w", err)
	}
	return nil
}

func (r *Repository) Update(ctx context.Context, tx *sql.Tx, id string, encryptedData []byte, dataVersion storage.DataVersion, updatedAt storage.Timestamp, revision string) error {
	result, err := tx.ExecContext(ctx, `
		UPDATE projects
		SET encrypted_data = ?, data_version = ?, updated_at = ?, revision = ?
		WHERE id = ? AND deleted_at IS NULL`,
		encryptedData, dataVersion, updatedAt, revision, id,
	)
	if err != nil {
		return fmt.Errorf("update project: %w", err)
	}
	return storage.RequireAffected(result, ErrNotFound)
}

func (r *Repository) SoftDelete(ctx context.Context, tx *sql.Tx, id string, deletedAt storage.Timestamp, revision string) error {
	result, err := tx.ExecContext(ctx, `
		UPDATE projects
		SET deleted_at = ?, updated_at = ?, revision = ?
		WHERE id = ? AND deleted_at IS NULL`,
		deletedAt, deletedAt, revision, id,
	)
	if err != nil {
		return fmt.Errorf("soft delete project: %w", err)
	}
	return storage.RequireAffected(result, ErrNotFound)
}

func (r *Repository) GetIcon(ctx context.Context, projectID string) (*storage.ProjectIcon, error) {
	row := r.db.Read().QueryRowContext(ctx, `
		SELECT project_id, encrypted_icon, icon_mime, local_status, sync_state,
		       created_at, updated_at, server_updated_at, revision
		FROM project_icons
		WHERE project_id = ?`, projectID)

	icon, err := scanProjectIcon(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrIconNotFound
	}
	if err != nil {
		return nil, err
	}
	return &icon, nil
}

func (r *Repository) UpsertIcon(ctx context.Context, tx *sql.Tx, icon storage.ProjectIcon) error {
	_, err := tx.ExecContext(ctx, `
		INSERT INTO project_icons (
			project_id, encrypted_icon, icon_mime, local_status, sync_state,
			created_at, updated_at, server_updated_at, revision
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(project_id) DO UPDATE SET
			encrypted_icon = excluded.encrypted_icon,
			icon_mime = excluded.icon_mime,
			local_status = excluded.local_status,
			sync_state = excluded.sync_state,
			updated_at = excluded.updated_at,
			server_updated_at = excluded.server_updated_at,
			revision = excluded.revision`,
		icon.ProjectID, icon.EncryptedIcon, icon.IconMIME, icon.LocalStatus, icon.SyncState,
		icon.CreatedAt, icon.UpdatedAt, storage.NullableTimestamp(icon.ServerUpdatedAt), icon.Revision,
	)
	if err != nil {
		return fmt.Errorf("upsert project icon: %w", err)
	}
	return nil
}

func (r *Repository) DeleteIcon(ctx context.Context, tx *sql.Tx, projectID string) error {
	_, err := tx.ExecContext(ctx, `DELETE FROM project_icons WHERE project_id = ?`, projectID)
	if err != nil {
		return fmt.Errorf("delete project icon: %w", err)
	}
	return nil
}

func (r *Repository) InsertOutbox(ctx context.Context, tx *sql.Tx, entityID string, operation storage.SyncOperation, revision string, createdAt storage.Timestamp) error {
	_, err := tx.ExecContext(ctx, `
		INSERT INTO sync_outbox (entity_type, entity_id, operation, revision, created_at)
		VALUES (?, ?, ?, ?, ?)`,
		storage.SyncEntityProject, entityID, operation, revision, createdAt,
	)
	if err != nil {
		return fmt.Errorf("insert project outbox entry: %w", err)
	}
	return nil
}

func (r *Repository) InsertIconOutbox(ctx context.Context, tx *sql.Tx, projectID string, operation storage.SyncOperation, revision string, createdAt storage.Timestamp) error {
	_, err := tx.ExecContext(ctx, `
		INSERT INTO sync_outbox (entity_type, entity_id, operation, revision, created_at)
		VALUES (?, ?, ?, ?, ?)`,
		storage.SyncEntityProjectIcon, projectID, operation, revision, createdAt,
	)
	if err != nil {
		return fmt.Errorf("insert project icon outbox entry: %w", err)
	}
	return nil
}

type projectScanner interface {
	Scan(dest ...any) error
}

func scanProject(scanner projectScanner) (storage.Project, error) {
	var project storage.Project
	var dataVersion int
	var createdAt, updatedAt int64
	var serverUpdatedAt, deletedAt sql.NullInt64

	err := scanner.Scan(
		&project.ID, &project.EncryptedData, &dataVersion, &project.SortOrder,
		&createdAt, &updatedAt, &serverUpdatedAt, &deletedAt, &project.Revision,
	)
	if err != nil {
		return storage.Project{}, err
	}

	project.DataVersion = storage.DataVersion(dataVersion)
	project.CreatedAt = storage.Timestamp(createdAt)
	project.UpdatedAt = storage.Timestamp(updatedAt)
	project.ServerUpdatedAt = storage.TimestampPtr(serverUpdatedAt)
	project.DeletedAt = storage.TimestampPtr(deletedAt)
	return project, nil
}

func scanProjectIcon(scanner projectScanner) (storage.ProjectIcon, error) {
	var icon storage.ProjectIcon
	var createdAt, updatedAt int64
	var serverUpdatedAt sql.NullInt64

	err := scanner.Scan(
		&icon.ProjectID, &icon.EncryptedIcon, &icon.IconMIME, &icon.LocalStatus, &icon.SyncState,
		&createdAt, &updatedAt, &serverUpdatedAt, &icon.Revision,
	)
	if err != nil {
		return storage.ProjectIcon{}, err
	}

	icon.CreatedAt = storage.Timestamp(createdAt)
	icon.UpdatedAt = storage.Timestamp(updatedAt)
	icon.ServerUpdatedAt = storage.TimestampPtr(serverUpdatedAt)
	return icon, nil
}
