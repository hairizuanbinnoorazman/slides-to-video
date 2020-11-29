// Package user contains all the functionality that relates to user management
package user

import (
	"context"
	"time"
)

// User represents a single user - current assumes that user would be one that a google account
type User struct {
	ID           string
	Email        string
	RefreshToken string
	AuthToken    string
	Type         string
	DateCreated  time.Time
	DateModified time.Time
}

type Store interface {
	StoreUser(ctx context.Context, u User) error
	GetUser(ctx context.Context, ID string) (User, error)
	GetUserByEmail(ctx context.Context, Email string) (User, error)
}
