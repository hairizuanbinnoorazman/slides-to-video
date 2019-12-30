package jobs

import (
	"context"
	"fmt"

	"cloud.google.com/go/datastore"
)

type GoogleDatastore struct {
	EntityName string
	Client     *datastore.Client
}

func NewGoogleDatastore(ds *datastore.Client, en string) *GoogleDatastore {
	datastore := GoogleDatastore{Client: ds, EntityName: en}
	return &datastore
}

func (g *GoogleDatastore) StoreParentJob(ctx context.Context, e ParentJob) error {
	newKey := datastore.NameKey(g.EntityName, e.ID, nil)
	_, err := g.Client.Put(ctx, newKey, &e)
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
	for i, key := range keys {
		emailDetails[i].ID = key.Name
	}
	return emailDetails, nil
}

func (g *GoogleDatastore) StorePDFToImageJob(ctx context.Context, e PDFToImageJob) error {
	newKey := datastore.NameKey(g.EntityName, e.ID, nil)
	_, err := g.Client.Put(ctx, newKey, &e)
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
		emailDetails[i].ID = key.Name
	}
	return emailItems, nil
}

func (g *GoogleDatastore) StoreImageToVideoJob(ctx context.Context, e ImageToVideoJob) error {
	newKey := datastore.NameKey(g.EntityName, e.ID, nil)
	_, err := g.Client.Put(ctx, newKey, &e)
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
		emailDetails[i].ID = key.Name
	}
	return emailItems, nil
}

func (g *GoogleDatastore) StoreVideoConcatJob(ctx context.Context, e VideoConcatJob) error {
	newKey := datastore.NameKey(g.EntityName, e.ID, nil)
	_, err := g.Client.Put(ctx, newKey, &e)
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
		emailDetails[i].ID = key.Name
	}
	return emailItems, nil
}