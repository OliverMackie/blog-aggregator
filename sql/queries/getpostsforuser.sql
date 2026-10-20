-- name: GetPostsForUser :many
SELECT posts.*
FROM posts
INNER JOIN feed_follows ON posts.feed_id = feed_follows.feed_id
INNER JOIN users ON feed_follows.user_id = users.id
WHERE users.name = $1
ORDER BY posts.published_at DESC
LIMIT $2;