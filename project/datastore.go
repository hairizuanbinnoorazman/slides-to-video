package project

import (
	"context"
	"fmt"
	"time"

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

func (g *GoogleDatastore) CreateProject(ctx context.Context, e Project) error {
	newKey := datastore.NameKey(g.EntityName, e.ID, nil)
	_, err := g.Client.Put(ctx, newKey, &e)
	if err != nil {
		return fmt.Errorf("unable to send record to datastore: err: %v", err)
	}
	return nil
}

func (g *GoogleDatastore) GetProject(ctx context.Context, ID string) (Project, error) {
	key := datastore.NameKey(g.EntityName, ID, nil)
	parentJob := Project{}
	if err := g.Client.Get(ctx, key, &parentJob); err != nil {
		return Project{}, fmt.Errorf("unable to retrieve value from datastore. err: %v", err)
	}
	parentJob.ID = ID
	return parentJob, nil
}

func (g *GoogleDatastore) UpdateProject(ctx context.Context, ID string, setters ...func(*Project)) (Project, error) {
	key := datastore.NameKey(g.EntityName, ID, nil)
	project := Project{}
	_, err := g.Client.RunInTransaction(context.Background(), func(tx *datastore.Transaction) error {
		if err := g.Client.Get(ctx, key, &project); err != nil {
			return fmt.Errorf("unable to retrieve value from datastore. err: %v", err)
		}
		for _, setFunc := range setters {
			setFunc(&project)
		}
		project.DateModified = time.Now()
		_, err := g.Client.Put(ctx, key, &project)
		if err != nil {
			return fmt.Errorf("unable to send record to datastore: err: %v", err)
		}
		return nil
	})
	if err != nil {
		return Project{}, fmt.Errorf("unable to send record to datastore: err: %v", err)
	}
	return project, nil
}

func (g *GoogleDatastore) GetAllProjects(ctx context.Context) ([]Project, error) {
	emailDetails := []Project{}
	keys, err := g.Client.GetAll(ctx, datastore.NewQuery(g.EntityName), &emailDetails)
	if err != nil {
		return []Project{}, fmt.Errorf("unable to retrieve all results. err: %v", err)
	}
	for i, key := range keys {
		emailDetails[i].ID = key.Name
	}
	return emailDetails, nil
}

func (g *GoogleDatastore) DeleteProject(ctx context.Context, ID string) error {
	key := datastore.NameKey(g.EntityName, ID, nil)
	err := g.Client.Delete(ctx, key)
	if err != nil {
		return fmt.Errorf("unable to delete project. err: %v", err)
	}
	return nil
}
