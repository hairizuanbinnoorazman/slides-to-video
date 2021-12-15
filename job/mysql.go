package job

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

func (m mysql) Create(ctx context.Context, e Job) error {
	result := m.db.Create(&e)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (m mysql) Get(ctx context.Context, ID string) (Job, error) {
	p := Job{}
	result := m.db.Where("id = ?", ID).First(&p)
	if result.Error != nil {
		return p, result.Error
	}
	return p, nil
}

func (m mysql) GetAll(ctx context.Context, Limit, After int) ([]Job, error) {
	var jobs []Job
	result := m.db.Limit(Limit).Offset(After).Order("start_time").Find(&jobs)
	if result.Error != nil {
		return []Job{}, result.Error
	}
	return jobs, nil
}

func (m mysql) Delete(ctx context.Context, ID string) error {
	result := m.db.Where("id = ?", ID).Delete(Job{})
	if result.Error != nil {
		return result.Error
	}
	return nil
}
