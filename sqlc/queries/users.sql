
-- name: InsertUser :one
-- add user to database
INSERT INTO users (uuid, usernm, email)
VALUES (
    ?, ?, ?
) RETURNING *;

-- name: ReadUserDetails :one
-- read user details by uuid
SELECT usernm, email
FROM users
WHERE uuid = ?;

-- name: SearchUserDetails :many
-- find all users who match search query
SELECT uuid, usernm
FROM users
WHERE usernm LIKE ? 
    OR email LIKE ?
LIMIT 10;

-- name: ReadUser :one
-- read user by uuid
SELECT *
FROM users
WHERE uuid = ?;

-- name: UpdateUser :one
UPDATE users
SET
    usernm = ?,
    email = ?,
    updated_at = CURRENT_TIMESTAMP
WHERE uuid = ?
RETURNING *;