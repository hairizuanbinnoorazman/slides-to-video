package user

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/datastore"
)

type GoogleDatastore struct {
	entityName string
	client     *datastore.Client
}

func NewGoogleDatastore(ds *datastore.Client, en string) *GoogleDatastore {
	datastore := GoogleDatastore{client: ds, entityName: en}
	return &datastore
}

func (g *GoogleDatastore) Create(ctx context.Context, u User) error {
	newKey := datastore.NameKey(g.entityName, u.ID, nil)
	_, err := g.client.Put(ctx, newKey, &u)
	if err != nil {
		return fmt.Errorf("unable to send record to datastore: err: %v", err)
	}
	return nil
}

func (g *GoogleDatastore) GetUser(ctx context.Context, ID string) (User, error) {
	key := datastore.NameKey(g.entityName, ID, nil)
	user := User{}
	if err := g.client.Get(ctx, key, &user); err != nil {
		return User{}, fmt.Errorf("unable to retrieve value from datastore. err: %v", err)
	}
	return user, nil
}

func (g *GoogleDatastore) GetUserByEmail(ctx context.Context, Email string) (User, error) {
	var users []User
	q := datastore.NewQuery(g.entityName).Filter("Email =", Email).Limit(1)
	if _, err := g.client.GetAll(ctx, q, &users); err != nil {
		return User{}, fmt.Errorf("unable to retrieve list of users. err: %v", err)
	}
	if len(users) == 0 {
		return User{}, nil
	}
	return users[0], nil
}

func (g *GoogleDatastore) Update(ctx context.Context, userID string, setters ...func(*User) error) (User, error) {
	key := datastore.NameKey(g.entityName, userID, nil)
	user := User{}
	_, err := g.client.RunInTransaction(context.Background(), func(tx *datastore.Transaction) error {
		if err := g.client.Get(ctx, key, &user); err != nil {
			return fmt.Errorf("unable to retrieve value from datastore. err: %v", err)
		}
		for _, setFunc := range setters {
			err := setFunc(&user)
			if err != nil {
				return err
			}
		}
		user.DateModified = time.Now()
		_, err := g.client.Put(ctx, key, &user)
		if err != nil {
			return fmt.Errorf("unable to send record to datastore: err: %v", err)
		}
		return nil
	})
	if err != nil {
		return User{}, fmt.Errorf("unable to send record to datastore: err: %v", err)
	}
	user.ID = userID
	return user, nil
}
