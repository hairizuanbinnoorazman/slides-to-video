package project

import (
	"context"

	"github.com/hairizuanbinnoorazman/slides-to-video-manager/videosegment"

	"github.com/hairizuanbinnoorazman/slides-to-video-manager/pdfslideimages"

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
	// a := acl.ACL{}
	// result := m.db.Where("user_id = ? and project_id = ?", UserID, ID).First(&a)
	// if result.Error != nil {
	// 	return p, result.Error
	// }
	result := m.db.Where("id = ?", ID).First(&p)
	if result.Error != nil {
		return p, result.Error
	}
	var slideImages []pdfslideimages.PDFSlideImages
	result = m.db.Where("project_id = ?", ID).Find(&slideImages)
	if result.Error != nil {
		return p, result.Error
	}
	for k, s := range slideImages {
		var asset []pdfslideimages.SlideAsset
		result = m.db.Where("pdf_slide_image_id = ?", s.ID).Find(&asset)
		if result.Error != nil {
			return p, result.Error
		}
		slideImages[k].SlideAssets = asset
	}
	p.PDFSlideImages = slideImages
	var segments []videosegment.VideoSegment
	result = m.db.Where("project_id = ?", ID).Find(&segments)
	if result.Error != nil {
		return p, result.Error
	}
	p.VideoSegments = segments
	return p, nil
}

func (m mysql) GetAll(ctx context.Context, UserID string, Limit, After int) ([]Project, error) {
	var projects []Project
	result := m.db.Order("date_created desc").Limit(Limit).Offset(After).Find(&projects)
	// result := m.db.Model(&acl.ACL{}).Where("user_id = ?", UserID).Joins("left join projects on acl.project_id = projects.id").Find(&projects)
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
	result = m.db.Save(&p)
	if result.Error != nil {
		return Project{}, result.Error
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
