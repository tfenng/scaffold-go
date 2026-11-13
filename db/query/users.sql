-- name: CreateUser :one
INSERT INTO users (uid, name, email, used_name, company, birth)
VALUES (
  sqlc.arg('uid'),
  sqlc.arg('name'),
  sqlc.narg('email'),
  sqlc.arg('used_name'),
  sqlc.arg('company'),
  sqlc.narg('birth')
)
RETURNING id, uid, email, name, used_name, company, birth, created_at, updated_at;

-- name: GetUserByID :one
SELECT id, uid, email, name, used_name, company, birth, created_at, updated_at
FROM users
WHERE id = sqlc.arg('id');

-- name: ListUsers :many
SELECT id, uid, email, name, used_name, company, birth, created_at, updated_at
FROM users
WHERE (CAST(sqlc.narg('email') AS TEXT) IS NULL OR email = CAST(sqlc.narg('email') AS TEXT))
  AND (CAST(sqlc.narg('name_like') AS TEXT) IS NULL OR name LIKE '%' || CAST(sqlc.narg('name_like') AS TEXT) || '%')
ORDER BY created_at DESC, id DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: CountUsers :one
SELECT COUNT(1)
FROM users
WHERE (CAST(sqlc.narg('email') AS TEXT) IS NULL OR email = CAST(sqlc.narg('email') AS TEXT))
  AND (CAST(sqlc.narg('name_like') AS TEXT) IS NULL OR name LIKE '%' || CAST(sqlc.narg('name_like') AS TEXT) || '%');

-- name: UpdateUser :one
UPDATE users
SET name = sqlc.arg('name'),
    email = sqlc.narg('email'),
    used_name = sqlc.arg('used_name'),
    company = sqlc.arg('company'),
    birth = sqlc.narg('birth'),
    updated_at = strftime('%Y-%m-%dT%H:%M:%SZ', 'now')
WHERE id = sqlc.arg('id')
RETURNING id, uid, email, name, used_name, company, birth, created_at, updated_at;

-- name: DeleteUser :execrows
DELETE FROM users
WHERE id = sqlc.arg('id');
