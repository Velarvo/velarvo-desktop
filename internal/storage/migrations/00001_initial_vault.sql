-- +goose Up
-- +goose StatementBegin

CREATE TABLE IF NOT EXISTS app_meta (
    key   TEXT PRIMARY KEY,
    value TEXT NOT NULL
) STRICT;

INSERT INTO app_meta (key, value) VALUES
    ('app_version',     '1'),
    ('initialized_at',  CAST(strftime('%s', 'now') AS TEXT))
    ON CONFLICT(key) DO NOTHING;

CREATE TABLE IF NOT EXISTS vault_meta (
    id                  INTEGER PRIMARY KEY CHECK (id = 1),

    schema_version      INTEGER NOT NULL,

    crypto_version      INTEGER NOT NULL DEFAULT 1,

    created_at          INTEGER NOT NULL,

    kdf_id              INTEGER NOT NULL,
    kdf_time            INTEGER NOT NULL,
    kdf_memory          INTEGER NOT NULL,
    kdf_threads         INTEGER NOT NULL,
    kdf_keylen          INTEGER NOT NULL,

    salt_kek            BLOB NOT NULL,
    salt_auth           BLOB NOT NULL,

    device_id           TEXT NOT NULL,

    auto_lock_seconds   INTEGER NOT NULL DEFAULT 900,

    CHECK (schema_version >= 1),
    CHECK (crypto_version >= 1),
    CHECK (kdf_id = 1),
    CHECK (kdf_time >= 1),
    CHECK (kdf_memory >= 8192),
    CHECK (kdf_threads >= 1),
    CHECK (kdf_keylen = 32),
    CHECK (length(salt_kek) = 16),
    CHECK (length(salt_auth) = 16),
    CHECK (length(device_id) = 36),
    CHECK (auto_lock_seconds >= 0),
    CHECK (created_at > 0)
) STRICT;

CREATE TABLE IF NOT EXISTS dek_envelopes (
    method        TEXT PRIMARY KEY,

    envelope      BLOB NOT NULL,

    metadata_json TEXT,

    created_at    INTEGER NOT NULL,
    updated_at    INTEGER NOT NULL,

    CHECK (method IN ('password')),
    CHECK (length(envelope) >= 29),
    CHECK (metadata_json IS NULL OR json_valid(metadata_json)),
    CHECK (created_at > 0),
    CHECK (updated_at >= created_at)
) STRICT;

CREATE TABLE IF NOT EXISTS projects (
    id                  TEXT PRIMARY KEY,

    encrypted_data      BLOB NOT NULL,

    data_version        INTEGER NOT NULL,

    sort_order          INTEGER NOT NULL DEFAULT 0,

    created_at          INTEGER NOT NULL,
    updated_at          INTEGER NOT NULL,

    server_updated_at   INTEGER,

    deleted_at          INTEGER,

    revision            TEXT NOT NULL,

    CHECK (length(id) = 36),
    CHECK (length(revision) = 36),
    CHECK (length(encrypted_data) >= 29),
    CHECK (data_version >= 1),
    CHECK (created_at > 0),
    CHECK (updated_at >= created_at),
    CHECK (server_updated_at IS NULL OR server_updated_at > 0),
    CHECK (deleted_at IS NULL OR deleted_at >= updated_at)
) STRICT;

CREATE INDEX IF NOT EXISTS idx_projects_sort_active
    ON projects(sort_order, created_at)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_projects_updated_at
    ON projects(updated_at);

CREATE INDEX IF NOT EXISTS idx_projects_deleted_at
    ON projects(deleted_at)
    WHERE deleted_at IS NOT NULL;

CREATE TABLE IF NOT EXISTS project_icons (
    project_id          TEXT PRIMARY KEY,

    encrypted_icon      BLOB NOT NULL,
    icon_mime           TEXT NOT NULL,

    local_status        TEXT NOT NULL DEFAULT 'present',

    sync_state          TEXT NOT NULL DEFAULT 'local_only',

    created_at          INTEGER NOT NULL,
    updated_at          INTEGER NOT NULL,
    server_updated_at   INTEGER,
    revision            TEXT NOT NULL,

    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE,

    CHECK (length(project_id) = 36),
    CHECK (length(revision) = 36),
    CHECK (length(encrypted_icon) >= 29),
    CHECK (length(encrypted_icon) <= 524288),
    CHECK (icon_mime IN ('image/png', 'image/jpeg', 'image/webp')),
    CHECK (local_status IN ('present', 'missing')),
    CHECK (sync_state IN ('local_only', 'syncing', 'synced', 'remote_only', 'failed')),
    CHECK (created_at > 0),
    CHECK (updated_at >= created_at),
    CHECK (server_updated_at IS NULL OR server_updated_at > 0)
) STRICT;

CREATE INDEX IF NOT EXISTS idx_project_icons_sync
    ON project_icons(sync_state)
    WHERE sync_state IN ('local_only', 'syncing', 'failed');

CREATE TABLE IF NOT EXISTS workspaces (
    id                  TEXT PRIMARY KEY,
    project_id          TEXT NOT NULL,

    encrypted_data      BLOB NOT NULL,
    data_version        INTEGER NOT NULL,

    sort_order          INTEGER NOT NULL DEFAULT 0,

    created_at          INTEGER NOT NULL,
    updated_at          INTEGER NOT NULL,
    server_updated_at   INTEGER,
    deleted_at          INTEGER,
    revision            TEXT NOT NULL,

    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE,

    CHECK (length(id) = 36),
    CHECK (length(project_id) = 36),
    CHECK (length(revision) = 36),
    CHECK (length(encrypted_data) >= 29),
    CHECK (data_version >= 1),
    CHECK (created_at > 0),
    CHECK (updated_at >= created_at),
    CHECK (server_updated_at IS NULL OR server_updated_at > 0),
    CHECK (deleted_at IS NULL OR deleted_at >= updated_at)
) STRICT;

CREATE INDEX IF NOT EXISTS idx_workspaces_project_sort_active
    ON workspaces(project_id, sort_order, created_at)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_workspaces_updated_at
    ON workspaces(updated_at);

CREATE INDEX IF NOT EXISTS idx_workspaces_deleted_at
    ON workspaces(deleted_at)
    WHERE deleted_at IS NOT NULL;

CREATE TABLE IF NOT EXISTS collections (
    id                  TEXT PRIMARY KEY,
    workspace_id        TEXT NOT NULL,

    kind                TEXT NOT NULL,

    encrypted_data      BLOB NOT NULL,
    data_version        INTEGER NOT NULL,

    sort_order          INTEGER NOT NULL DEFAULT 0,

    created_at          INTEGER NOT NULL,
    updated_at          INTEGER NOT NULL,
    server_updated_at   INTEGER,
    deleted_at          INTEGER,
    revision            TEXT NOT NULL,

    FOREIGN KEY (workspace_id) REFERENCES workspaces(id) ON DELETE CASCADE,

    CHECK (kind IN ('api', 'ssh', 'database', 'vault')),
    CHECK (length(id) = 36),
    CHECK (length(workspace_id) = 36),
    CHECK (length(revision) = 36),
    CHECK (length(encrypted_data) >= 29),
    CHECK (data_version >= 1),
    CHECK (created_at > 0),
    CHECK (updated_at >= created_at),
    CHECK (server_updated_at IS NULL OR server_updated_at > 0),
    CHECK (deleted_at IS NULL OR deleted_at >= updated_at)
) STRICT;

CREATE INDEX IF NOT EXISTS idx_collections_workspace_kind_sort_active
    ON collections(workspace_id, kind, sort_order)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_collections_updated_at
    ON collections(updated_at);

CREATE INDEX IF NOT EXISTS idx_collections_deleted_at
    ON collections(deleted_at)
    WHERE deleted_at IS NOT NULL;

CREATE TABLE IF NOT EXISTS vault_blocks (
    id                  TEXT PRIMARY KEY,
    workspace_id        TEXT NOT NULL,

    collection_id       TEXT,

    kind                TEXT NOT NULL,

    listing_encrypted   BLOB NOT NULL,
    listing_version     INTEGER NOT NULL,


    secret_encrypted    BLOB NOT NULL,
    secret_version      INTEGER NOT NULL,

    last_used_at        INTEGER,

    created_at          INTEGER NOT NULL,
    updated_at          INTEGER NOT NULL,
    server_updated_at   INTEGER,
    deleted_at          INTEGER,
    revision            TEXT NOT NULL,

    FOREIGN KEY (workspace_id) REFERENCES workspaces(id) ON DELETE CASCADE,
    FOREIGN KEY (collection_id) REFERENCES collections(id) ON DELETE SET NULL,

    CHECK (kind IN ('credential', 'api_key', 'ssh_connection', 'db_connection',
                    'api_request', 'env_var', 'note', 'secret')),
    CHECK (length(id) = 36),
    CHECK (length(workspace_id) = 36),
    CHECK (collection_id IS NULL OR length(collection_id) = 36),
    CHECK (length(revision) = 36),
    CHECK (length(listing_encrypted) >= 29),
    CHECK (length(secret_encrypted) >= 29),
    CHECK (listing_version >= 1),
    CHECK (secret_version >= 1),
    CHECK (last_used_at IS NULL OR last_used_at > 0),
    CHECK (created_at > 0),
    CHECK (updated_at >= created_at),
    CHECK (server_updated_at IS NULL OR server_updated_at > 0),
    CHECK (deleted_at IS NULL OR deleted_at >= updated_at)
) STRICT;

CREATE INDEX IF NOT EXISTS idx_vault_blocks_workspace_kind_active
    ON vault_blocks(workspace_id, kind, last_used_at DESC)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_vault_blocks_collection_active
    ON vault_blocks(collection_id, last_used_at DESC)
    WHERE deleted_at IS NULL AND collection_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_vault_blocks_updated_at
    ON vault_blocks(updated_at);

CREATE INDEX IF NOT EXISTS idx_vault_blocks_deleted_at
    ON vault_blocks(deleted_at)
    WHERE deleted_at IS NOT NULL;

CREATE TABLE IF NOT EXISTS sync_state (
    id                  INTEGER PRIMARY KEY CHECK (id = 1),

    enabled             INTEGER NOT NULL DEFAULT 0,

    server_url          TEXT,

    last_sync_at        INTEGER,

    last_sync_cursor    TEXT,

    status              TEXT NOT NULL DEFAULT 'idle',

    last_error          TEXT,
    last_error_at       INTEGER,

    pending_uploads     INTEGER NOT NULL DEFAULT 0,

    CHECK (enabled IN (0, 1)),
    CHECK (status IN ('idle', 'syncing', 'error', 'paused')),
    CHECK (last_sync_at IS NULL OR last_sync_at > 0),
    CHECK (last_error_at IS NULL OR last_error_at > 0),
    CHECK (pending_uploads >= 0)
) STRICT;

INSERT INTO sync_state (id, enabled, status)
    VALUES (1, 0, 'idle')
    ON CONFLICT(id) DO NOTHING;

CREATE TABLE IF NOT EXISTS sync_outbox (
    seq                 INTEGER PRIMARY KEY AUTOINCREMENT,

    entity_type         TEXT NOT NULL,
    entity_id           TEXT NOT NULL,

    operation           TEXT NOT NULL,

    revision            TEXT NOT NULL,

    created_at          INTEGER NOT NULL,

    last_attempt_at     INTEGER,
    attempt_count       INTEGER NOT NULL DEFAULT 0,

    last_error          TEXT,

    CHECK (entity_type IN ('project', 'project_icon', 'workspace', 'collection', 'vault_block')),
    CHECK (operation IN ('create', 'update', 'delete')),
    CHECK (length(entity_id) = 36),
    CHECK (length(revision) = 36),
    CHECK (created_at > 0),
    CHECK (last_attempt_at IS NULL OR last_attempt_at > 0),
    CHECK (attempt_count >= 0)
) STRICT;

CREATE INDEX IF NOT EXISTS idx_sync_outbox_attempt
    ON sync_outbox(last_attempt_at, attempt_count);

CREATE TABLE IF NOT EXISTS sync_conflicts (
    id                  TEXT PRIMARY KEY,

    entity_type         TEXT NOT NULL,
    entity_id           TEXT NOT NULL,

    local_revision      TEXT NOT NULL,
    local_data          BLOB NOT NULL,
    local_updated_at    INTEGER NOT NULL,

    server_revision     TEXT NOT NULL,
    server_data         BLOB NOT NULL,
    server_updated_at   INTEGER NOT NULL,

    resolution          TEXT,
    resolved_at         INTEGER,

    detected_at         INTEGER NOT NULL,

    CHECK (entity_type IN ('project', 'project_icon', 'workspace', 'collection', 'vault_block')),
    CHECK (length(id) = 36),
    CHECK (length(entity_id) = 36),
    CHECK (length(local_revision) = 36),
    CHECK (length(server_revision) = 36),
    CHECK (resolution IS NULL OR resolution IN ('local', 'server', 'merged')),
    CHECK (local_updated_at > 0),
    CHECK (server_updated_at > 0),
    CHECK (resolved_at IS NULL OR resolved_at >= detected_at),
    CHECK (detected_at > 0)
) STRICT;

CREATE INDEX IF NOT EXISTS idx_sync_conflicts_unresolved
    ON sync_conflicts(entity_type, entity_id)
    WHERE resolution IS NULL;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_sync_conflicts_unresolved;
DROP TABLE IF EXISTS sync_conflicts;

DROP INDEX IF EXISTS idx_sync_outbox_attempt;
DROP TABLE IF EXISTS sync_outbox;

DROP TABLE IF EXISTS sync_state;

DROP INDEX IF EXISTS idx_vault_blocks_deleted_at;
DROP INDEX IF EXISTS idx_vault_blocks_updated_at;
DROP INDEX IF EXISTS idx_vault_blocks_collection_active;
DROP INDEX IF EXISTS idx_vault_blocks_workspace_kind_active;
DROP TABLE IF EXISTS vault_blocks;

DROP INDEX IF EXISTS idx_collections_deleted_at;
DROP INDEX IF EXISTS idx_collections_updated_at;
DROP INDEX IF EXISTS idx_collections_workspace_kind_sort_active;
DROP TABLE IF EXISTS collections;

DROP INDEX IF EXISTS idx_workspaces_deleted_at;
DROP INDEX IF EXISTS idx_workspaces_updated_at;
DROP INDEX IF EXISTS idx_workspaces_project_sort_active;
DROP TABLE IF EXISTS workspaces;

DROP INDEX IF EXISTS idx_project_icons_sync;
DROP TABLE IF EXISTS project_icons;

DROP INDEX IF EXISTS idx_projects_deleted_at;
DROP INDEX IF EXISTS idx_projects_updated_at;
DROP INDEX IF EXISTS idx_projects_sort_active;
DROP TABLE IF EXISTS projects;

DROP TABLE IF EXISTS dek_envelopes;
DROP TABLE IF EXISTS vault_meta;
DROP TABLE IF EXISTS app_meta;
-- +goose StatementEnd
