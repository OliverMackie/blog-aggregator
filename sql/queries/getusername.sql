-- name: GetUserName :one
SELECT *
FROM users
WHERE id = $1;