// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.18.0

package repository

import (
	"database/sql"
	"time"
)

type Contact struct {
	Uuid          string
	OriginUuid    string
	RecipientUuid string
	CreatedAt     time.Time
}

type Conversation struct {
	Uuid      string
	CreatedAt time.Time
	UpdatedAt sql.NullTime
}

type Message struct {
	Uuid             string
	ConversationUuid string
	Sender           string
	Body             string
	CreatedAt        time.Time
}

type MmConversationsUser struct {
	Uuid             string
	ConversationUuid string
	UserUuid         string
}

type User struct {
	Uuid      string
	Usernm    string
	Email     string
	CreatedAt time.Time
	UpdatedAt sql.NullTime
	Pssword   string
}
