-- Clean schema for direct execution by SQLite (no migration comments)
PRAGMA journal_mode = WAL;
PRAGMA foreign_keys = ON;
PRAGMA busy_timeout = 5000;
PRAGMA synchronous = NORMAL;

-- Roles table
CREATE TABLE IF NOT EXISTS roles (
    name TEXT PRIMARY KEY
);
INSERT OR IGNORE INTO roles (name) VALUES ('user'), ('assistant'), ('system'), ('tool');

-- Providers table
CREATE TABLE IF NOT EXISTS providers (
    name TEXT PRIMARY KEY
);

-- Models table
CREATE TABLE IF NOT EXISTS models (
    name TEXT PRIMARY KEY,
    provider TEXT REFERENCES providers(name),
    input_cost REAL DEFAULT 0,
    output_cost REAL DEFAULT 0
);

-- Chats table
CREATE TABLE IF NOT EXISTS chats (
    id TEXT PRIMARY KEY,
    parent_chat_id TEXT REFERENCES chats(id),
    title TEXT NOT NULL DEFAULT 'New Chat',
    message_count INTEGER NOT NULL DEFAULT 0,
    prompt_tokens INTEGER NOT NULL DEFAULT 0,
    completion_tokens INTEGER NOT NULL DEFAULT 0,
    cost REAL NOT NULL DEFAULT 0,
    summary_message_id TEXT,
    metadata TEXT DEFAULT '{}',
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);

-- Messages table
CREATE TABLE IF NOT EXISTS messages (
    id TEXT PRIMARY KEY,
    chat_id TEXT NOT NULL REFERENCES chats(id) ON DELETE CASCADE,
    role TEXT NOT NULL REFERENCES roles(name),
    content TEXT NOT NULL,
    model TEXT,
    provider TEXT,
    is_summary INTEGER NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now')),
    finished_at TEXT
);

-- Files table (snapshots/attachments)
CREATE TABLE IF NOT EXISTS files (
    id TEXT PRIMARY KEY,
    chat_id TEXT NOT NULL REFERENCES chats(id) ON DELETE CASCADE,
    path TEXT NOT NULL,
    version INTEGER NOT NULL DEFAULT 1,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now')),
    UNIQUE(chat_id, path)
);

-- Read files table (file access history)
CREATE TABLE IF NOT EXISTS read_files (
    chat_id TEXT NOT NULL REFERENCES chats(id) ON DELETE CASCADE,
    path TEXT NOT NULL,
    read_at TEXT NOT NULL DEFAULT (datetime('now')),
    PRIMARY KEY (chat_id, path, read_at)
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_messages_chat_id ON messages(chat_id);
CREATE INDEX IF NOT EXISTS idx_messages_created_at ON messages(created_at);
CREATE INDEX IF NOT EXISTS idx_messages_chat_created ON messages(chat_id, created_at);
CREATE INDEX IF NOT EXISTS idx_files_chat_id ON files(chat_id);
CREATE INDEX IF NOT EXISTS idx_chats_created_at ON chats(created_at);
CREATE INDEX IF NOT EXISTS idx_chats_parent_chat_id ON chats(parent_chat_id);
CREATE INDEX IF NOT EXISTS idx_read_files_chat_id ON read_files(chat_id);
CREATE INDEX IF NOT EXISTS idx_read_files_read_at ON read_files(read_at);

-- Trigger: auto-update updated_at on chats
CREATE TRIGGER IF NOT EXISTS update_chats_updated_at
AFTER UPDATE ON chats
BEGIN
    UPDATE chats SET updated_at = datetime('now') WHERE id = NEW.id;
END;

-- Trigger: auto-update message_count on chats after insert
CREATE TRIGGER IF NOT EXISTS update_chats_message_count_insert
AFTER INSERT ON messages
BEGIN
    UPDATE chats SET message_count = message_count + 1 WHERE id = NEW.chat_id;
END;

-- Trigger: auto-update message_count on chats after delete
CREATE TRIGGER IF NOT EXISTS update_chats_message_count_delete
AFTER DELETE ON messages
BEGIN
    UPDATE chats SET message_count = message_count - 1 WHERE id = OLD.chat_id;
END;

-- Trigger: auto-update updated_at on messages
CREATE TRIGGER IF NOT EXISTS update_messages_updated_at
AFTER UPDATE ON messages
BEGIN
    UPDATE messages SET updated_at = datetime('now') WHERE id = NEW.id;
END;

-- Trigger: auto-update updated_at on files
CREATE TRIGGER IF NOT EXISTS update_files_updated_at
AFTER UPDATE ON files
BEGIN
    UPDATE files SET updated_at = datetime('now') WHERE id = NEW.id;
END;