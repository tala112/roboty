package db_test

import (
	"context"
	"os"
	"testing"

	"Roboty/internal/db"
)

const schemaSQL = `
PRAGMA journal_mode = WAL;
PRAGMA foreign_keys = ON;
PRAGMA busy_timeout = 5000;
PRAGMA synchronous = NORMAL;

CREATE TABLE IF NOT EXISTS roles (
    name TEXT PRIMARY KEY
);
INSERT INTO roles (name) VALUES ('user'), ('assistant'), ('system'), ('tool') ON CONFLICT DO NOTHING;

CREATE TABLE IF NOT EXISTS providers (
    name TEXT PRIMARY KEY
);

CREATE TABLE IF NOT EXISTS models (
    name TEXT PRIMARY KEY,
    provider TEXT REFERENCES providers(name),
    input_cost REAL DEFAULT 0,
    output_cost REAL DEFAULT 0
);

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

CREATE TABLE IF NOT EXISTS files (
    id TEXT PRIMARY KEY,
    chat_id TEXT NOT NULL REFERENCES chats(id) ON DELETE CASCADE,
    path TEXT NOT NULL,
    version INTEGER NOT NULL DEFAULT 1,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now')),
    UNIQUE(chat_id, path)
);

CREATE TABLE IF NOT EXISTS read_files (
    chat_id TEXT NOT NULL REFERENCES chats(id) ON DELETE CASCADE,
    path TEXT NOT NULL,
    read_at TEXT NOT NULL DEFAULT (datetime('now')),
    PRIMARY KEY (chat_id, path, read_at)
);

CREATE INDEX IF NOT EXISTS idx_messages_chat_id ON messages(chat_id);
CREATE INDEX IF NOT EXISTS idx_messages_created_at ON messages(created_at);
CREATE INDEX IF NOT EXISTS idx_messages_chat_created ON messages(chat_id, created_at);
CREATE INDEX IF NOT EXISTS idx_files_chat_id ON files(chat_id);
CREATE INDEX IF NOT EXISTS idx_chats_created_at ON chats(created_at);
CREATE INDEX IF NOT EXISTS idx_chats_parent_chat_id ON chats(parent_chat_id);
CREATE INDEX IF NOT EXISTS idx_read_files_chat_id ON read_files(chat_id);
CREATE INDEX IF NOT EXISTS idx_read_files_read_at ON read_files(read_at);

CREATE TRIGGER IF NOT EXISTS update_chats_updated_at
AFTER UPDATE ON chats
BEGIN
    UPDATE chats SET updated_at = datetime('now') WHERE id = NEW.id;
END;

CREATE TRIGGER IF NOT EXISTS update_chats_message_count_insert
AFTER INSERT ON messages
BEGIN
    UPDATE chats SET message_count = message_count + 1 WHERE id = NEW.chat_id;
END;

CREATE TRIGGER IF NOT EXISTS update_chats_message_count_delete
AFTER DELETE ON messages
BEGIN
    UPDATE chats SET message_count = message_count - 1 WHERE id = OLD.chat_id;
END;

CREATE TRIGGER IF NOT EXISTS update_messages_updated_at
AFTER UPDATE ON messages
BEGIN
    UPDATE messages SET updated_at = datetime('now') WHERE id = NEW.id;
END;

CREATE TRIGGER IF NOT EXISTS update_files_updated_at
AFTER UPDATE ON files
BEGIN
    UPDATE files SET updated_at = datetime('now') WHERE id = NEW.id;
END;
`

func TestDatabaseIntegration(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := tmpDir + "/roboty.db"

	database, err := db.NewDBWithPath(dbPath)
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	defer database.Close()

	_, err = database.DB().Exec(schemaSQL)
	if err != nil {
		t.Fatalf("failed to create schema: %v", err)
	}

	queries := db.NewQueries(database.DB())
	ctx := context.Background()

	// Test: Create chat
	chat, err := queries.CreateChat(ctx, db.CreateChatParams{
		ID:       "test-chat-1",
		Title:    "Test Chat",
		Metadata: "{}",
	})
	if err != nil {
		t.Fatalf("failed to create chat: %v", err)
	}
	if chat.Title != "Test Chat" {
		t.Errorf("expected title 'Test Chat', got %q", chat.Title)
	}

	// Test: Get chat
	chat, err = queries.GetChatByID(ctx, "test-chat-1")
	if err != nil {
		t.Fatalf("failed to get chat: %v", err)
	}
	if chat.MessageCount != 0 {
		t.Errorf("expected message_count 0, got %d", chat.MessageCount)
	}

	// Test: Create message
	msg, err := queries.CreateMessage(ctx, db.CreateMessageParams{
		ID:       "test-msg-1",
		ChatID:   "test-chat-1",
		Role:     db.RoleUser,
		Content:  `{"text": "Hello"}`,
	})
	if err != nil {
		t.Fatalf("failed to create message: %v", err)
	}
	if msg.Role != db.RoleUser {
		t.Errorf("expected role user, got %s", msg.Role)
	}

	// Test: Verify message_count was auto-updated by trigger
	chat, err = queries.GetChatByID(ctx, "test-chat-1")
	if err != nil {
		t.Fatalf("failed to get chat: %v", err)
	}
	if chat.MessageCount != 1 {
		t.Errorf("expected message_count 1, got %d", chat.MessageCount)
	}

	// Test: Get messages
	messages, err := queries.GetMessagesByChatID(ctx, "test-chat-1")
	if err != nil {
		t.Fatalf("failed to get messages: %v", err)
	}
	if len(messages) != 1 {
		t.Errorf("expected 1 message, got %d", len(messages))
	}

	// Test: Update chat stats
	err = queries.UpdateChatStats(ctx, db.UpdateChatStatsParams{
		ID:               "test-chat-1",
		PromptTokens:    100,
		CompletionTokens: 50,
		Cost:            0.002,
	})
	if err != nil {
		t.Fatalf("failed to update stats: %v", err)
	}
	chat, err = queries.GetChatByID(ctx, "test-chat-1")
	if err != nil {
		t.Fatalf("failed to get chat: %v", err)
	}
	if chat.PromptTokens != 100 || chat.CompletionTokens != 50 || chat.Cost != 0.002 {
		t.Errorf("stats not updated: prompt=%d, completion=%d, cost=%f", chat.PromptTokens, chat.CompletionTokens, chat.Cost)
	}

	// Test: Record read file
	err = queries.RecordReadFile(ctx, db.RecordReadFileParams{
		ChatID: "test-chat-1",
		Path:  "C:\\test\\file.txt",
	})
	if err != nil {
		t.Fatalf("failed to record read file: %v", err)
	}

	// Test: CASCADE delete
	err = queries.DeleteChat(ctx, "test-chat-1")
	if err != nil {
		t.Fatalf("failed to delete chat: %v", err)
	}

	// Verify messages were cascade-deleted
	messages, err = queries.GetMessagesByChatID(ctx, "test-chat-1")
	if err != nil {
		t.Fatalf("failed to get messages after delete: %v", err)
	}
	if len(messages) != 0 {
		t.Errorf("expected 0 messages after cascade delete, got %d", len(messages))
	}

	os.Remove(dbPath)
}