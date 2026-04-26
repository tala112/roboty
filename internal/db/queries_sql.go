package db

// Chat queries
const (
	createChat = `
INSERT INTO chats (
    id, parent_chat_id, title, metadata, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, datetime('now'), datetime('now')
) RETURNING *`

	getChatByID = `
SELECT id, parent_chat_id, title, message_count, prompt_tokens, completion_tokens, cost, summary_message_id, metadata, created_at, updated_at
FROM chats WHERE id = $1`

	getAllChats = `
SELECT id, parent_chat_id, title, message_count, prompt_tokens, completion_tokens, cost, summary_message_id, metadata, created_at, updated_at
FROM chats ORDER BY updated_at DESC`

	getChatsByParentID = `
SELECT id, parent_chat_id, title, message_count, prompt_tokens, completion_tokens, cost, summary_message_id, metadata, created_at, updated_at
FROM chats WHERE parent_chat_id = $1 ORDER BY created_at ASC`

	updateChatTitle = `
UPDATE chats SET title = $2, updated_at = datetime('now') WHERE id = $1`

	updateChatMetadata = `
UPDATE chats SET metadata = $2, updated_at = datetime('now') WHERE id = $1`

	updateChatStats = `
UPDATE chats SET 
    prompt_tokens = prompt_tokens + $2,
    completion_tokens = completion_tokens + $3,
    cost = cost + $4,
    updated_at = datetime('now')
WHERE id = $1`

	setChatSummary = `
UPDATE chats SET summary_message_id = $2, updated_at = datetime('now') WHERE id = $1`

	deleteChat = `DELETE FROM chats WHERE id = $1`
)

// Message queries
const (
	createMessage = `
INSERT INTO messages (
    id, chat_id, role, content, model, provider, is_summary, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, datetime('now'), datetime('now')
) RETURNING *`

	getMessageByID = `
SELECT id, chat_id, role, content, model, provider, is_summary, created_at, updated_at, finished_at
FROM messages WHERE id = $1`

	getMessagesByChatID = `
SELECT id, chat_id, role, content, model, provider, is_summary, created_at, updated_at, finished_at
FROM messages WHERE chat_id = $1 ORDER BY created_at ASC`

	getMessagesByChatIDPaginated = `
SELECT id, chat_id, role, content, model, provider, is_summary, created_at, updated_at, finished_at
FROM messages WHERE chat_id = $1 ORDER BY created_at ASC LIMIT $2 OFFSET $3`

	getChatLatestMessage = `
SELECT id, chat_id, role, content, model, provider, is_summary, created_at, updated_at, finished_at
FROM messages WHERE chat_id = $1 ORDER BY created_at DESC LIMIT 1`

	updateMessageContent = `
UPDATE messages SET content = $2, updated_at = datetime('now') WHERE id = $1`

	updateMessageFinishedAt = `
UPDATE messages SET finished_at = datetime('now') WHERE id = $1`

	deleteMessage = `DELETE FROM messages WHERE id = $1`

	deleteMessagesByChatID = `DELETE FROM messages WHERE chat_id = $1`

	countMessagesByChatID = `SELECT COUNT(*) FROM messages WHERE chat_id = $1`
)

// File queries
const (
	createFile = `
INSERT INTO files (
    id, chat_id, path, version, created_at, updated_at
) VALUES (
    $1, $2, $3, 1, datetime('now'), datetime('now')
) RETURNING *`

	getFileByID = `
SELECT id, chat_id, path, version, created_at, updated_at
FROM files WHERE id = $1`

	getFileByChatAndPath = `
SELECT id, chat_id, path, version, created_at, updated_at
FROM files WHERE chat_id = $1 AND path = $2`

	getFilesByChatID = `
SELECT id, chat_id, path, version, created_at, updated_at
FROM files WHERE chat_id = $1 ORDER BY created_at DESC`

	updateFileVersion = `
UPDATE files SET version = version + 1, updated_at = datetime('now') WHERE id = $1`

	deleteFile = `DELETE FROM files WHERE id = $1`

	deleteFilesByChatID = `DELETE FROM files WHERE chat_id = $1`
)

// Read file queries
const (
	recordReadFile = `
INSERT OR REPLACE INTO read_files (chat_id, path, read_at) VALUES ($1, $2, datetime('now'))`

	getReadFilesByChatID = `
SELECT chat_id, path, read_at FROM read_files WHERE chat_id = $1 ORDER BY read_at DESC`

	getRecentlyReadFiles = `
SELECT path, MAX(read_at) as read_at FROM read_files WHERE chat_id = $1 GROUP BY path ORDER BY read_at DESC LIMIT $2`

	deleteReadFilesByChatID = `DELETE FROM read_files WHERE chat_id = $1`
)