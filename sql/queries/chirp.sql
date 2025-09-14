-- name: CreateChirp :one
INSERT INTO chirp (id, created_at, updated_at, body, user_id)
VALUES (
    gen_random_uuid(),
    NOW(),
    NOW(),
    $1,
    $2
)
RETURNING *;

-- name: GetChirpsSortedByCreateTime :many
SELECT * FROM chirp ORDER BY created_at;

-- name: GetChirpsSortedByCreateTimeForAuthor :many
SELECT * FROM chirp WHERE user_id = $1 ORDER BY created_at;

-- name: GetChirpById :one
SELECT * FROM chirp WHERE id = $1;

-- name: DeleteChirpById :exec
DELETE FROM chirp
WHERE id = $1;
