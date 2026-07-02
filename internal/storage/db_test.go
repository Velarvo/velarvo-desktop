package storage_test

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"

	"github.com/Velarvo/velarvo-desktop/internal/storage"
	"github.com/stretchr/testify/require"
)

func TestOpenRunsInitialMigrations(t *testing.T) {
	t.Parallel()

	db, err := storage.Open(t.TempDir())
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, db.Close())
	})

	require.FileExists(t, filepath.Join(filepath.Dir(db.Path()), "vault.db"))
	assertTableExists(t, db.Read(), "projects")
	assertTableExists(t, db.Read(), "project_icons")
	assertTableExists(t, db.Read(), "workspaces")
	assertTableExists(t, db.Read(), "collections")
	assertTableExists(t, db.Read(), "vault_blocks")
	assertTableExists(t, db.Read(), "vault_meta")
	assertTableExists(t, db.Read(), "dek_envelopes")
	assertTableExists(t, db.Read(), "sync_state")
	assertTableExists(t, db.Read(), "sync_outbox")
	assertTableExists(t, db.Read(), "sync_conflicts")
	assertColumnExists(t, db.Read(), "projects", "encrypted_data")
	assertColumnExists(t, db.Read(), "projects", "data_version")
	assertColumnExists(t, db.Read(), "project_icons", "encrypted_icon")
	assertColumnExists(t, db.Read(), "vault_blocks", "listing_encrypted")
	assertColumnExists(t, db.Read(), "vault_blocks", "secret_encrypted")
}

func assertTableExists(t *testing.T, conn *sql.DB, tableName string) {
	t.Helper()

	ctx := context.Background()
	var count int
	err := conn.QueryRowContext(
		ctx,
		"SELECT COUNT(1) FROM sqlite_master WHERE type = 'table' AND name = ?",
		tableName,
	).Scan(&count)
	require.NoError(t, err)
	require.Equal(t, 1, count, "expected table %s to exist", tableName)
}

func assertColumnExists(t *testing.T, conn *sql.DB, tableName, columnName string) {
	t.Helper()

	rows, err := conn.QueryContext(context.Background(), "PRAGMA table_info("+tableName+")")
	require.NoError(t, err)
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var (
			cid        int
			name       string
			columnType string
			notNull    int
			defaultVal sql.NullString
			pk         int
		)

		err := rows.Scan(&cid, &name, &columnType, &notNull, &defaultVal, &pk)
		require.NoError(t, err)
		if name == columnName {
			return
		}
	}

	require.NoError(t, rows.Err())
	t.Fatalf("expected column %s to exist in table %s", columnName, tableName)
}
