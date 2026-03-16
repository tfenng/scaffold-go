-- name: GetUserByID :one
SELECT id, uid, email, name, used_name, company, birth, created_at, updated_at
FROM users
WHERE id = $1;

-- name: GetUserByUID :one
SELECT id, uid, email, name, used_name, company, birth, created_at, updated_at
FROM users
WHERE uid = $1;

-- name: GetUserByEmail :one
SELECT id, uid, email, name, used_name, company, birth, created_at, updated_at
FROM users
WHERE email = $1;

-- name: CreateUser :one
INSERT INTO users (uid, name, email, used_name, company, birth)
VALUES ($1, $2, sqlc.narg('email')::text, $3, $4, $5)
RETURNING id, uid, email, name, used_name, company, birth, created_at, updated_at;

-- name: ListUsers :many
SELECT id, uid, email, name, used_name, company, birth, created_at, updated_at
FROM users
WHERE (sqlc.narg('email')::text IS NULL OR email = sqlc.narg('email')::text)
  AND (sqlc.narg('name_like')::text IS NULL OR name ILIKE ('%' || sqlc.narg('name_like')::text || '%'))
ORDER BY created_at DESC, id DESC
LIMIT $1 OFFSET $2;

-- name: CountUsers :one
SELECT COUNT(1)
FROM users
WHERE (sqlc.narg('email')::text IS NULL OR email = sqlc.narg('email')::text)
  AND (sqlc.narg('name_like')::text IS NULL OR name ILIKE ('%' || sqlc.narg('name_like')::text || '%'));

-- name: UpdateUser :one
UPDATE users
SET name = $2, email = sqlc.narg('email')::text, used_name = $3, company = $4, birth = $5, updated_at = now()
WHERE id = $1
RETURNING id, uid, email, name, used_name, company, birth, created_at, updated_at;

-- name: DeleteUser :exec
DELETE FROM users WHERE id = $1;
