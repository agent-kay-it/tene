package vault

const schemaSQL = `
CREATE TABLE IF NOT EXISTS vault_meta (
    key   TEXT PRIMARY KEY,
    value TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS secrets (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    name            TEXT    NOT NULL,
    encrypted_value TEXT    NOT NULL,
    environment     TEXT    NOT NULL DEFAULT 'default',
    version         INTEGER NOT NULL DEFAULT 1,
    created_at      TEXT    NOT NULL DEFAULT (datetime('now')),
    updated_at      TEXT    NOT NULL DEFAULT (datetime('now')),
    UNIQUE(name, environment)
);

CREATE INDEX IF NOT EXISTS idx_secrets_env ON secrets(environment);
CREATE INDEX IF NOT EXISTS idx_secrets_name ON secrets(name);

CREATE TABLE IF NOT EXISTS environments (
    name       TEXT    PRIMARY KEY,
    is_active  INTEGER NOT NULL DEFAULT 0,
    created_at TEXT    NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS audit_log (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    action        TEXT NOT NULL,
    resource_name TEXT,
    details       TEXT,
    timestamp     TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_audit_timestamp ON audit_log(timestamp);
`

const currentSchemaVersion = 1

func (v *Vault) initSchema() error {
	_, err := v.db.Exec(schemaSQL)
	return err
}
