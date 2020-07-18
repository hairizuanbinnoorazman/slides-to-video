package project

import (
	"context"
)

type Store interface {
	Create(ctx context.Context, e Project) error
	Get(ctx context.Context, ID string, UserID string) (Project, error)
	GetAll(ctx context.Context, UserID string, Limit, After int) ([]Project, error)
	Update(ctx context.Context, ID string, UserID string, setters ...func(*Project)) (Project, error)
	Delete(ctx context.Context, ID string, UserID string) error
}

// func SetEmptySlideAsset() func(*Project) {
// 	return func(a *Project) {
// 		a.SlideAssets = []SlideAsset{}
// 	}
// }

func SetVideoOutputID(videoOutputID string) func(*Project) {
	return func(a *Project) {
		a.VideoOutputID = videoOutputID
	}
}

func SetStatus(status string) func(*Project) {
	return func(a *Project) {
		if status == "running" {
			a.Status = running
		} else if status == "completed" {
			a.Status = completed
		}
	}
}

func SetIdemKey(idemKey string) func(*Project) {
	return func(a *Project) {
		a.IdemKey = idemKey
	}
}

// func SetImage(imageID string, slideNo int) func(*Project) {
// 	return func(a *Project) {
// 		for _, item := range a.SlideAssets {
// 			if item.SlideNo == slideNo {
// 				return
// 			}
// 		}
// 		a.SlideAssets = append(a.SlideAssets, SlideAsset{ImageID: imageID, SlideNo: slideNo})
// 	}
// }

// func SetSlideText(imageID, slideText string) func(*Project) {
// 	return func(a *Project) {
// 		for idx, item := range a.SlideAssets {
// 			if item.ImageID == imageID {
// 				a.SlideAssets[idx].Text = slideText
// 			}
// 		}
// 		return
// 	}
// }

// func SetVideoID(imageID, videoID string) func(*Project) {
// 	return func(a *Project) {
// 		for idx, item := range a.SlideAssets {
// 			if item.ImageID == imageID {
// 				a.SlideAssets[idx].VideoID = videoID
// 			}
// 		}
// 		return
// 	}
// }

type ByDateCreated []Project

func (s ByDateCreated) Len() int { return len(s) }

func (s ByDateCreated) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

func (s ByDateCreated) Less(i, j int) bool { return s[i].DateCreated.Before(s[j].DateCreated) }
