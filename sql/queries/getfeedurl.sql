-- name: GetFeedUrl :one
SELECT *
FROM feeds
WHERE url = $1;