package project

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/hairizuanbinnoorazman/slides-to-video-manager/logger"
	"github.com/hairizuanbinnoorazman/slides-to-video-manager/pdfslideimages"

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
	time.Sleep(20 * time.Second)
	defer req.Terminate(context.TODO())

	port, err := req.MappedPort(context.TODO(), "3306")

	db := databaseConnProvider(port.Int())
	db.AutoMigrate(&pdfslideimages.PDFSlideImages{})
	db.AutoMigrate(&Project{})
	db.Model(&pdfslideimages.PDFSlideImages{}).AddForeignKey("project_id", "projects(id)", "CASCADE", "CASCADE")
	a := mysql{
		db: db,
	}
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
	pdfDB := pdfslideimages.NewMySQL(logger.LoggerForTests{Tester: t}, db)

	// Creating of record
	err = a.Create(context.TODO(), p)
	if err != nil {
		t.Fatalf("Failed to create record in mysql database. Err: %v", err)
	}
	err = a.Create(context.TODO(), p2)
	if err != nil {
		t.Fatalf("Failed to create record in mysql database. Err: %v", err)
	}

	err = pdfDB.Create(context.TODO(), pdfItem)
	if err != nil {
		t.Fatalf("Failed to create record in mysql database for pdf slide images. Err: %v", err)
	}

	// Single get of record
	retrieveProject, err := a.Get(context.TODO(), "1234", "")
	if err != nil {
		t.Fatalf("Failed to retrieve record from mysql database. Err: %v", err)
	}
	if retrieveProject.ID != "1234" {
		t.Fatalf("Unexpectd project ID in retrieved record. Err: %v", err)
	}
	if len(retrieveProject.PDFSlideImages) == 0 {
		t.Fatalf("Unexpected - pdf slide images are not fetched")
	}

	// Get all records
	projects, err := a.GetAll(context.TODO(), "", 10, 0)
	if err != nil {
		t.Fatalf("Unexpected error when retrieving all records. Err: %v", err)
	}
	if len(projects) != 2 {
		t.Fatalf("Unexpected no of projects. Projects: %+v", projects)
	}

	// Update a single record
	p, err = a.Update(context.TODO(), "1235", "", recreateIdemKeys())
	if err != nil {
		t.Fatalf("Unexpected error when updating record. Err: %v", err)
	}
	if p.SetRunningIdemKey == "" || p.CompleteRecIdemKey == "" {
		t.Errorf("Bad update - idemkeys are not created. Project: %+v", p)
	}

	// Update status
	p, err = a.Update(context.TODO(), "1235", "", setStatus(running))
	if err != nil {
		t.Fatalf("Unexpected error when updating record. Err: %v", err)
	}
	if p.SetRunningIdemKey != "" && p.Status != running {
		t.Errorf("Bad update - status is not created accordingly. Project: %+v", p)
	}

	// Grab updated record and check status
	p, err = a.Get(context.TODO(), "1235", "")
	if err != nil {
		t.Fatalf("Unexpected error when getting record. Err: %v", err)
	}
	if p.Status != running {
		t.Errorf("Bad update - status is not updated accordingly. Project: %+v", p)
	}

	// Delete single record
	err = a.Delete(context.TODO(), "1234", "")
	if err != nil {
		t.Fatalf("Unexpected error when deleting record. Err: %v", err)
	}

	// Get all records
	projects, err = a.GetAll(context.TODO(), "", 10, 0)
	if err != nil {
		t.Fatalf("Unexpected error when retrieving all records. Err: %v", err)
	}
	if len(projects) != 1 {
		t.Fatalf("Unexpected no of projects. Projects: %+v", projects)
	}
}
