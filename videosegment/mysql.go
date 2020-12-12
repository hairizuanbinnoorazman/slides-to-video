package videosegment

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

func (m mysql) Create(ctx context.Context, e VideoSegment) error {
	result := m.db.Create(&e)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (m mysql) Get(ctx context.Context, projectID, ID string) (VideoSegment, error) {
	p := VideoSegment{}
	result := m.db.Where("id = ? AND project_id = ?", ID, projectID).First(&p)
	if result.Error != nil {
		return p, result.Error
	}
	return p, nil
}

func (m mysql) GetAll(ctx context.Context, ProjectID string, Limit, After int) ([]VideoSegment, error) {
	var videosegments []VideoSegment
	result := m.db.Where("project_id = ?", ProjectID).Limit(Limit).Offset(After).Find(&videosegments)
	if result.Error != nil {
		return []VideoSegment{}, result.Error
	}
	return videosegments, nil
}

func (m mysql) Update(ctx context.Context, projectID, ID string, setters ...func(*VideoSegment) error) (VideoSegment, error) {
	var p VideoSegment
	result := m.db.Where("project_id = ? AND id = ?", projectID, ID).First(&p)
	if result.Error != nil {
		return VideoSegment{}, result.Error
	}
	for _, s := range setters {
		err := s(&p)
		if err != nil {
			return VideoSegment{}, err
		}
	}
	result = m.db.Save(&p)
	if result.Error != nil {
		return VideoSegment{}, result.Error
	}
	return p, nil
}

func (m mysql) Delete(ctx context.Context, projectID, ID string) error {
	result := m.db.Where("id = ? and project_id = ?", ID, projectID).Delete(VideoSegment{})
	if result.Error != nil {
		return result.Error
	}
	return nil
}
