-- name: CreateUser :one
INSERT INTO users (first_name, last_name, email, password_hash, github_link)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1;

-- name: UpdateUser :one
UPDATE users
SET first_name = $2,
    last_name = $3,
    email = $4,
    github_link = $5,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: CountUsers :one
SELECT COUNT(*) FROM users;
