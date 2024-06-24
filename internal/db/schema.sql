-- API Keys table
CREATE TABLE IF NOT EXISTS api_keys (
    api_name TEXT PRIMARY KEY,
    api_key TEXT NOT NULL
);

-- Chats table
CREATE TABLE IF NOT EXISTS chats (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Messages table
CREATE TABLE IF NOT EXISTS messages (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    chat_id INTEGER NOT NULL,
    role TEXT NOT NULL,
    api_name TEXT NOT NULL,
    content TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (chat_id) REFERENCES chats (id) ON DELETE CASCADE
);

-- Trigger to update chat timestamp
CREATE TRIGGER IF NOT EXISTS update_chat_timestamp
AFTER INSERT ON messages
BEGIN
    UPDATE chats SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.chat_id;
END;

-- Schema version table
CREATE TABLE IF NOT EXISTS schema_version (
    version INTEGER PRIMARY KEY
);
