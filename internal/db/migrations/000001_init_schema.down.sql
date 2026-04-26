-- +migrate Down
DROP TRIGGER IF EXISTS update_files_updated_at;
DROP TRIGGER IF EXISTS update_messages_updated_at;
DROP TRIGGER IF EXISTS update_chats_message_count_delete;
DROP TRIGGER IF EXISTS update_chats_message_count_insert;
DROP TRIGGER IF EXISTS update_chats_updated_at;

DROP INDEX IF EXISTS idx_read_files_read_at;
DROP INDEX IF EXISTS idx_read_files_chat_id;
DROP INDEX IF EXISTS idx_chats_parent_chat_id;
DROP INDEX IF EXISTS idx_chats_created_at;
DROP INDEX IF EXISTS idx_files_chat_id;
DROP INDEX IF EXISTS idx_messages_chat_created;
DROP INDEX IF EXISTS idx_messages_created_at;
DROP INDEX IF EXISTS idx_messages_chat_id;

DROP TABLE IF EXISTS read_files;
DROP TABLE IF EXISTS files;
DROP TABLE IF EXISTS messages;
DROP TABLE IF EXISTS chats;
DROP TABLE IF EXISTS models;
DROP TABLE IF EXISTS providers;
DROP TABLE IF EXISTS roles;