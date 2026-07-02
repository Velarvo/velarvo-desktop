package vault

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

func (r *Repository) HasVault(ctx context.Context) (bool, error) {
	var exists int
	err := r.db.Read().QueryRowContext(ctx, `SELECT 1 FROM vault_meta WHERE id = 1`).Scan(&exists)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("check vault meta: %w", err)
	}
	return true, nil
}

func (r *Repository) GetMeta(ctx context.Context) (*storage.VaultMeta, error) {
	row := r.db.Read().QueryRowContext(ctx, `
		SELECT schema_version, crypto_version, created_at, kdf_id, kdf_time,
		       kdf_memory, kdf_threads, kdf_keylen, salt_kek, salt_auth,
		       device_id, auto_lock_seconds
		FROM vault_meta
		WHERE id = 1`)

	meta, err := scanVaultMeta(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotSetup
	}
	if err != nil {
		return nil, err
	}
	return &meta, nil
}

func (r *Repository) GetEnvelope(ctx context.Context, method storage.DEKEnvelopeMethod) (*storage.DEKEnvelope, error) {
	row := r.db.Read().QueryRowContext(ctx, `
		SELECT method, envelope, metadata_json, created_at, updated_at
		FROM dek_envelopes
		WHERE method = ?`, method)

	envelope, err := scanDEKEnvelope(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotSetup
	}
	if err != nil {
		return nil, err
	}
	return &envelope, nil
}

func (r *Repository) InsertInitialVault(ctx context.Context, meta storage.VaultMeta, envelope storage.DEKEnvelope) error {
	return r.db.Tx(ctx, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx, `
			INSERT INTO vault_meta (
				id, schema_version, crypto_version, created_at, kdf_id, kdf_time,
				kdf_memory, kdf_threads, kdf_keylen, salt_kek, salt_auth,
				device_id, auto_lock_seconds
			) VALUES (1, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			meta.SchemaVersion, meta.CryptoVersion, meta.CreatedAt, meta.KDFID, meta.KDFTime,
			meta.KDFMemory, meta.KDFThreads, meta.KDFKeyLen, meta.SaltKEK, meta.SaltAuth,
			meta.DeviceID, meta.AutoLockSeconds,
		)
		if err != nil {
			return fmt.Errorf("insert vault meta: %w", err)
		}

		_, err = tx.ExecContext(ctx, `
			INSERT INTO dek_envelopes (method, envelope, metadata_json, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?)`,
			envelope.Method, envelope.Envelope, nullableString(envelope.MetadataJSON),
			envelope.CreatedAt, envelope.UpdatedAt,
		)
		if err != nil {
			return fmt.Errorf("insert password envelope: %w", err)
		}

		return nil
	})
}

func (r *Repository) UpsertEnvelope(ctx context.Context, envelope storage.DEKEnvelope) error {
	_, err := r.db.Write().ExecContext(ctx, `
		INSERT INTO dek_envelopes (method, envelope, metadata_json, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(method) DO UPDATE SET
			envelope = excluded.envelope,
			metadata_json = excluded.metadata_json,
			updated_at = excluded.updated_at`,
		envelope.Method, envelope.Envelope, nullableString(envelope.MetadataJSON),
		envelope.CreatedAt, envelope.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("upsert dek envelope: %w", err)
	}
	return nil
}

type scanner interface {
	Scan(dest ...any) error
}

func scanVaultMeta(row scanner) (storage.VaultMeta, error) {
	var meta storage.VaultMeta
	var schemaVersion, cryptoVersion, kdfID int
	var createdAt int64

	err := row.Scan(
		&schemaVersion, &cryptoVersion, &createdAt, &kdfID, &meta.KDFTime,
		&meta.KDFMemory, &meta.KDFThreads, &meta.KDFKeyLen, &meta.SaltKEK, &meta.SaltAuth,
		&meta.DeviceID, &meta.AutoLockSeconds,
	)
	if err != nil {
		return storage.VaultMeta{}, err
	}

	meta.SchemaVersion = storage.SchemaVersion(schemaVersion)
	meta.CryptoVersion = storage.CryptoVersion(cryptoVersion)
	meta.CreatedAt = storage.Timestamp(createdAt)
	meta.KDFID = storage.KDFID(kdfID)
	return meta, nil
}

func scanDEKEnvelope(row scanner) (storage.DEKEnvelope, error) {
	var envelope storage.DEKEnvelope
	var metadata sql.NullString
	var createdAt, updatedAt int64

	err := row.Scan(&envelope.Method, &envelope.Envelope, &metadata, &createdAt, &updatedAt)
	if err != nil {
		return storage.DEKEnvelope{}, err
	}

	if metadata.Valid {
		envelope.MetadataJSON = &metadata.String
	}
	envelope.CreatedAt = storage.Timestamp(createdAt)
	envelope.UpdatedAt = storage.Timestamp(updatedAt)
	return envelope, nil
}

func nullableString(value *string) any {
	if value == nil {
		return nil
	}
	return *value
}
