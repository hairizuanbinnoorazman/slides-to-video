package acl

import (
	"context"
	"fmt"

	"cloud.google.com/go/datastore"
	"github.com/hairizuanbinnoorazman/slides-to-video-manager/logger"
)

type googleDatastore struct {
	logger     logger.Logger
	entityName string
	client     *datastore.Client
}

func NewGoogleDatastore(logger logger.Logger, ds *datastore.Client, en string) (*googleDatastore, error) {
	if logger == nil || ds == nil || en == "" {
		return &googleDatastore{}, fmt.Errorf("logger or datastore or entity not defined")
	}
	return &googleDatastore{
		logger:     logger,
		client:     ds,
		entityName: en,
	}, nil
}

func (g *googleDatastore) Create(ctx context.Context, e ACL) error {
	newKey := datastore.NameKey(g.entityName, e.ID, nil)
	_, err := g.client.Put(ctx, newKey, &e)
	if err != nil {
		return fmt.Errorf("unable to send record to datastore: err: %v", err)
	}
	return nil
}

func (g *googleDatastore) Get(ctx context.Context, ProjectID, UserID string) (ACL, error) {
	query := datastore.NewQuery(g.entityName).Filter("ProjectID =", ProjectID).Filter("UserID =", UserID)
	acls := []ACL{}
	_, err := g.client.GetAll(ctx, query, &acls)
	if err != nil {
		return ACL{}, err
	}
	if len(acls) == 0 || len(acls) > 1 {
		return ACL{}, fmt.Errorf("no records found")
	}
	if len(acls) > 1 {
		return ACL{}, fmt.Errorf("bad query to get acl for project id and user id combination. expected only 1 combination")
	}
	return acls[0], nil
}

func (g *googleDatastore) GetAll(ctx context.Context, ProjectID string, Limit, After int) ([]ACL, error) {
	query := datastore.NewQuery(g.entityName).Filter("ProjectID =", ProjectID).Limit(Limit).Offset(After)
	acls := []ACL{}
	_, err := g.client.GetAll(ctx, query, &acls)
	if err != nil {
		return []ACL{}, err
	}
	if len(acls) == 0 || len(acls) > 1 {
		return []ACL{}, fmt.Errorf("bad query to get acl for project id and user id combination")
	}
	return acls, nil
}

func (g *googleDatastore) Update(ctx context.Context, ProjectID, UserID string, setters ...func(*ACL) error) (ACL, error) {
	return ACL{}, nil
}

func (g *googleDatastore) Delete(ctx context.Context, ProjectID, UserID string) error {
	return nil
}
