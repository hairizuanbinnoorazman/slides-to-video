package user

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

func (g *GoogleDatastore) StoreUser(ctx context.Context, u User) error {
	newKey := datastore.NameKey(g.EntityName, u.ID, nil)
	_, err := g.Client.Put(ctx, newKey, &u)
	if err != nil {
		return fmt.Errorf("unable to send record to datastore: err: %v", err)
	}
	return nil
}

func (g *GoogleDatastore) GetUser(ctx context.Context, ID string) (User, error) {
	key := datastore.NameKey(g.EntityName, ID, nil)
	user := User{}
	if err := g.Client.Get(ctx, key, &user); err != nil {
		return User{}, fmt.Errorf("unable to retrieve value from datastore. err: %v", err)
	}
	return user, nil
}

func (g *GoogleDatastore) GetUserByEmail(ctx context.Context, Email string) (User, error) {
	var users []User
	q := datastore.NewQuery(g.EntityName).Filter("Email =", Email).Limit(1)
	if _, err := g.Client.GetAll(ctx, q, &users); err != nil {
		return User{}, fmt.Errorf("unable to retrieve list of users. err: %v", err)
	}
	if len(users) == 0 {
		return User{}, nil
	}
	return users[0], nil
}
