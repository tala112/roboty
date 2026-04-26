-- name: CreateMessage :one
INSERT INTO messages (
    id, chat_id, role, content, model, provider, is_summary, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, datetime('now'), datetime('now')
) RETURNING *;

-- name: GetMessageByID :one
SELECT * FROM messages WHERE id = $1;

-- name: GetMessagesByChatID :many
SELECT * FROM messages 
WHERE chat_id = $1 
ORDER BY created_at ASC;

-- name: GetMessagesByChatIDPaginated :many
SELECT * FROM messages 
WHERE chat_id = $1 
ORDER BY created_at ASC 
LIMIT $2 OFFSET $3;

-- name: GetChatLatestMessage :one
SELECT * FROM messages 
WHERE chat_id = $1 
ORDER BY created_at DESC 
LIMIT 1;

-- name: UpdateMessageContent :exec
UPDATE messages SET content = $2, updated_at = datetime('now') WHERE id = $1;

-- name: UpdateMessageFinishedAt :exec
UPDATE messages SET finished_at = datetime('now') WHERE id = $1;

-- name: DeleteMessage :exec
DELETE FROM messages WHERE id = $1;

-- name: DeleteMessagesByChatID :exec
DELETE FROM messages WHERE chat_id = $1;

-- name: CountMessagesByChatID :one
SELECT COUNT(*) FROM messages WHERE chat_id = $1;