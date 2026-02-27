CREATE TABLE IF NOT EXISTS builtin_tools (
    name            VARCHAR(100) PRIMARY KEY,
    display_name    VARCHAR(255) NOT NULL,
    description     TEXT NOT NULL DEFAULT '',
    category        VARCHAR(50) NOT NULL DEFAULT 'general',
    enabled         BOOLEAN NOT NULL DEFAULT true,
    settings        JSONB NOT NULL DEFAULT '{}',
    requires        TEXT[] DEFAULT '{}',
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_builtin_tools_category ON builtin_tools(category);
