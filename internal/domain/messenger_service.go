package domain

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/trevatk/go-chat/internal/repository"
)

// NewConversation application layer new conversation model
type NewConversation struct {
	Recipients []uuid.UUID
}

// Conversation application layer conversation model
type Conversation struct {
	UID       uuid.UUID
	CreatedAt time.Time
	UpdatedAt time.Time
}

// NewEnvelope application layer new envelope model
type NewEnvelope struct {
	Sender           uuid.UUID
	Message          string
	ConversationUUID uuid.UUID
}

// Envelope application layer envelope model
type Envelope struct {
	UID              uuid.UUID
	Sender           uuid.UUID
	Message          string
	ConversationUUID uuid.UUID
	CreatedAt        time.Time
}

// MessengerService messenger management service
type MessengerService struct {
	db *sql.DB
}

func newMessengerService(db *sql.DB) *MessengerService {
	return &MessengerService{db: db}
}

// CreateConversation add new conversation to database
func (ms *MessengerService) CreateConversation(ctx context.Context, newConversation *NewConversation) (*Conversation, error) {

	uid := uuid.New()

	co, e := ms.db.Conn(ctx)
	if e != nil {
		return nil, fmt.Errorf("failed to get database connection from pool %v", e)
	}
	defer func() { _ = co.Close() }()

	tx, e := co.BeginTx(ctx, nil)
	if e != nil {
		return nil, fmt.Errorf("unable to begin transaction %v", e)
	}
	defer func() { _ = tx.Rollback() }()

	if len(newConversation.Recipients) < 2 {
		return nil, ErrMinRecipients
	}

	q := repository.New(co).WithTx(tx)

	c, e := q.InsertConversation(ctx, uid.String())
	if e != nil {
		return nil, fmt.Errorf("unable to add conversation to database %v", e)
	}

	for _, r := range newConversation.Recipients {

		cuID := uuid.New()

		r, e := q.InsertMMConversationUser(ctx, &repository.InsertMMConversationUserParams{
			Uuid:             cuID.String(),
			ConversationUuid: uid.String(),
			UserUuid:         r.String(),
		})
		if e != nil {
			return nil, fmt.Errorf("unable to add users to conversation %v", e)
		}

		if af, e := r.RowsAffected(); e != nil || af < 1 {
			return nil, fmt.Errorf("no rows affected %v", e)
		}
	}

	e = tx.Commit()
	if e != nil {
		return nil, fmt.Errorf("failed to commit transaction %v", e)
	}

	return transformSQLConversation(c), nil
}

// ListConversations retrieve all conversations by recipient uuid
func (ms *MessengerService) ListConversations(ctx context.Context, uid uuid.UUID) ([]*Conversation, error) {

	co, e := ms.db.Conn(ctx)
	if e != nil {
		return nil, fmt.Errorf("failed to get database connection from pool %v", e)
	}
	defer func() { _ = co.Close() }()

	scl, e := repository.New(co).ReadAllConversations(ctx, uid.String())
	if e != nil {

		return nil, fmt.Errorf("error excuting read all conversations query %v", e)
	}

	fmt.Println(scl)

	if len(scl) == 0 {
		return []*Conversation{}, ErrEmptyResult
	}

	cl := make([]*Conversation, 0, len(scl))

	for _, sc := range scl {
		cl = append(cl, transformReadAllConversationsRow(sc))
	}

	return cl, nil
}

// CreateMessage add new envelope into database as message model
func (ms *MessengerService) CreateMessage(ctx context.Context, newEnvelope *NewEnvelope) (*Envelope, error) {

	uid := uuid.New()

	co, e := ms.db.Conn(ctx)
	if e != nil {
		return nil, fmt.Errorf("failed to get database connection from pool %v", e)
	}
	defer func() { _ = co.Close() }()

	m, e := repository.New(co).InsertMessage(ctx, &repository.InsertMessageParams{
		Uuid:             uid.String(),
		ConversationUuid: newEnvelope.ConversationUUID.String(),
		Sender:           newEnvelope.Sender.String(),
		Body:             newEnvelope.Message,
	})
	if e != nil {
		return nil, fmt.Errorf("error executing insert message query %v", e)
	}

	return transformSQLMessage(m), nil
}

// ListMessages retrive all messages by conversation uuid
func (ms *MessengerService) ListMessages(ctx context.Context, conversationUUID uuid.UUID) ([]*Envelope, error) {

	co, e := ms.db.Conn(ctx)
	if e != nil {
		return nil, fmt.Errorf("failed to get database connection from pool %v", e)
	}
	defer func() { _ = co.Close() }()

	sml, e := repository.New(co).ReadAllMessages(ctx, conversationUUID.String())
	if e != nil {
		return nil, fmt.Errorf("error executing read all messages %v", e)
	}

	el := make([]*Envelope, 0, len(sml))

	for _, sm := range sml {
		el = append(el, transformSQLMessage(sm))
	}

	return el, nil
}

func transformSQLConversation(conv *repository.Conversation) *Conversation {

	u := time.Time{}

	if conv.UpdatedAt.Valid {
		u = conv.UpdatedAt.Time
	}

	return &Conversation{
		UID:       uuid.MustParse(conv.Uuid),
		CreatedAt: conv.CreatedAt,
		UpdatedAt: u,
	}
}

func transformReadAllConversationsRow(row *repository.ReadAllConversationsRow) *Conversation {

	cID := uuid.MustParse(row.Uuid)

	u := time.Time{}
	if row.UpdatedAt.Valid {
		u = row.UpdatedAt.Time
	}

	return &Conversation{
		UID:       cID,
		UpdatedAt: u,
	}
}

func transformSQLMessage(message *repository.Message) *Envelope {
	return &Envelope{
		UID:              uuid.MustParse(message.Uuid),
		Sender:           uuid.MustParse(message.Sender),
		Message:          message.Body,
		ConversationUUID: uuid.MustParse(message.ConversationUuid),
		CreatedAt:        message.CreatedAt,
	}
}
