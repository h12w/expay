package expay

import "errors"

var (
	// ErrNotFound is returned when an item is not found in the DB
	ErrNotFound = errors.New("item not found")
)

// verification errors
//
// TODO: add specific errors here
var (
	// ErrInvalidPayment is a general verification error for payment format
	ErrInvalidPayment = errors.New("invalid payment")
)
