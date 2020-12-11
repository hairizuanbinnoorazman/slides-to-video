package queue

import (
	"context"
	"fmt"

	"github.com/hairizuanbinnoorazman/slides-to-video-manager/logger"
	nats "github.com/nats-io/nats.go"
)

type Nats struct {
	Logger       logger.Logger
	Conn         *nats.Conn
	Topic        string
	Subscription *nats.Subscription
}

func NewNats(logger logger.Logger, natsEndpoint string, topic string) (Nats, error) {
	conn, err := nats.Connect(natsEndpoint)
	if err != nil {
		return Nats{}, fmt.Errorf("Error with connecting to Nats. Err: %v", err)
	}
	s, err := conn.SubscribeSync(topic)
	if err != nil {
		return Nats{}, fmt.Errorf("Error with creating the subscriber. Err: %v", err)
	}
	return Nats{
		Logger:       logger,
		Conn:         conn,
		Topic:        topic,
		Subscription: s,
	}, nil
}

func (n Nats) Add(ctx context.Context, message []byte) error {
	err := n.Conn.Publish(n.Topic, message)
	if err != nil {
		return err
	}
	n.Logger.Infof("Message successful transmitted via Nats")
	return nil
}

func (n Nats) Pop(ctx context.Context) ([]byte, error) {
	m, err := n.Subscription.NextMsgWithContext(ctx)
	if err != nil {
		return []byte{}, fmt.Errorf("Unable to retrieve message from nats. Err: %v", err)
	}
	return m.Data, nil
}
