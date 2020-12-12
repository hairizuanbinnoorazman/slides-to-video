package user

import (
	"context"

	"github.com/hairizuanbinnoorazman/slides-to-video-manager/logger"
	"github.com/jinzhu/gorm"
)

type mysql struct {
	db     *gorm.DB
	logger logger.Logger
}

func NewMySQL(logger logger.Logger, dbClient *gorm.DB) mysql {
	return mysql{
		db:     dbClient,
		logger: logger,
	}
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
