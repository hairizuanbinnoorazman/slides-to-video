package jobs

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

func (g *GoogleDatastore) CreateJob(ctx context.Context, e Job) error {
	newKey := datastore.NameKey(g.EntityName, e.ID, nil)
	_, err := g.Client.Put(ctx, newKey, &e)
	if err != nil {
		return fmt.Errorf("unable to send record to datastore: err: %v", err)
	}
	return nil
}

func (g *GoogleDatastore) GetJob(ctx context.Context, ID string) (Job, error) {
	key := datastore.NameKey(g.EntityName, ID, nil)
	job := Job{}
	if err := g.Client.Get(ctx, key, &job); err != nil {
		return Job{}, fmt.Errorf("unable to retrieve value from datastore. err: %v", err)
	}
	job.ID = ID
	return job, nil
}

func (g *GoogleDatastore) UpdateJob(ctx context.Context, ID string, setters ...func(*Job)) (Job, error) {
	key := datastore.NameKey(g.EntityName, ID, nil)
	job := Job{}
	_, err := g.Client.RunInTransaction(context.Background(), func(tx *datastore.Transaction) error {
		if err := g.Client.Get(ctx, key, &job); err != nil {
			return fmt.Errorf("unable to retrieve value from datastore. err: %v", err)
		}
		for _, setFunc := range setters {
			setFunc(&job)
		}
		job.DateModified = time.Now()
		_, err := g.Client.Put(ctx, key, &job)
		if err != nil {
			return fmt.Errorf("unable to send record to datastore: err: %v", err)
		}
		return nil
	})
	if err != nil {
		return Job{}, fmt.Errorf("unable to send complete update job transaction properly. err: %v", err)
	}
	return job, nil
}

func (g *GoogleDatastore) GetAllJobs(ctx context.Context, filters ...filter) ([]Job, error) {
	jobs := []Job{}
	query := datastore.NewQuery(g.EntityName)
	for _, singleFilter := range filters {
		query.Filter(singleFilter.Key+" "+singleFilter.Operator, singleFilter.Value)
	}
	keys, err := g.Client.GetAll(ctx, query, &jobs)
	if err != nil {
		return []Job{}, fmt.Errorf("unable to retrieve all results. err: %v", err)
	}
	for i, key := range keys {
		jobs[i].ID = key.Name
	}
	return jobs, nil
}

func (g *GoogleDatastore) DeleteJob(ctx context.Context, ID string) error {
	key := datastore.NameKey(g.EntityName, ID, nil)
	err := g.Client.Delete(ctx, key)
	if err != nil {
		return fmt.Errorf("unable to retrieve all results. err: %v", err)
	}
	return nil
}

func (g *GoogleDatastore) DeleteJobs(ctx context.Context, filters ...filter) error {
	jobs := []Job{}
	query := datastore.NewQuery(g.EntityName)
	for _, singleFilter := range filters {
		query.Filter(singleFilter.Key+" "+singleFilter.Operator, singleFilter.Value)
	}
	_, err := g.Client.RunInTransaction(context.Background(), func(tx *datastore.Transaction) error {
		keys, err := g.Client.GetAll(ctx, query, &jobs)
		if err != nil {
			return fmt.Errorf("unable to retrive keys for deletion. err: %v", err)
		}
		err = g.Client.DeleteMulti(ctx, keys)
		if err != nil {
			return fmt.Errorf("unable to delete data. err: %v", err)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("unable to complete delete jobs transaction. err: %v", err)
	}
	return nil
}
