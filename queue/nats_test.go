package queue

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/hairizuanbinnoorazman/slides-to-video-manager/logger"
	"github.com/testcontainers/testcontainers-go"
)

func Test_nats_ops(t *testing.T) {
	// Following command is similar to this docker command:
	// docker run --name some-mysql -e MYSQL_ROOT_PASSWORD=root -e MYSQL_DATABASE=test-database -e MYSQL_USER=user -e MYSQL_PASSWORD=password -d -p 3306:3306 mysql:5.7
	req, err := testcontainers.GenericContainer(context.TODO(), testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "nats:2.1.9",
			Name:         "some-nats",
			ExposedPorts: []string{"4222/tcp"},
		},
		Started: true,
	})
	time.Sleep(2 * time.Second)
	defer req.Terminate(context.TODO())
	if err != nil {
		t.Fatalf("Unable to set nats environment. Err: %v", err)
	}

	port, err := req.MappedPort(context.TODO(), "4222")
	connectionString := fmt.Sprintf("nats://localhost:%v", port)

	queueNats, err := NewNats(logger.LoggerForTests{Tester: t}, connectionString, "testtest")
	if err != nil {
		t.Fatalf("Unable to achieve connection to nats. ConnectionString: %v, Err: %v", connectionString, err)
	}

	testingString := "This is a test"

	err = queueNats.Add(context.TODO(), []byte(testingString))
	if err != nil {
		t.Errorf("Expected no errors from attempting to send message. Err: %v", err)
	}

	resp, err := queueNats.Pop(context.TODO())
	if err != nil {
		t.Errorf("Expected no errors from attempting to receive message. Err: %v", err)
	}
	if string(resp) != testingString {
		t.Errorf("Expected %v but received '%v'", testingString, string(resp))
	}
}
