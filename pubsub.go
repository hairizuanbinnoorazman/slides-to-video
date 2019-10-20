package main

import (
	"context"
	"fmt"

	"cloud.google.com/go/pubsub"
)

type Pubsub struct {
	logger Logger
	client *pubsub.Client
	topic  string
}

func (p *Pubsub) publish(ctx context.Context, message []byte) error {
	result := p.client.Topic(p.topic).Publish(ctx, &pubsub.Message{
		Data: message,
	})
	id, err := result.Get(ctx)
	if err != nil {
		return fmt.Errorf("Get: %v", err)
	}
	p.logger.Infof("Published a message; msg ID: %v\n", id)
	return nil
}
