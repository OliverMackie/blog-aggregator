-- name: Reset :exec
TRUNCATE TABLE users, feeds, feed_follows, posts RESTART IDENTITY CASCADE;