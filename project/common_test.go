package project

import (
	"context"
	"testing"
	"time"

	"github.com/hairizuanbinnoorazman/slides-to-video-manager/acl"
	"github.com/hairizuanbinnoorazman/slides-to-video-manager/pdfslideimages"
)

func common_ops(t *testing.T, projectStore Store, pdfDB pdfslideimages.Store, aclDB acl.Store) {
	p := Project{
		ID:           "1234",
		DateCreated:  time.Now(),
		DateModified: time.Now(),
	}
	p2 := Project{
		ID:           "1235",
		DateCreated:  time.Now(),
		DateModified: time.Now(),
	}

	pdfItem := pdfslideimages.PDFSlideImages{
		ID:          "1234",
		ProjectID:   "1234",
		DateCreated: time.Now(),
	}

	// Creating of record
	err := projectStore.Create(context.TODO(), p)
	if err != nil {
		t.Fatalf("Failed to create record in mysql database. Err: %v", err)
	}
	err = projectStore.Create(context.TODO(), p2)
	if err != nil {
		t.Fatalf("Failed to create record in mysql database. Err: %v", err)
	}

	acl1 := acl.New("1234", "1111")
	acl2 := acl.New("1235", "1111")

	err = aclDB.Create(context.TODO(), acl1)
	if err != nil {
		t.Fatalf("Failed to create record in mysql database. Err: %v", err)
	}
	err = aclDB.Create(context.TODO(), acl2)
	if err != nil {
		t.Fatalf("Failed to create record in mysql database. Err: %v", err)
	}

	err = pdfDB.Create(context.TODO(), pdfItem)
	if err != nil {
		t.Fatalf("Failed to create record in mysql database for pdf slide images. Err: %v", err)
	}

	// Single get of record
	retrieveProject, err := projectStore.Get(context.TODO(), "1234")
	if err != nil {
		t.Fatalf("Failed to retrieve record from mysql database. Err: %v", err)
	}
	if retrieveProject.ID != "1234" {
		t.Fatalf("Unexpectd project ID in retrieved record. Err: %v", err)
	}
	if len(retrieveProject.PDFSlideImages) == 0 {
		t.Fatalf("Unexpected - pdf slide images are not fetched")
	}

	// Get number of records
	projectCount, err := projectStore.Count(context.TODO(), "1111")
	if err != nil {
		t.Fatalf("Failed to retrieve record from mysql database. Err: %v", err)
	}
	if projectCount != 2 {
		t.Fatalf("Unexpectd project count. Expected %v Actual %v", 2, projectCount)
	}

	// Invalid project id
	_, err = projectStore.Get(context.TODO(), "9999")
	if err == nil {
		t.Fatalf("Expected that record will not be found - invalid userid passed in")
	}

	// Get all records
	projects, err := projectStore.GetAll(context.TODO(), "1111", 10, 0)
	if err != nil {
		t.Fatalf("Unexpected error when retrieving all records. Err: %v", err)
	}
	if len(projects) != 2 {
		t.Fatalf("Unexpected no of projects. Projects: %+v", projects)
	}

	// Update a single record
	p, err = projectStore.Update(context.TODO(), "1235", recreateIdemKeys())
	if err != nil {
		t.Fatalf("Unexpected error when updating record. Err: %v", err)
	}
	if p.SetRunningIdemKey == "" || p.CompleteRecIdemKey == "" {
		t.Errorf("Bad update - idemkeys are not created. Project: %+v", p)
	}

	// Update status
	p, err = projectStore.Update(context.TODO(), "1235", setStatus(running))
	if err != nil {
		t.Fatalf("Unexpected error when updating record. Err: %v", err)
	}
	if p.SetRunningIdemKey != "" && p.Status != running {
		t.Errorf("Bad update - status is not created accordingly. Project: %+v", p)
	}

	// Grab updated record and check status
	p, err = projectStore.Get(context.TODO(), "1235")
	if err != nil {
		t.Fatalf("Unexpected error when getting record. Err: %v", err)
	}
	if p.Status != running {
		t.Errorf("Bad update - status is not updated accordingly. Project: %+v", p)
	}

	// Delete single record
	err = projectStore.Delete(context.TODO(), "1234")
	if err != nil {
		t.Fatalf("Unexpected error when deleting record. Err: %v", err)
	}

	// Get all records
	projects, err = projectStore.GetAll(context.TODO(), "1111", 10, 0)
	if err != nil {
		t.Fatalf("Unexpected error when retrieving all records. Err: %v", err)
	}
	if len(projects) != 1 {
		t.Fatalf("Unexpected no of projects. Projects: %+v", projects)
	}

}
