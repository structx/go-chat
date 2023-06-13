
-- name: InsertContact :one
-- add contact to database
INSERT INTO contacts (uuid, origin_uuid, recipient_uuid)
VALUES (
    ?, ?, ?
) RETURNING *;

-- name: ReadContact :one
SELECT *
FROM contacts
WHERE uuid = ?;

-- name: ReadAllContacts :many
-- retrieve all contacts assigned to specific user
SELECT *
FROM contacts
WHERE origin_uuid = ?;

-- name: SearchContacts :many
-- retrieve all contacts by name and email
SELECT *
FROM contacts 
JOIN users
    ON contacts.recipient_uuid = users.uuid
WHERE users.usernm LIKE ?
    OR users.email LIKE ?
LIMIT 10;

-- name: DeleteContact :execresult
-- hard delete contact 
DELETE FROM contacts
WHERE uuid = ?;