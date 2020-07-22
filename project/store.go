package project

import (
	"context"
)

type Store interface {
	Create(ctx context.Context, e Project) error
	Get(ctx context.Context, ID string, UserID string) (Project, error)
	GetAll(ctx context.Context, UserID string, Limit, After int) ([]Project, error)
	Update(ctx context.Context, ID string, UserID string, setters ...func(*Project) error) (Project, error)
	Delete(ctx context.Context, ID string, UserID string) error
}

func SetVideoOutputID(videoOutputID string) func(*Project) error {
	return func(a *Project) error {
		a.VideoOutputID = videoOutputID
		return nil
	}
}

func SetStatus(status string) func(*Project) error {
	return func(a *Project) error {
		if status == "running" {
			a.Status = running
		} else if status == "completed" {
			a.Status = completed
		}
		return nil
	}
}

type ByDateCreated []Project

func (s ByDateCreated) Len() int { return len(s) }

func (s ByDateCreated) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

func (s ByDateCreated) Less(i, j int) bool { return s[i].DateCreated.Before(s[j].DateCreated) }
