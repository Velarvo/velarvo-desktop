package sshconn

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

const selectColumns = `
	id, workspace_id, listing_encrypted, listing_version, secret_encrypted, secret_version,
	sort_order, last_used_at, created_at, updated_at, server_updated_at, deleted_at, revision`

func (r *Repository) List(ctx context.Context, workspaceID string) ([]storage.SSHConnection, error) {
	rows, err := r.db.Read().QueryContext(ctx, `
		SELECT `+selectColumns+`
		FROM ssh_connections
		WHERE workspace_id = ? AND deleted_at IS NULL
		ORDER BY sort_order ASC, created_at ASC`, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("list ssh connections: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var out []storage.SSHConnection
	for rows.Next() {
		conn, err := scan(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, conn)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate ssh connections: %w", err)
	}
	return out, nil
}

func (r *Repository) Get(ctx context.Context, id string) (*storage.SSHConnection, error) {
	row := r.db.Read().QueryRowContext(ctx, `
		SELECT `+selectColumns+`
		FROM ssh_connections
		WHERE id = ? AND deleted_at IS NULL`, id)

	conn, err := scan(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &conn, nil
}

func (r *Repository) MaxSortOrder(ctx context.Context, workspaceID string) (int, error) {
	var max sql.NullInt64
	err := r.db.Read().QueryRowContext(ctx, `
		SELECT MAX(sort_order)
		FROM ssh_connections
		WHERE workspace_id = ? AND deleted_at IS NULL`, workspaceID).Scan(&max)
	if err != nil {
		return 0, fmt.Errorf("read max ssh sort order: %w", err)
	}
	if !max.Valid {
		return 0, nil
	}
	return int(max.Int64), nil
}

func (r *Repository) WorkspaceExists(ctx context.Context, workspaceID string) (bool, error) {
	var count int
	err := r.db.Read().QueryRowContext(ctx, `
		SELECT COUNT(1)
		FROM workspaces
		WHERE id = ? AND deleted_at IS NULL`, workspaceID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("check workspace existence: %w", err)
	}
	return count > 0, nil
}

func (r *Repository) Insert(ctx context.Context, tx *sql.Tx, conn storage.SSHConnection) error {
	_, err := tx.ExecContext(ctx, `
		INSERT INTO ssh_connections (
			id, workspace_id, listing_encrypted, listing_version, secret_encrypted, secret_version,
			sort_order, last_used_at, created_at, updated_at, server_updated_at, deleted_at, revision
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		conn.ID, conn.WorkspaceID, conn.ListingEncrypted, conn.ListingVersion, conn.SecretEncrypted, conn.SecretVersion,
		conn.SortOrder, storage.NullableTimestamp(conn.LastUsedAt), conn.CreatedAt, conn.UpdatedAt,
		storage.NullableTimestamp(conn.ServerUpdatedAt), storage.NullableTimestamp(conn.DeletedAt), conn.Revision,
	)
	if err != nil {
		return fmt.Errorf("insert ssh connection: %w", err)
	}
	return nil
}

func (r *Repository) Update(ctx context.Context, tx *sql.Tx, conn storage.SSHConnection) error {
	result, err := tx.ExecContext(ctx, `
		UPDATE ssh_connections
		SET listing_encrypted = ?, listing_version = ?, secret_encrypted = ?, secret_version = ?,
		    updated_at = ?, revision = ?
		WHERE id = ? AND deleted_at IS NULL`,
		conn.ListingEncrypted, conn.ListingVersion, conn.SecretEncrypted, conn.SecretVersion,
		conn.UpdatedAt, conn.Revision, conn.ID,
	)
	if err != nil {
		return fmt.Errorf("update ssh connection: %w", err)
	}
	return storage.RequireAffected(result, ErrNotFound)
}

func (r *Repository) TouchLastUsed(ctx context.Context, id string, when storage.Timestamp) error {
	_, err := r.db.Write().ExecContext(ctx, `
		UPDATE ssh_connections
		SET last_used_at = ?
		WHERE id = ? AND deleted_at IS NULL`, when, id)
	if err != nil {
		return fmt.Errorf("touch ssh connection: %w", err)
	}
	return nil
}

func (r *Repository) SoftDelete(ctx context.Context, tx *sql.Tx, id string, deletedAt storage.Timestamp, revision string) error {
	result, err := tx.ExecContext(ctx, `
		UPDATE ssh_connections
		SET deleted_at = ?, updated_at = ?, revision = ?
		WHERE id = ? AND deleted_at IS NULL`,
		deletedAt, deletedAt, revision, id,
	)
	if err != nil {
		return fmt.Errorf("soft delete ssh connection: %w", err)
	}
	return storage.RequireAffected(result, ErrNotFound)
}

type scanner interface {
	Scan(dest ...any) error
}

func scan(s scanner) (storage.SSHConnection, error) {
	var conn storage.SSHConnection
	var listingVersion, secretVersion int
	var createdAt, updatedAt int64
	var lastUsedAt, serverUpdatedAt, deletedAt sql.NullInt64

	err := s.Scan(
		&conn.ID, &conn.WorkspaceID, &conn.ListingEncrypted, &listingVersion, &conn.SecretEncrypted, &secretVersion,
		&conn.SortOrder, &lastUsedAt, &createdAt, &updatedAt, &serverUpdatedAt, &deletedAt, &conn.Revision,
	)
	if err != nil {
		return storage.SSHConnection{}, err
	}

	conn.ListingVersion = storage.DataVersion(listingVersion)
	conn.SecretVersion = storage.DataVersion(secretVersion)
	conn.CreatedAt = storage.Timestamp(createdAt)
	conn.UpdatedAt = storage.Timestamp(updatedAt)
	conn.LastUsedAt = storage.TimestampPtr(lastUsedAt)
	conn.ServerUpdatedAt = storage.TimestampPtr(serverUpdatedAt)
	conn.DeletedAt = storage.TimestampPtr(deletedAt)
	return conn, nil
}
