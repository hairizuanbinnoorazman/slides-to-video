package user

import "context"

type Store interface {
	StoreUser(ctx context.Context, u User) error
	GetUser(ctx context.Context, ID string) (User, error)
	GetUserByEmail(ctx context.Context, Email string) (User, error)
}
