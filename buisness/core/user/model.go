package user

import (
	"net/mail"
	"time"

	"github.com/google/uuid"
)

// Data model name should match the package name
// User represents information about an individual user
type User struct {
	ID           uuid.UUID
	Name         string
	Email        mail.Address
	Roles        []Role
	PasswordHash []byte
	Department   string
	Enabled      bool
	DateCreated  time.Time
	DateUpdated  time.Time
}

// NewUser contains information needed to create a new user.
type NewUser struct {
	Name            string
	Email           mail.Address
	Roles           []Role
	Department      string
	Password        string
	PasswordConfirm string
}

// UpdateUser contains information needed to update a user.
// You have to make some choices about update
// Updates are really hard, especially saying a relational DB
// Where you got 2 or 3 users doing update at the same time
// We are using pointer semantics here as a way of describing the concept
// of null
type UpdateUser struct {
	Name            *string
	Email           *mail.Address
	Roles           []Role
	Department      *string
	Password        *string
	PasswordConfirm *string
	Enabled         *bool
}
