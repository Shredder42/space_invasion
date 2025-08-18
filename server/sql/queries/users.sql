-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, user_name, hashed_password)
VALUES (
    gen_random_uuid(),
    Now(),
    Now(),
    $1,
    $2
)
RETURNING *;

-- name: DeleteUsers :exec
DELETE FROM users;

-- name: GetUserByUsername :one
SELECT *
FROM users
WHERE user_name = $1;