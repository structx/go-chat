package domain

import "errors"

var (
	// ErrResourceNotFound no resource returned by query using id
	ErrResourceNotFound = errors.New("no resource found by id")
	// ErrUniqueExists request username already exists in database
	ErrUniqueExists = errors.New("requested username/email is already assigned")
	// ErrEmptyResult query was successful but returned no records
	ErrEmptyResult = errors.New("empty result set")
	// ErrMinRecipients less than two recipients provided for a new conversation
	ErrMinRecipients = errors.New("invalid number recipients")
)
