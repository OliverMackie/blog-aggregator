-- name: DeleteFeedFollow :exec
DELETE FROM feed_follows
WHERE user_id = (SELECT id FROM users WHERE users.name = $1)
AND feed_id = (SELECT id FROM feeds WHERE feeds.url = $2);