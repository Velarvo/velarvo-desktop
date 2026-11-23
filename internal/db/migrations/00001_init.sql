-- +goose Up

CREATE TABLE settings (
    key   TEXT PRIMARY KEY,
    value TEXT NOT NULL
) STRICT;

CREATE TABLE vault (
    id                INTEGER PRIMARY KEY CHECK (id = 1),
    kdf_time          INTEGER NOT NULL,
    kdf_memory        INTEGER NOT NULL,
    kdf_threads       INTEGER NOT NULL,
    salt              BLOB    NOT NULL CHECK (length(salt) = 16),
    wrapped_key       BLOB    NOT NULL,
    auto_lock_seconds INTEGER NOT NULL CHECK (auto_lock_seconds >= 0),
    created_at        INTEGER NOT NULL
) STRICT;

CREATE TABLE projects (
    id         TEXT    PRIMARY KEY,
    data       BLOB    NOT NULL,
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL
) STRICT;

CREATE TABLE project_icons (
    project_id TEXT    PRIMARY KEY REFERENCES projects (id) ON DELETE CASCADE,
    mime       TEXT    NOT NULL CHECK (mime IN ('image/png', 'image/jpeg', 'image/webp')),
    data       BLOB    NOT NULL,
    updated_at INTEGER NOT NULL
) STRICT;

CREATE TABLE workspaces (
    id         TEXT    PRIMARY KEY,
    project_id TEXT    NOT NULL REFERENCES projects (id) ON DELETE CASCADE,
    data       BLOB    NOT NULL,
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL
) STRICT;

CREATE INDEX workspaces_project ON workspaces (project_id, created_at);

CREATE TABLE ssh_keys (
    id             TEXT    PRIMARY KEY,
    workspace_id   TEXT    NOT NULL REFERENCES workspaces (id) ON DELETE CASCADE,
    label          TEXT    NOT NULL,
    key_type       TEXT    NOT NULL CHECK (key_type IN ('ed25519', 'rsa4096', 'ecdsa384')),
    public_key     TEXT    NOT NULL,
    fingerprint     TEXT    NOT NULL,
    has_passphrase INTEGER NOT NULL CHECK (has_passphrase IN (0, 1)),
    private_key    BLOB    NOT NULL,
    last_used_at   INTEGER,
    created_at     INTEGER NOT NULL,

    UNIQUE (workspace_id, fingerprint)
) STRICT;

CREATE TABLE ssh_connections (
    id           TEXT    PRIMARY KEY,
    workspace_id TEXT    NOT NULL REFERENCES workspaces (id) ON DELETE CASCADE,
    key_id       TEXT    REFERENCES ssh_keys (id) ON DELETE SET NULL,
    data         BLOB    NOT NULL,
    password     BLOB,
    last_used_at INTEGER,
    created_at   INTEGER NOT NULL,
    updated_at   INTEGER NOT NULL
) STRICT;

CREATE INDEX ssh_connections_workspace ON ssh_connections (workspace_id, created_at);

-- +goose Down

DROP TABLE ssh_connections;
DROP TABLE ssh_keys;
DROP TABLE workspaces;
DROP TABLE project_icons;
DROP TABLE projects;
DROP TABLE vault;
DROP TABLE settings;
