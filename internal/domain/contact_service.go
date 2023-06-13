package domain

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/trevatk/go-chat/internal/repository"
)

// NewContact application layer new contact model
type NewContact struct {
	Owner     uuid.UUID
	Recipient uuid.UUID
}

// Contact application layer contact model
type Contact struct {
	UID       uuid.UUID
	Owner     uuid.UUID
	Recipient uuid.UUID
	CreatedAt time.Time
}

// ContactService contact management service
type ContactService struct {
	db *sql.DB
}

func newContactService(db *sql.DB) *ContactService {
	return &ContactService{db: db}
}

// Create add contact record to database
func (cs *ContactService) Create(ctx context.Context, newContact *NewContact) (*Contact, error) {

	uid := uuid.New()

	co, e := cs.db.Conn(ctx)
	if e != nil {
		return nil, fmt.Errorf("unable to pull database connection from pool %v", e)
	}
	defer func() { _ = co.Close() }()

	c, e := repository.New(co).InsertContact(ctx, &repository.InsertContactParams{
		Uuid:          uid.String(),
		OriginUuid:    newContact.Owner.String(),
		RecipientUuid: newContact.Recipient.String(),
	})
	if e != nil {
		return nil, fmt.Errorf("failed to add contact to database %v", e)
	}

	return transformSQLContact(c), nil
}

// Read retrieve a single contact record by uuid
func (cs *ContactService) Read(ctx context.Context, uid uuid.UUID) (*Contact, error) {

	co, e := cs.db.Conn(ctx)
	if e != nil {
		return nil, fmt.Errorf("unable to pull database connection from pool %v", e)
	}
	defer func() { _ = co.Close() }()

	sc, e := repository.New(co).ReadContact(ctx, uid.String())
	if e != nil {

		if errors.Is(e, sql.ErrNoRows) {
			return &Contact{}, ErrResourceNotFound
		}

		return nil, fmt.Errorf("error executing read contact query %v", e)
	}

	return transformSQLContact(sc), nil
}

// List retrieve all contacts where owner matches uuid
func (cs *ContactService) List(ctx context.Context, uuid uuid.UUID) ([]*Contact, error) {

	co, e := cs.db.Conn(ctx)
	if e != nil {
		return nil, fmt.Errorf("unable to pull database connection from pool %v", e)
	}
	defer func() { _ = co.Close() }()

	scl, e := repository.New(co).ReadAllContacts(ctx, uuid.String())
	if e != nil {

		if errors.Is(e, sql.ErrNoRows) {
			return []*Contact{}, ErrResourceNotFound
		}

		return nil, fmt.Errorf("error excuting read all contacts query %v", e)
	}

	cl := make([]*Contact, 0, len(scl))

	for _, sc := range scl {
		cl = append(cl, transformSQLContact(sc))
	}

	return cl, nil
}

// Search query for all contacts by username or email
func (cs *ContactService) Search(ctx context.Context, search string) ([]*Contact, error) {

	co, e := cs.db.Conn(ctx)
	if e != nil {
		return nil, fmt.Errorf("unable to pull database connection from pool %v", e)
	}
	defer func() { _ = co.Close() }()

	scs, e := repository.New(co).SearchContacts(ctx, &repository.SearchContactsParams{
		Usernm: "%" + strings.ToUpper(search) + "%",
		Email:  "%" + strings.ToUpper(search) + "%",
	})
	if e != nil {

		if errors.Is(e, sql.ErrNoRows) {
			return []*Contact{}, nil
		}

		return nil, fmt.Errorf("error executing search contacts query %v", e)
	}

	dcs := make([]*Contact, 0, len(scs))

	for _, sc := range scs {
		dcs = append(dcs, &Contact{
			UID:       uuid.MustParse(sc.Uuid),
			Owner:     uuid.MustParse(sc.OriginUuid),
			Recipient: uuid.MustParse(sc.RecipientUuid),
		})
	}

	return dcs, nil
}

// Delete hard delete contact from database
func (cs *ContactService) Delete(ctx context.Context, uuid uuid.UUID) error {

	co, e := cs.db.Conn(ctx)
	if e != nil {
		return fmt.Errorf("unable to pull database connection from pool %v", e)
	}
	defer func() { _ = co.Close() }()

	r, e := repository.New(co).DeleteContact(ctx, uuid.String())
	if e != nil {
		return fmt.Errorf("unable to delete contact %v", e)
	}

	af, e := r.RowsAffected()
	if e != nil {
		return fmt.Errorf("failed to get rows affected %v", e)
	}

	if af < 1 {
		return ErrResourceNotFound
	}

	return nil
}

func transformSQLContact(contact *repository.Contact) *Contact {
	return &Contact{
		UID:       uuid.MustParse(contact.Uuid),
		Owner:     uuid.MustParse(contact.OriginUuid),
		Recipient: uuid.MustParse(contact.RecipientUuid),
		CreatedAt: contact.CreatedAt,
	}
}
