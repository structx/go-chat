
-- name: InsertConversation :one
-- add conversation to database
INSERT INTO conversations (uuid)
VALUES (
    ?
) RETURNING *;

-- name: ReadAllConversations :many
-- retrieve all conversations that includes user uuid in recipients field
SELECT *
FROM conversations
JOIN mm_conversations_users
    ON conversation.uuid = mm_conversations_users.conversation_uuid
WHERE mm_conversations_users.user_uuid = ?;

-- name: InsertMessage :one
-- add new message to database
INSERT INTO messages (uuid, conversation_uuid, sender, body)
VALUES (
    ?, ?, ?, ?
) RETURNING *;

-- name: ReadAllMessages :many
-- retrieve all messages by conversation uuid
SELECT *
FROM messages
WHERE conversation_uuid = ?;
