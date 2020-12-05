package user

import (
	"context"

	"github.com/jinzhu/gorm"
)

type mysql struct {
	db *gorm.DB
}

func (m mysql) StoreUser(ctx context.Context, u User) error {
	return nil
}

func (m mysql) GetUser(ctx context.Context, ID string) (User, error) {
	return User{}, nil
}

func (m mysql) GetUserByEmail(ctx context.Context, Email string) (User, error) {
	return User{}, nil
}
