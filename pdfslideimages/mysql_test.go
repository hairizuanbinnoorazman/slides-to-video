package pdfslideimages

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/testcontainers/testcontainers-go"
)

func databaseConnProvider(port int) *gorm.DB {
	connectionString := fmt.Sprintf("user:password@tcp(localhost:%v)/test-database?parseTime=True", port)
	db, err := gorm.Open("mysql", connectionString)
	if err != nil {
		panic(err)
	}
	return db
}

func Test_mysql_ops(t *testing.T) {
	// Following command is similar to this docker command:
	// docker run --name some-mysql -e MYSQL_ROOT_PASSWORD=root -e MYSQL_DATABASE=test-database -e MYSQL_USER=user -e MYSQL_PASSWORD=password -d -p 3306:3306 mysql:5.7
	req, err := testcontainers.GenericContainer(context.TODO(), testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image: "mysql:5.7",
			Name:  "some-mysql",
			Env: map[string]string{
				"MYSQL_ROOT_PASSWORD": "root",
				"MYSQL_DATABASE":      "test-database",
				"MYSQL_USER":          "user",
				"MYSQL_PASSWORD":      "password",
			},
			ExposedPorts: []string{"3306/tcp"},
		},
		Started: true,
	})
	if err != nil {
		t.Fatalf("Unable to set nats environment. Err: %v", err)
	}
	time.Sleep(20 * time.Second)
	defer req.Terminate(context.TODO())

	port, err := req.MappedPort(context.TODO(), "3306")

	db := databaseConnProvider(port.Int())
	db.AutoMigrate(&PDFSlideImages{})
	db.AutoMigrate(&SlideAsset{})
	db.Model(&SlideAsset{}).AddForeignKey("pdf_slide_image_id", "pdf_slide_images(id)", "CASCADE", "RESTRICT")
	a := mysql{
		db: db,
	}
	p := PDFSlideImages{
		ID:          "1234",
		ProjectID:   "1234",
		DateCreated: time.Now(),
	}
	p2 := PDFSlideImages{
		ID:                "1235",
		ProjectID:         "1234",
		DateCreated:       time.Now(),
		SetRunningIdemKey: "1235",
	}
	s1 := []SlideAsset{
		SlideAsset{
			ImageID:         "1111",
			PDFSlideImageID: "1235",
		},
	}

	// Creating of record
	err = a.Create(context.TODO(), p)
	if err != nil {
		t.Fatalf("Failed to create record in mysql database. Err: %v", err)
	}
	err = a.Create(context.TODO(), p2)
	if err != nil {
		t.Fatalf("Failed to create record in mysql database. Err: %v", err)
	}

	// Single get of record
	retrieveSlidesAsset, err := a.Get(context.TODO(), "1234", "1234")
	if err != nil {
		t.Fatalf("Failed to retrieve record from mysql database. Err: %v", err)
	}
	if retrieveSlidesAsset.ID != "1234" {
		t.Fatalf("Unexpectd project ID in retrieved record. Err: %v", err)
	}

	// Get all records
	pdfimages, err := a.GetAll(context.TODO(), "1234", 10, 0)
	if err != nil {
		t.Fatalf("Unexpected error when retrieving all records. Err: %v", err)
	}
	if len(pdfimages) != 2 {
		t.Fatalf("Unexpected no of projects. Projects: %+v", pdfimages)
	}

	// Update status
	p, err = a.Update(context.TODO(), "1234", "1235", setStatus(running), setSlideAssets(s1))
	if err != nil {
		t.Fatalf("Unexpected error when updating record. Err: %v", err)
	}
	if p.SetRunningIdemKey != "" && p.Status != running {
		t.Errorf("Bad update - status is not created accordingly. Project: %+v", p)
	}

	// Grab updated record and check status
	p, err = a.Get(context.TODO(), "1234", "1235")
	if err != nil {
		t.Fatalf("Unexpected error when getting record. Err: %v", err)
	}
	if p.Status != running {
		t.Errorf("Bad update - status is not updated accordingly. Project: %+v", p)
	}
	if len(p.SlideAssets) != 1 {
		t.Errorf("Bad update - expected to store 1 slide asset by this momemnt for this pdfslideimage storage")
	}

	// Delete single record
	err = a.Delete(context.TODO(), "1234", "1235")
	if err != nil {
		t.Fatalf("Unexpected error when deleting record. Err: %v", err)
	}

	// Get all records
	slideImages, err := a.GetAll(context.TODO(), "1234", 10, 0)
	if err != nil {
		t.Fatalf("Unexpected error when retrieving all records. Err: %v", err)
	}
	if len(slideImages) != 1 {
		t.Fatalf("Unexpected no of slide image assets. Assets: %+v", slideImages)
	}
}
