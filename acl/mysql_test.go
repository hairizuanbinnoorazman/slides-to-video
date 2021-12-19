package acl

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/hairizuanbinnoorazman/slides-to-video-manager/logger"
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
	db.AutoMigrate(&ACL{})

	aclDB := NewMySQL(logger.LoggerForTests{Tester: t}, db)

	common_ops(t, aclDB)
}
