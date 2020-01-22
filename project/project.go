package project

import (
	"context"
	"time"

	"github.com/gofrs/uuid"
)

var SuccessfulStatus = "success"
var FailureStatus = "failed"
var RunningStatus = "running"

type Project struct {
	ID            string       `json:"id" datastore:"-"`
	PDFFile       string       `json:"pdf_file"`
	SlideAssets   []SlideAsset `json:"slide_assets"`
	VideoOutputID string       `json:"video_output_id"`
	DateCreated   time.Time    `json:"date_created"`
	DateModified  time.Time    `json:"date_modified"`
	Status        string       `json:"status"`
}

func NewProject() Project {
	projectID, _ := uuid.NewV4()
	return Project{
		ID:           projectID.String(),
		DateCreated:  time.Now(),
		DateModified: time.Now(),
	}
}

type SlideAsset struct {
	ImageID string `json:"image_id"`
	Text    string `json:"text"`
	VideoID string `json:"video_id"`
	SlideNo int    `json:"slide_no"`
}

type ProjectStore interface {
	CreateProject(ctx context.Context, e Project) error
	GetProject(ctx context.Context, ID string) (Project, error)
	GetAllProjects(ctx context.Context) ([]Project, error)
	UpdateProject(ctx context.Context, ID string, setters ...func(*Project)) error
	DeleteProject(ctx context.Context, ID string) error
}

func SetPDFFile(file string) func(*Project) {
	return func(a *Project) {
		a.PDFFile = file
	}
}

func SetEmptySlideAsset() func(*Project) {
	return func(a *Project) {
		a.SlideAssets = []SlideAsset{}
	}
}

func SetVideoOutputID(videoOutputID string) func(*Project) {
	return func(a *Project) {
		a.VideoOutputID = videoOutputID
	}
}

func SetImage(imageID string, slideNo int) func(*Project) {
	return func(a *Project) {
		for _, item := range a.SlideAssets {
			if item.SlideNo == slideNo {
				return
			}
		}
		a.SlideAssets = append(a.SlideAssets, SlideAsset{ImageID: imageID, SlideNo: slideNo})
	}
}

func SetSlideText(imageID, slideText string) func(*Project) {
	return func(a *Project) {
		for _, item := range a.SlideAssets {
			if item.ImageID == imageID {
				item.Text = slideText
			}
		}
		return
	}
}

func SetVideoID(imageID, videoID string) func(*Project) {
	return func(a *Project) {
		for _, item := range a.SlideAssets {
			if item.ImageID == imageID {
				item.VideoID = videoID
			}
		}
		return
	}
}

type BySlideNo []SlideAsset

func (s BySlideNo) Len() int { return len(s) }

func (s BySlideNo) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

func (s BySlideNo) Less(i, j int) bool { return s[i].SlideNo < s[j].SlideNo }
