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

-- name: GetChirpsDortedByCreateTime :many
SELECT * FROM chirp ORDER BY created_at;

-- name: GetChirpById :one
SELECT * FROM chirp WHERE id = $1;
