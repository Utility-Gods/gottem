-- schema.sql

-- Chats table
CREATE TABLE IF NOT EXISTS chats (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    context TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- API Keys table
CREATE TABLE IF NOT EXISTS api_keys (
    api_name TEXT PRIMARY KEY,
    api_key TEXT NOT NULL
);

-- Schema version table
CREATE TABLE IF NOT EXISTS schema_version (
    version INTEGER PRIMARY KEY
);

-- Insert initial schema version
INSERT OR REPLACE INTO schema_version (version) VALUES (1);

-- Trigger to update chat timestamp
CREATE TRIGGER IF NOT EXISTS update_chats_timestamp
AFTER UPDATE ON chats
FOR EACH ROW
BEGIN
    UPDATE chats SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;

-- Index for faster querying on updated_at
CREATE INDEX IF NOT EXISTS idx_chats_updated_at ON chats(updated_at);
