// Package domain application layer
package domain

import "database/sql"

// Bundle domain service bundle
type Bundle struct {
	UserService      *UserService
	MessengerService *MessengerService
	ContactService   *ContactService
}

// NewBundle create new service bundle
func NewBundle(db *sql.DB) *Bundle {
	return &Bundle{
		UserService:      newUserService(db),
		MessengerService: newMessengerService(db),
		ContactService:   newContactService(db),
	}
}
