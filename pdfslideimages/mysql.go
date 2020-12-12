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
	return nil
}

func (m mysql) Get(ctx context.Context, projectID, ID string) (PDFSlideImages, error) {
	return PDFSlideImages{}, nil
}

func (m mysql) GetAll(ctx context.Context, ProjectID string, Limit, After int) ([]PDFSlideImages, error) {
	return []PDFSlideImages{}, nil
}

func (m mysql) Update(ctx context.Context, projectID, ID string, setters ...func(*PDFSlideImages) error) (PDFSlideImages, error) {
	return PDFSlideImages{}, nil
}

func (m mysql) Delete(ctx context.Context, projectID, ID string) error {
	return nil
}
