package acl

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

func (m mysql) Create(ctx context.Context, e ACL) error {
	result := m.db.Create(&e)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (m mysql) Get(ctx context.Context, ProjectID, UserID string) (ACL, error) {
	p := ACL{}
	result := m.db.Where("project_id = ? AND user_id = ?", ProjectID, UserID).First(&p)
	if result.Error != nil {
		return p, result.Error
	}
	return p, nil
}

func (m mysql) GetAll(ctx context.Context, ProjectID string, Limit, After int) ([]ACL, error) {
	var projects []ACL
	result := m.db.Where("project_id = ?", ProjectID).Limit(Limit).Offset(After).Find(&projects)
	if result.Error != nil {
		return []ACL{}, result.Error
	}
	return projects, nil
}

func (m mysql) Update(ctx context.Context, ProjectID, UserID string, setters ...func(*ACL) error) (ACL, error) {
	var p ACL
	result := m.db.Where("project_id = ? AND user_id = ?", ProjectID, UserID).First(&p)
	if result.Error != nil {
		return ACL{}, result.Error
	}
	for _, s := range setters {
		err := s(&p)
		if err != nil {
			return ACL{}, err
		}
	}
	result = m.db.Save(&p)
	if result.Error != nil {
		return ACL{}, result.Error
	}
	return p, nil
}

func (m mysql) Delete(ctx context.Context, ProjectID, UserID string) error {
	result := m.db.Where("project_id = ? AND user_id = ?", ProjectID, UserID).Delete(ACL{})
	if result.Error != nil {
		return result.Error
	}
	return nil
}
