-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, user_name)
VALUES (
    gen_random_uuid(),
    Now(),
    Now(),
    $1
)
RETURNING *;

-- name: DeleteUsers :exec
DELETE FROM users;