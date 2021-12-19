package acl

import (
	"context"
	"testing"
)

func common_ops(t *testing.T, aclDB Store) {
	acl1 := New("1", "1111")
	acl2 := New("2", "1111")
	acl3 := New("3", "1112")
	acl4 := New("4", "1112")

	err := aclDB.Create(context.TODO(), acl1)
	if err != nil {
		t.Fatalf("unable to store data into store. Err: %v", err)
	}

	aclDB.Create(context.TODO(), acl2)
	aclDB.Create(context.TODO(), acl3)
	aclDB.Create(context.TODO(), acl4)

	tACL, err := aclDB.Get(context.TODO(), "1", "1111")
	if err != nil {
		t.Fatalf("unable to pull data from store. Err: %v", err)
	}
	if tACL.ProjectID != "1" || tACL.UserID != "1111" {
		t.Fatalf("bad acl pulled from store. Expected %+v. Actual %v", acl1, tACL)
	}
}
