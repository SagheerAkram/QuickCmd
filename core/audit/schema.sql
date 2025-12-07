-- QuickCMD Audit Log Schema
-- This schema stores all command executions for audit and history purposes

CREATE TABLE IF NOT EXISTS runs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    timestamp TEXT NOT NULL,
    user TEXT,
    prompt TEXT NOT NULL,
    selected_command TEXT NOT NULL,
    sandbox_id TEXT,
    exit_code INTEGER,
    stdout BLOB,
    stderr BLOB,
    risk_level TEXT NOT NULL,
    snapshot TEXT,
    executed BOOLEAN DEFAULT 0,
    duration_ms INTEGER,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_runs_timestamp ON runs(timestamp);
CREATE INDEX IF NOT EXISTS idx_runs_user ON runs(user);
CREATE INDEX IF NOT EXISTS idx_runs_executed ON runs(executed);
CREATE INDEX IF NOT EXISTS idx_runs_risk_level ON runs(risk_level);

-- Metadata table for schema versioning
CREATE TABLE IF NOT EXISTS schema_version (
    version INTEGER PRIMARY KEY,
    applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

INSERT OR IGNORE INTO schema_version (version) VALUES (1);
