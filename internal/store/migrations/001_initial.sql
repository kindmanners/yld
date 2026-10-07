PRAGMA foreign_keys = ON;

CREATE TABLE users (
    id TEXT PRIMARY KEY,
    handle TEXT NOT NULL UNIQUE,
    display_name TEXT NOT NULL,
    timezone TEXT NOT NULL,
    password_hash TEXT NOT NULL,
    is_pseudonym INTEGER NOT NULL DEFAULT 0 CHECK (is_pseudonym IN (0, 1)),
    created_at TEXT NOT NULL
);

CREATE TABLE spaces (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    is_personal INTEGER NOT NULL DEFAULT 0 CHECK (is_personal IN (0, 1)),
    created_at TEXT NOT NULL
);

CREATE TABLE space_members (
    space_id TEXT NOT NULL REFERENCES spaces(id) ON DELETE CASCADE,
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role TEXT NOT NULL CHECK (role IN ('owner', 'member', 'former')),
    joined_at TEXT NOT NULL,
    PRIMARY KEY (space_id, user_id)
);

CREATE TABLE metrics (
    id TEXT PRIMARY KEY,
    owner_space_id TEXT REFERENCES spaces(id) ON DELETE RESTRICT,
    key TEXT NOT NULL,
    name TEXT NOT NULL,
    category TEXT NOT NULL DEFAULT '',
    kind TEXT NOT NULL CHECK (kind IN ('number', 'integer', 'duration')),
    aggregation TEXT NOT NULL CHECK (aggregation IN ('sum', 'average', 'minimum', 'maximum', 'latest', 'count')),
    unit TEXT NOT NULL DEFAULT '',
    archived INTEGER NOT NULL DEFAULT 0 CHECK (archived IN (0, 1)),
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    UNIQUE (owner_space_id, key)
);

CREATE UNIQUE INDEX metrics_builtin_key_unique ON metrics(key) WHERE owner_space_id IS NULL;

CREATE TABLE activities (
    id TEXT PRIMARY KEY,
    space_id TEXT NOT NULL REFERENCES spaces(id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    occurred_at TEXT NOT NULL,
    local_date TEXT NOT NULL,
    timezone TEXT NOT NULL,
    created_at TEXT NOT NULL
);

CREATE TABLE entries (
    id TEXT PRIMARY KEY,
    space_id TEXT NOT NULL REFERENCES spaces(id) ON DELETE RESTRICT,
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    metric_id TEXT NOT NULL REFERENCES metrics(id) ON DELETE RESTRICT,
    activity_id TEXT REFERENCES activities(id) ON DELETE SET NULL,
    occurred_at TEXT NOT NULL,
    local_date TEXT NOT NULL,
    timezone TEXT NOT NULL,
    value REAL NOT NULL,
    note TEXT NOT NULL DEFAULT '',
    source TEXT NOT NULL DEFAULT 'manual',
    external_id TEXT,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    deleted_at TEXT
);

CREATE INDEX entries_recap_lookup ON entries(user_id, space_id, local_date, metric_id) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX entries_source_identity ON entries(user_id, source, external_id) WHERE external_id IS NOT NULL;

CREATE TABLE recaps (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    space_id TEXT NOT NULL REFERENCES spaces(id) ON DELETE CASCADE,
    period_kind TEXT NOT NULL CHECK (period_kind IN ('month', 'quarter', 'year')),
    start_date TEXT NOT NULL,
    end_date TEXT NOT NULL,
    generated_at TEXT NOT NULL,
    sent_at TEXT,
    source_updated_at TEXT NOT NULL,
    UNIQUE (user_id, space_id, period_kind, start_date)
);

CREATE TABLE recap_items (
    recap_id TEXT NOT NULL REFERENCES recaps(id) ON DELETE CASCADE,
    metric_id TEXT NOT NULL REFERENCES metrics(id) ON DELETE RESTRICT,
    metric_key TEXT NOT NULL,
    metric_name TEXT NOT NULL,
    unit TEXT NOT NULL,
    value REAL NOT NULL,
    entry_count INTEGER NOT NULL,
    PRIMARY KEY (recap_id, metric_id)
);
