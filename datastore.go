package main

import (
	"context"
	"fmt"

	"cloud.google.com/go/datastore"
)

type ParentJob struct {
	ID               string `json:"id"`
	OriginalFilename string `json:"original_filename"`
	Filename         string `json:"filename"`
	Script           string `json:"script"`
	Status           string `json:"status"`
	VideoFile        string `json:"video_file"`
}

type ParentJobDetails struct {
	OriginalFilename string
	Filename         string
	Script           string
	Status           string
	VideoFile        string
}

type PDFToImageJob struct {
	ID          string
	ParentJobID string
	Filename    string
	Status      string
}

type PDFToImageJobDetails struct {
	ParentJobID string
	Filename    string
	Status      string
}

type ImageToVideoJob struct {
	ID          string
	ParentJobID string
	ImageID     string
	SlideID     int
	Text        string
	Status      string
	OutputFile  string
}

type ImageToVideoJobDetails struct {
	ParentJobID string
	ImageID     string
	SlideID     int
	Text        string
	Status      string
	OutputFile  string
}

type VideoConcatJob struct {
	ID          string
	ParentJobID string
	Videos      []string
	Status      string
	OutputFile  string
}

type VideoConcatJobDetails struct {
	ParentJobID string
	Videos      []string
	Status      string
	OutputFile  string
}

type GoogleDatastore struct {
	EntityName string
	Client     *datastore.Client
}

func NewStore(ds *datastore.Client, en string) *GoogleDatastore {
	datastore := GoogleDatastore{Client: ds, EntityName: en}
	return &datastore
}

func (g *GoogleDatastore) StoreParentJob(ctx context.Context, e ParentJob) error {
	newKey := datastore.NameKey(g.EntityName, e.ID, nil)
	_, err := g.Client.Put(ctx, newKey, &ParentJobDetails{
		OriginalFilename: e.OriginalFilename,
		Filename:         e.Filename,
		Script:           e.Script,
		Status:           e.Status,
		VideoFile:        e.VideoFile,
	})
	if err != nil {
		return fmt.Errorf("unable to send record to datastore: err: %v", err)
	}
	return nil
}

func (g *GoogleDatastore) GetParentJob(ctx context.Context, ID string) (ParentJob, error) {
	key := datastore.NameKey(g.EntityName, ID, nil)
	parentJob := ParentJob{}
	if err := g.Client.Get(ctx, key, &parentJob); err != nil {
		return ParentJob{}, fmt.Errorf("unable to retrieve value from datastore. err: %v", err)
	}
	parentJob.ID = ID
	return parentJob, nil
}

func (g *GoogleDatastore) GetAllParentJobs(ctx context.Context) ([]ParentJob, error) {
	emailDetails := []ParentJob{}
	keys, err := g.Client.GetAll(ctx, datastore.NewQuery(g.EntityName), &emailDetails)
	if err != nil {
		return []ParentJob{}, fmt.Errorf("unable to retrieve all results. err: %v", err)
	}
	emailItems := []ParentJob{}
	for i, key := range keys {
		emailItems = append(emailItems, ParentJob{
			ID:               key.Name,
			OriginalFilename: emailDetails[i].OriginalFilename,
			Filename:         emailDetails[i].Filename,
			Script:           emailDetails[i].Script,
			Status:           emailDetails[i].Status,
			VideoFile:        emailDetails[i].VideoFile,
		})
	}
	return emailItems, nil
}

func (g *GoogleDatastore) StorePDFToImageJob(ctx context.Context, e PDFToImageJob) error {
	newKey := datastore.NameKey(g.EntityName, e.ID, nil)
	_, err := g.Client.Put(ctx, newKey, &PDFToImageJobDetails{
		ParentJobID: e.ParentJobID,
		Filename:    e.Filename,
		Status:      e.Status,
	})
	if err != nil {
		return fmt.Errorf("unable to send record to datastore: err: %v", err)
	}
	return nil
}

func (g *GoogleDatastore) GetPDFToImageJob(ctx context.Context, ID string) (PDFToImageJob, error) {
	key := datastore.NameKey(g.EntityName, ID, nil)
	pdfToImageJob := PDFToImageJob{}
	if err := g.Client.Get(ctx, key, &pdfToImageJob); err != nil {
		return PDFToImageJob{}, fmt.Errorf("unable to retrieve value from datastore. err: %v", err)
	}
	pdfToImageJob.ID = ID
	return pdfToImageJob, nil
}

func (g *GoogleDatastore) GetAllPDFToImageJobs(ctx context.Context) ([]PDFToImageJob, error) {
	emailDetails := []PDFToImageJob{}
	keys, err := g.Client.GetAll(ctx, datastore.NewQuery(g.EntityName), &emailDetails)
	if err != nil {
		return []PDFToImageJob{}, fmt.Errorf("unable to retrieve all results. err: %v", err)
	}
	emailItems := []PDFToImageJob{}
	for i, key := range keys {
		emailItems = append(emailItems, PDFToImageJob{
			ID:          key.Name,
			ParentJobID: emailDetails[i].ParentJobID,
			Filename:    emailDetails[i].Filename,
			Status:      emailDetails[i].Status,
		})
	}
	return emailItems, nil
}

func (g *GoogleDatastore) StoreImageToVideoJob(ctx context.Context, e ImageToVideoJob) error {
	newKey := datastore.NameKey(g.EntityName, e.ID, nil)
	_, err := g.Client.Put(ctx, newKey, &ImageToVideoJobDetails{
		ParentJobID: e.ParentJobID,
		ImageID:     e.ImageID,
		SlideID:     e.SlideID,
		Text:        e.Text,
		Status:      e.Status,
		OutputFile:  e.OutputFile,
	})
	if err != nil {
		return fmt.Errorf("unable to send record to datastore: err: %v", err)
	}
	return nil
}

func (g *GoogleDatastore) GetImageToVideoJob(ctx context.Context, ID string) (ImageToVideoJob, error) {
	key := datastore.NameKey(g.EntityName, ID, nil)
	imageToVideoJob := ImageToVideoJob{}
	if err := g.Client.Get(ctx, key, &imageToVideoJob); err != nil {
		return ImageToVideoJob{}, fmt.Errorf("unable to retrieve value from datastore. err: %v", err)
	}
	imageToVideoJob.ID = ID
	return imageToVideoJob, nil
}

func (g *GoogleDatastore) GetAllImageToVideoJobs(ctx context.Context, filterByParentID string) ([]ImageToVideoJob, error) {
	emailDetails := []ImageToVideoJob{}
	q := datastore.NewQuery(g.EntityName)
	if filterByParentID != "" {
		q = q.Filter("ParentJobID = ", filterByParentID)
	}
	keys, err := g.Client.GetAll(ctx, q, &emailDetails)
	if err != nil {
		return []ImageToVideoJob{}, fmt.Errorf("unable to retrieve all results. err: %v", err)
	}
	emailItems := []ImageToVideoJob{}
	for i, key := range keys {
		emailItems = append(emailItems, ImageToVideoJob{
			ID:          key.Name,
			ParentJobID: emailDetails[i].ParentJobID,
			ImageID:     emailDetails[i].ImageID,
			SlideID:     emailDetails[i].SlideID,
			Text:        emailDetails[i].Text,
			Status:      emailDetails[i].Status,
			OutputFile:  emailDetails[i].OutputFile,
		})
	}
	return emailItems, nil
}

func (g *GoogleDatastore) StoreVideoConcatJob(ctx context.Context, e VideoConcatJob) error {
	newKey := datastore.NameKey(g.EntityName, e.ID, nil)
	_, err := g.Client.Put(ctx, newKey, &VideoConcatJobDetails{
		ParentJobID: e.ParentJobID,
		Videos:      e.Videos,
		Status:      e.Status,
		OutputFile:  e.OutputFile,
	})
	if err != nil {
		return fmt.Errorf("unable to send record to datastore: err: %v", err)
	}
	return nil
}

func (g *GoogleDatastore) GetVideoConcatJob(ctx context.Context, ID string) (VideoConcatJob, error) {
	key := datastore.NameKey(g.EntityName, ID, nil)
	videoConcatJob := VideoConcatJob{}
	if err := g.Client.Get(ctx, key, &videoConcatJob); err != nil {
		return VideoConcatJob{}, fmt.Errorf("unable to retrieve value from datastore. err: %v", err)
	}
	videoConcatJob.ID = ID
	return videoConcatJob, nil
}

func (g *GoogleDatastore) GetAllVideoConcatJobs(ctx context.Context) ([]VideoConcatJob, error) {
	emailDetails := []VideoConcatJob{}
	keys, err := g.Client.GetAll(ctx, datastore.NewQuery(g.EntityName), &emailDetails)
	if err != nil {
		return []VideoConcatJob{}, fmt.Errorf("unable to retrieve all results. err: %v", err)
	}
	emailItems := []VideoConcatJob{}
	for i, key := range keys {
		emailItems = append(emailItems, VideoConcatJob{
			ID:          key.Name,
			ParentJobID: emailDetails[i].ParentJobID,
			Videos:      emailDetails[i].Videos,
			OutputFile:  emailDetails[i].OutputFile,
			Status:      emailDetails[i].Status,
		})
	}
	return emailItems, nil
}
