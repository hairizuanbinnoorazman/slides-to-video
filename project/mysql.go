package project

import (
	"context"

	"github.com/jinzhu/gorm"
)

type mysql struct {
	db *gorm.DB
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
	return []Project{}, nil
}

func (m mysql) Update(ctx context.Context, ID string, UserID string, setters ...func(*Project) error) (Project, error) {
	return Project{}, nil
}

func (m mysql) Delete(ctx context.Context, ID string, UserID string) error {
	return nil
}