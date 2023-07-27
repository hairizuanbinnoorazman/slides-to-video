package project

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/datastore"
	"github.com/hairizuanbinnoorazman/slides-to-video-manager/acl"
	"github.com/hairizuanbinnoorazman/slides-to-video-manager/logger"
	"github.com/hairizuanbinnoorazman/slides-to-video-manager/pdfslideimages"
	"github.com/hairizuanbinnoorazman/slides-to-video-manager/videosegment"
)

type googleDatastore struct {
	logger                   logger.Logger
	entityName               string
	pdfSlideImagesEntityName string
	videoSegmentEntityName   string
	aclEntityName            string
	client                   *datastore.Client
}

func NewGoogleDatastore(logger logger.Logger, ds *datastore.Client, en, pdfslideimagesEn, videoSegmentEn string) *googleDatastore {
	datastore := googleDatastore{
		logger:                   logger,
		client:                   ds,
		entityName:               en,
		pdfSlideImagesEntityName: pdfslideimagesEn,
		videoSegmentEntityName:   videoSegmentEn,
	}
	return &datastore
}

func (g *googleDatastore) Create(ctx context.Context, e Project) error {
	newKey := datastore.NameKey(g.entityName, e.ID, nil)
	_, err := g.client.Put(ctx, newKey, &e)
	if err != nil {
		return fmt.Errorf("unable to send record to datastore: err: %v", err)
	}
	return nil
}

func (g *googleDatastore) Get(ctx context.Context, ID string) (Project, error) {
	key := datastore.NameKey(g.entityName, ID, nil)
	project := Project{}
	if err := g.client.Get(ctx, key, &project); err != nil {
		return Project{}, fmt.Errorf("unable to retrieve value from datastore. err: %v", err)
	}
	project.ID = ID
	pdfSlideImages := []pdfslideimages.PDFSlideImages{}
	query := datastore.NewQuery(g.pdfSlideImagesEntityName)
	query = query.Ancestor(key)
	keys, err := g.client.GetAll(ctx, query, &pdfSlideImages)
	if err != nil {
		return Project{}, err
	}
	for i, key := range keys {
		pdfSlideImages[i].ID = key.Name
		pdfSlideImages[i].ProjectID = ID
	}
	videoSegments := []videosegment.VideoSegment{}
	query = datastore.NewQuery(g.videoSegmentEntityName)
	query = query.Ancestor(key)
	keys, err = g.client.GetAll(ctx, query, &videoSegments)
	if err != nil {
		return Project{}, err
	}
	for i, key := range keys {
		videoSegments[i].ID = key.Name
		videoSegments[i].ProjectID = ID
	}
	project.PDFSlideImages = pdfSlideImages
	project.VideoSegments = videoSegments
	return project, nil
}

func (g *googleDatastore) Update(ctx context.Context, ID string, setters ...func(*Project) error) (Project, error) {
	key := datastore.NameKey(g.entityName, ID, nil)
	project := Project{}
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
		project.DateModified = time.Now()
		_, err := g.client.Put(ctx, key, &project)
		if err != nil {
			return fmt.Errorf("unable to send record to datastore: err: %v", err)
		}
		return nil
	})
	if err != nil {
		return Project{}, fmt.Errorf("unable to send record to datastore: err: %v", err)
	}
	project.ID = ID
	return project, nil
}

func (g *googleDatastore) GetAll(ctx context.Context, userID string, limit, after int) ([]Project, error) {
	acls := []acl.ACL{}
	query := datastore.NewQuery(g.aclEntityName)
	query = query.Limit(limit)
	query = query.Offset(after)
	query = query.Filter("UserID =", userID)
	keys, err := g.client.GetAll(ctx, query, &acls)
	if err != nil {
		return []Project{}, fmt.Errorf("unable to retrieve all results. err: %v", err)
	}

	projects := []Project{}
	err = g.client.GetMulti(ctx, keys, &projects)
	if err != nil {
		return []Project{}, fmt.Errorf("unable to retrieve all results. err: %v", err)
	}
	return projects, nil
}

func (g *googleDatastore) Delete(ctx context.Context, ID string) error {
	key := datastore.NameKey(g.entityName, ID, nil)
	err := g.client.Delete(ctx, key)
	if err != nil {
		return fmt.Errorf("unable to delete project. err: %v", err)
	}
	return nil
}

func (g *googleDatastore) Count(ctx context.Context, UserID string) (int, error) {
	projects := []Project{}
	query := datastore.NewQuery(g.entityName)
	query = query.KeysOnly()
	keys, err := g.client.GetAll(ctx, query, &projects)
	if err != nil {
		return 0, fmt.Errorf("unable to retrieve all results. err: %v", err)
	}
	return len(keys), nil
}
