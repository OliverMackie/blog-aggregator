-- name: CreateFeedFollow :one
WITH inserted_values AS (INSERT INTO feed_follows (id, created_at, updated_at, user_id, feed_id)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5
)
RETURNING *)
SELECT inserted_values.*, users.name AS user_name, feeds.name AS feed_name
FROM inserted_values
INNER JOIN users ON inserted_values.user_id = users.id
INNER JOIN feeds ON inserted_values.feed_id = feeds.id;