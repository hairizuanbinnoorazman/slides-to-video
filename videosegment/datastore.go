package videosegment

import (
	"context"
	"fmt"

	"cloud.google.com/go/datastore"
)

type googleDatastore struct {
	entityName        string
	projectEntityName string
	client            *datastore.Client
}

func NewGoogleDatastore(ds *datastore.Client, projectEntity, en string) *googleDatastore {
	datastore := googleDatastore{
		client:            ds,
		entityName:        en,
		projectEntityName: projectEntity,
	}
	return &datastore
}

func (g *googleDatastore) Create(ctx context.Context, e VideoSegment) error {
	projectKey := datastore.NameKey(g.projectEntityName, e.ProjectID, nil)
	newKey := datastore.NameKey(g.entityName, e.ID, projectKey)
	_, err := g.client.Put(ctx, newKey, &e)
	if err != nil {
		return fmt.Errorf("unable to send record to datastore: err: %v", err)
	}
	return nil
}

func (g *googleDatastore) Get(ctx context.Context, projectID, ID string) (VideoSegment, error) {
	projectKey := datastore.NameKey(g.projectEntityName, projectID, nil)
	key := datastore.NameKey(g.entityName, ID, projectKey)
	parentJob := VideoSegment{}
	if err := g.client.Get(ctx, key, &parentJob); err != nil {
		return VideoSegment{}, fmt.Errorf("unable to retrieve value from datastore. err: %v", err)
	}
	parentJob.ID = ID
	parentJob.ProjectID = projectID
	return parentJob, nil
}

func (g *googleDatastore) Update(ctx context.Context, projectID, ID string, setters ...func(*VideoSegment)) (VideoSegment, error) {
	projectKey := datastore.NameKey(g.projectEntityName, projectID, nil)
	key := datastore.NameKey(g.entityName, ID, projectKey)
	project := VideoSegment{}
	_, err := g.client.RunInTransaction(context.Background(), func(tx *datastore.Transaction) error {
		if err := g.client.Get(ctx, key, &project); err != nil {
			return fmt.Errorf("unable to retrieve value from datastore. err: %v", err)
		}
		for _, setFunc := range setters {
			setFunc(&project)
		}
		_, err := g.client.Put(ctx, key, &project)
		if err != nil {
			return fmt.Errorf("unable to send record to datastore: err: %v", err)
		}
		return nil
	})
	if err != nil {
		return VideoSegment{}, fmt.Errorf("unable to send record to datastore: err: %v", err)
	}
	project.ID = ID
	project.ProjectID = projectID
	return project, nil
}

func (g *googleDatastore) GetAll(ctx context.Context, projectID string, limit, after int) ([]VideoSegment, error) {
	emailDetails := []VideoSegment{}
	keys, err := g.client.GetAll(ctx, datastore.NewQuery(g.entityName), &emailDetails)
	if err != nil {
		return []VideoSegment{}, fmt.Errorf("unable to retrieve all results. err: %v", err)
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
