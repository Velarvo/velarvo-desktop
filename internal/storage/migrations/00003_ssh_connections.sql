-- +goose Up
-- +goose StatementBegin

CREATE TABLE IF NOT EXISTS ssh_connections (
    id                  TEXT PRIMARY KEY,

    listing_encrypted   BLOB NOT NULL,
    listing_version     INTEGER NOT NULL,

    secret_encrypted    BLOB NOT NULL,
    secret_version      INTEGER NOT NULL,

    sort_order          INTEGER NOT NULL DEFAULT 0,

    last_used_at        INTEGER,

    created_at          INTEGER NOT NULL,
    updated_at          INTEGER NOT NULL,
    server_updated_at   INTEGER,
    deleted_at          INTEGER,
    revision            TEXT NOT NULL,

    CHECK (length(id) = 36),
    CHECK (length(revision) = 36),
    CHECK (length(listing_encrypted) >= 29),
    CHECK (length(secret_encrypted) >= 29),
    CHECK (listing_version >= 1),
    CHECK (secret_version >= 1),
    CHECK (sort_order >= 0),
    CHECK (last_used_at IS NULL OR last_used_at > 0),
    CHECK (created_at > 0),
    CHECK (updated_at >= created_at),
    CHECK (server_updated_at IS NULL OR server_updated_at > 0),
    CHECK (deleted_at IS NULL OR deleted_at >= updated_at)
) STRICT;

CREATE INDEX IF NOT EXISTS idx_ssh_connections_sort_active
    ON ssh_connections(sort_order, created_at)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_ssh_connections_updated_at
    ON ssh_connections(updated_at);

CREATE INDEX IF NOT EXISTS idx_ssh_connections_deleted_at
    ON ssh_connections(deleted_at)
    WHERE deleted_at IS NOT NULL;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_ssh_connections_deleted_at;
DROP INDEX IF EXISTS idx_ssh_connections_updated_at;
DROP INDEX IF EXISTS idx_ssh_connections_sort_active;
DROP TABLE IF EXISTS ssh_connections;
-- +goose StatementEnd
