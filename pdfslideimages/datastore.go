package pdfslideimages

import (
	"context"
	"fmt"

	"cloud.google.com/go/datastore"
	"github.com/hairizuanbinnoorazman/slides-to-video-manager/logger"
)

type googleDatastore struct {
	logger            logger.Logger
	projectEntityName string
	entityName        string
	client            *datastore.Client
}

func NewGoogleDatastore(logger logger.Logger, ds *datastore.Client, projectEntity, en string) *googleDatastore {
	datastore := googleDatastore{
		logger:            logger,
		client:            ds,
		entityName:        en,
		projectEntityName: projectEntity,
	}
	return &datastore
}

func (g *googleDatastore) Create(ctx context.Context, e PDFSlideImages) error {
	projectKey := datastore.NameKey(g.projectEntityName, e.ProjectID, nil)
	newKey := datastore.NameKey(g.entityName, e.ID, projectKey)
	_, err := g.client.Put(ctx, newKey, &e)
	if err != nil {
		return fmt.Errorf("unable to send record to datastore: err: %v", err)
	}
	return nil
}

func (g *googleDatastore) Get(ctx context.Context, projectID, ID string) (PDFSlideImages, error) {
	projectKey := datastore.NameKey(g.projectEntityName, projectID, nil)
	key := datastore.NameKey(g.entityName, ID, projectKey)
	parentJob := PDFSlideImages{}
	if err := g.client.Get(ctx, key, &parentJob); err != nil {
		return PDFSlideImages{}, fmt.Errorf("unable to retrieve value from datastore. err: %v", err)
	}
	parentJob.ID = ID
	parentJob.ProjectID = projectID
	return parentJob, nil
}

func (g *googleDatastore) Update(ctx context.Context, projectID, ID string, setters ...func(*PDFSlideImages) error) (PDFSlideImages, error) {
	projectKey := datastore.NameKey(g.projectEntityName, projectID, nil)
	key := datastore.NameKey(g.entityName, ID, projectKey)
	project := PDFSlideImages{}
	_, err := g.client.RunInTransaction(context.Background(), func(tx *datastore.Transaction) error {
		if err := g.client.Get(ctx, key, &project); err != nil {
			return fmt.Errorf("unable to retrieve value from datastore. err: %v", err)
		}
		for _, setFunc := range setters {
			err := setFunc(&project)
			if err != nil {
				return err
			}
		}
		_, err := g.client.Put(ctx, key, &project)
		if err != nil {
			return fmt.Errorf("unable to send record to datastore: err: %v", err)
		}
		return nil
	})
	if err != nil {
		return PDFSlideImages{}, fmt.Errorf("unable to send record to datastore: err: %v", err)
	}
	project.ID = ID
	project.ProjectID = projectID
	return project, nil
}

func (g *googleDatastore) GetAll(ctx context.Context, projectID string, limit, after int) ([]PDFSlideImages, error) {
	emailDetails := []PDFSlideImages{}
	keys, err := g.client.GetAll(ctx, datastore.NewQuery(g.entityName), &emailDetails)
	if err != nil {
		return []PDFSlideImages{}, fmt.Errorf("unable to retrieve all results. err: %v", err)
	}
	for i, key := range keys {
		emailDetails[i].ID = key.Name
		emailDetails[i].ProjectID = projectID
	}
	return emailDetails, nil
}

func (g *googleDatastore) Delete(ctx context.Context, projectID, ID string) error {
	projectKey := datastore.NameKey(g.projectEntityName, projectID, nil)
	key := datastore.NameKey(g.entityName, ID, projectKey)
	err := g.client.Delete(ctx, key)
	if err != nil {
		return fmt.Errorf("unable to delete project. err: %v", err)
	}
	return nil
}
