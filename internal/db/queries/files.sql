-- name: CreateFile :one
INSERT INTO files (
    id, chat_id, path, version, created_at, updated_at
) VALUES (
    $1, $2, $3, 1, datetime('now'), datetime('now')
) RETURNING *;

-- name: GetFileByID :one
SELECT * FROM files WHERE id = $1;

-- name: GetFileByChatAndPath :one
SELECT * FROM files WHERE chat_id = $1 AND path = $2;

-- name: GetFilesByChatID :many
SELECT * FROM files WHERE chat_id = $1 ORDER BY created_at DESC;

-- name: UpdateFileVersion :exec
UPDATE files SET version = version + 1, updated_at = datetime('now') WHERE id = $1;

-- name: DeleteFile :exec
DELETE FROM files WHERE id = $1;

-- name: DeleteFilesByChatID :exec
DELETE FROM files WHERE chat_id = $1;