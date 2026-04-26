-- name: RecordReadFile :exec
INSERT OR REPLACE INTO read_files (chat_id, path, read_at)
VALUES ($1, $2, datetime('now'));

-- name: GetReadFilesByChatID :many
SELECT * FROM read_files 
WHERE chat_id = $1 
ORDER BY read_at DESC;

-- name: GetRecentlyReadFiles :many
SELECT DISTINCT path, MAX(read_at) as read_at 
FROM read_files 
WHERE chat_id = $1 
GROUP BY path 
ORDER BY read_at DESC 
LIMIT $2;

-- name: DeleteReadFilesByChatID :exec
DELETE FROM read_files WHERE chat_id = $1;