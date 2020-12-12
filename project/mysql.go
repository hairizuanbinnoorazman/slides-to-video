package project

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

func (m mysql) Create(ctx context.Context, e Project) error {
	result := m.db.Create(&e)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (m mysql) Get(ctx context.Context, ID string, UserID string) (Project, error) {
	p := Project{}
	result := m.db.Where("id = ?", ID).First(&p)
	if result.Error != nil {
		return p, result.Error
	}
	return p, nil
}

func (m mysql) GetAll(ctx context.Context, UserID string, Limit, After int) ([]Project, error) {
	var projects []Project
	result := m.db.Limit(Limit).Offset(After).Find(&projects)
	if result.Error != nil {
		return []Project{}, result.Error
	}
	return projects, nil
}

func (m mysql) Update(ctx context.Context, ID string, UserID string, setters ...func(*Project) error) (Project, error) {
	var p Project
	result := m.db.Where("id = ?", ID).First(&p)
	if result.Error != nil {
		return Project{}, result.Error
	}
	for _, s := range setters {
		err := s(&p)
		if err != nil {
			return Project{}, err
		}
	}
	return p, nil
}

func (m mysql) Delete(ctx context.Context, ID string, UserID string) error {
	result := m.db.Where("id = ?", ID).Delete(Project{})
	if result.Error != nil {
		return result.Error
	}
	return nil
}
