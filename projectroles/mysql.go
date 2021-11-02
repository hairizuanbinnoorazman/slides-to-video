package projectroles

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

func (m mysql) Create(ctx context.Context, e ProjectRole) error {
	result := m.db.Create(&e)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (m mysql) Get(ctx context.Context, ProjectID, EntityID string) (ProjectRole, error) {
	p := ProjectRole{}
	result := m.db.Where("project_id = ? AND entity_id = ?", ProjectID, EntityID).First(&p)
	if result.Error != nil {
		return p, result.Error
	}
	return p, nil
}

func (m mysql) GetAll(ctx context.Context, ProjectID string, Limit, After int) ([]ProjectRole, error) {
	var projects []ProjectRole
	result := m.db.Where("project_id = ?", ProjectID).Limit(Limit).Offset(After).Find(&projects)
	if result.Error != nil {
		return []ProjectRole{}, result.Error
	}
	return projects, nil
}

func (m mysql) Update(ctx context.Context, ProjectID, EntityID string, setters ...func(*ProjectRole) error) (ProjectRole, error) {
	var p ProjectRole
	result := m.db.Where("project_id = ? AND entity_id = ?", ProjectID, EntityID).First(&p)
	if result.Error != nil {
		return ProjectRole{}, result.Error
	}
	for _, s := range setters {
		err := s(&p)
		if err != nil {
			return ProjectRole{}, err
		}
	}
	result = m.db.Save(&p)
	if result.Error != nil {
		return ProjectRole{}, result.Error
	}
	return p, nil
}

func (m mysql) Delete(ctx context.Context, ProjectID, EntityID string) error {
	result := m.db.Where("project_id = ? AND entity_id = ?", ProjectID, EntityID).Delete(ProjectRole{})
	if result.Error != nil {
		return result.Error
	}
	return nil
}
