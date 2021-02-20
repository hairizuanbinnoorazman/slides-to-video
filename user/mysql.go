package user

import (
	"context"
	"errors"

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
	result := m.db.Create(&u)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (m mysql) GetUser(ctx context.Context, ID string) (User, error) {
	u := User{}
	result := m.db.Where("id = ?", ID).First(&u)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return User{}, nil
	}
	if result.Error != nil {
		return User{}, result.Error
	}
	return u, nil
}

func (m mysql) GetUserByEmail(ctx context.Context, Email string) (User, error) {
	u := User{}
	result := m.db.Where("email = ?", Email).First(&u)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return User{}, nil
	}
	if result.Error != nil {
		return User{}, result.Error
	}
	return u, nil
}
