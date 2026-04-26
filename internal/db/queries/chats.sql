-- name: CreateChat :one
INSERT INTO chats (
    id, parent_chat_id, title, metadata, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, datetime('now'), datetime('now')
) RETURNING *;

-- name: GetChatByID :one
SELECT * FROM chats WHERE id = $1;

-- name: GetAllChats :many
SELECT * FROM chats ORDER BY updated_at DESC;

-- name: GetChatsByParentID :many
SELECT * FROM chats WHERE parent_chat_id = $1 ORDER BY created_at ASC;

-- name: UpdateChatTitle :exec
UPDATE chats SET title = $2, updated_at = datetime('now') WHERE id = $1;

-- name: UpdateChatMetadata :exec
UPDATE chats SET metadata = $2, updated_at = datetime('now') WHERE id = $1;

-- name: UpdateChatStats :exec
UPDATE chats SET 
    prompt_tokens = prompt_tokens + $2,
    completion_tokens = completion_tokens + $3,
    cost = cost + $4,
    updated_at = datetime('now')
WHERE id = $1;

-- name: SetChatSummary :exec
UPDATE chats SET summary_message_id = $2, updated_at = datetime('now') WHERE id = $1;

-- name: DeleteChat :exec
DELETE FROM chats WHERE id = $1;