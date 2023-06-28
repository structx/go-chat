CREATE TABLE IF NOT EXISTS users (
    uuid VARCHAR(36) PRIMARY KEY,
    usernm VARCHAR(50) NOT NULL UNIQUE,
    email VARCHAR(50) NOT NULL UNIQUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP
);

CREATE TABLE IF NOT EXISTS contacts (
    uuid VARCHAR(36) PRIMARY KEY,
    origin_uuid VARCHAR(36) NOT NULL,
    recipient_uuid VARCHAR(36) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    FOREIGN KEY (origin_uuid) REFERENCES users (uuid),
    FOREIGN KEY (recipient_uuid) REFERENCES users (uuid)
);

CREATE TABLE IF NOT EXISTS conversations (
    uuid VARCHAR(36) PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP
);

CREATE TABLE IF NOT EXISTS mm_conversations_users (
    uuid VARCHAR(36) PRIMARY KEY,
    conversation_uuid VARCHAR(36) NOT NULL,
    user_uuid VARCHAR(36) NOT NULL,
    FOREIGN KEY (conversation_uuid) REFERENCES conversations(uuid),
    FOREIGN KEY (user_uuid) REFERENCES users(user_uuid)
);

CREATE TABLE IF NOT EXISTS messages (
    uuid VARCHAR(36) PRIMARY KEY,
    conversation_uuid VARCHAR(36) NOT NULL,
    sender VARCHAR(36) NOT NULL,
    body TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    FOREIGN KEY (conversation_uuid) REFERENCES conversations (uuid)
);