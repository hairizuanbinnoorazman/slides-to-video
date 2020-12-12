package pdfslideimages

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

func (m mysql) Create(ctx context.Context, e PDFSlideImages) error {
	result := m.db.Create(&e)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (m mysql) Get(ctx context.Context, projectID, ID string) (PDFSlideImages, error) {
	p := PDFSlideImages{}
	result := m.db.Where("id = ? AND project_id = ?", ID, projectID).First(&p)
	if result.Error != nil {
		return p, result.Error
	}
	return p, nil
}

func (m mysql) GetAll(ctx context.Context, ProjectID string, Limit, After int) ([]PDFSlideImages, error) {
	var pdfSlideImages []PDFSlideImages
	result := m.db.Where("project_id = ?", ProjectID).Limit(Limit).Offset(After).Find(&pdfSlideImages)
	if result.Error != nil {
		return []PDFSlideImages{}, result.Error
	}
	return pdfSlideImages, nil
}

func (m mysql) Update(ctx context.Context, projectID, ID string, setters ...func(*PDFSlideImages) error) (PDFSlideImages, error) {
	var p PDFSlideImages
	result := m.db.Where("project_id = ? AND id = ?", projectID, ID).First(&p)
	if result.Error != nil {
		return PDFSlideImages{}, result.Error
	}
	for _, s := range setters {
		err := s(&p)
		if err != nil {
			return PDFSlideImages{}, err
		}
	}
	result = m.db.Save(&p)
	if result.Error != nil {
		return PDFSlideImages{}, result.Error
	}
	return p, nil
}

func (m mysql) Delete(ctx context.Context, projectID, ID string) error {
	result := m.db.Where("id = ? and project_id = ?", ID, projectID).Delete(PDFSlideImages{})
	if result.Error != nil {
		return result.Error
	}
	return nil
}
