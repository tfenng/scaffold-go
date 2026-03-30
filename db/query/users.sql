-- name: CreateUser :one
INSERT INTO users (uid, name, email, used_name, company, birth, created_at, updated_at)
VALUES (
  sqlc.arg('uid'),
  sqlc.arg('name'),
  sqlc.narg('email')::text,
  sqlc.arg('used_name'),
  sqlc.arg('company'),
  sqlc.narg('birth')::date,
  NOW(),
  NOW()
)
RETURNING id, uid, email, name, used_name, company, birth, created_at, updated_at;

-- name: GetUserByID :one
SELECT id, uid, email, name, used_name, company, birth, created_at, updated_at
FROM users
WHERE id = sqlc.arg('id');

-- name: ListUsers :many
SELECT id, uid, email, name, used_name, company, birth, created_at, updated_at
FROM users
WHERE (sqlc.narg('email')::text IS NULL OR email = sqlc.narg('email')::text)
  AND (sqlc.narg('name_like')::text IS NULL OR name ILIKE '%' || sqlc.narg('name_like')::text || '%')
ORDER BY created_at DESC, id DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: CountUsers :one
SELECT COUNT(1)
FROM users
WHERE (sqlc.narg('email')::text IS NULL OR email = sqlc.narg('email')::text)
  AND (sqlc.narg('name_like')::text IS NULL OR name ILIKE '%' || sqlc.narg('name_like')::text || '%');

-- name: UpdateUser :one
UPDATE users
SET name = sqlc.arg('name'),
    email = sqlc.narg('email')::text,
    used_name = sqlc.arg('used_name'),
    company = sqlc.arg('company'),
    birth = sqlc.narg('birth')::date,
    updated_at = NOW()
WHERE id = sqlc.arg('id')
RETURNING id, uid, email, name, used_name, company, birth, created_at, updated_at;

-- name: DeleteUser :execrows
DELETE FROM users
WHERE id = sqlc.arg('id');
