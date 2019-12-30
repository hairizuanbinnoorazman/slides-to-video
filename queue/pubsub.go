package queue

import (
	"context"
	"fmt"

	"cloud.google.com/go/pubsub"
	"github.com/hairizuanbinnoorazman/slides-to-video-manager/logger"
)

type Pubsub struct {
	Logger logger.Logger
	Client *pubsub.Client
	Topic  string
}

func NewGooglePubsub(logger logger.Logger, client *pubsub.Client, topic string) Pubsub {
	return Pubsub{
		Logger: logger,
		Client: client,
		Topic:  topic,
	}
}

func (p Pubsub) Add(ctx context.Context, message []byte) error {
	result := p.Client.Topic(p.Topic).Publish(ctx, &pubsub.Message{
		Data: message,
	})
	id, err := result.Get(ctx)
	if err != nil {
		return fmt.Errorf("Get: %v", err)
	}
	p.Logger.Infof("Published a message; msg ID: %v\n", id)
	return nil
}
