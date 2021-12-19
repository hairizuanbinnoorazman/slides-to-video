package project

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/hairizuanbinnoorazman/slides-to-video-manager/acl"
	"github.com/hairizuanbinnoorazman/slides-to-video-manager/logger"
	"github.com/hairizuanbinnoorazman/slides-to-video-manager/pdfslideimages"
	"github.com/hairizuanbinnoorazman/slides-to-video-manager/videosegment"

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
	req, _ := testcontainers.GenericContainer(context.TODO(), testcontainers.GenericContainerRequest{
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

	port, _ := req.MappedPort(context.TODO(), "3306")

	db := databaseConnProvider(port.Int())
	db.AutoMigrate(&acl.ACL{})
	db.AutoMigrate(&pdfslideimages.PDFSlideImages{})
	db.AutoMigrate(&pdfslideimages.SlideAsset{})
	db.AutoMigrate(&videosegment.VideoSegment{})
	db.AutoMigrate(&Project{})
	db.Model(&pdfslideimages.PDFSlideImages{}).AddForeignKey("project_id", "projects(id)", "CASCADE", "RESTRICT")
	db.Model(&videosegment.VideoSegment{}).AddForeignKey("project_id", "projects(id)", "CASCADE", "RESTRICT")
	db.Model(&pdfslideimages.SlideAsset{}).AddForeignKey("pdf_slide_image_id", "pdf_slide_images(id)", "CASCADE", "RESTRICT")
	db.Model(&acl.ACL{}).AddForeignKey("project_id", "projects(id)", "CASCADE", "RESTRICT")

	projectStore := NewMySQL(logger.LoggerForTests{Tester: t}, db)
	pdfDB := pdfslideimages.NewMySQL(logger.LoggerForTests{Tester: t}, db)
	aclDB := acl.NewMySQL(logger.LoggerForTests{Tester: t}, db)

	common_ops(t, projectStore, pdfDB, aclDB)
}
