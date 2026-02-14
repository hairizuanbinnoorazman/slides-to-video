package queue

import (
	"context"
	"fmt"

	"github.com/hairizuanbinnoorazman/slides-to-video-manager/logger"
)

type Channels struct {
	Logger  logger.Logger
	Topic   string
	channel chan []byte
	ctx     context.Context
	cancel  context.CancelFunc
}

func NewChannels(logger logger.Logger, topic string, bufferSize int) *Channels {
	ctx, cancel := context.WithCancel(context.Background())
	return &Channels{
		Logger:  logger,
		Topic:   topic,
		channel: make(chan []byte, bufferSize),
		ctx:     ctx,
		cancel:  cancel,
	}
}

func (c *Channels) Add(ctx context.Context, message []byte) error {
	select {
	case <-c.ctx.Done():
		return fmt.Errorf("channel closed for topic: %s", c.Topic)
	case <-ctx.Done():
		return ctx.Err()
	case c.channel <- message:
		c.Logger.Debugf("Message added to channel topic: %s", c.Topic)
		return nil
	}
}

func (c *Channels) Pop(ctx context.Context) ([]byte, error) {
	select {
	case <-c.ctx.Done():
		return nil, fmt.Errorf("channel closed for topic: %s", c.Topic)
	case <-ctx.Done():
		return nil, ctx.Err()
	case msg := <-c.channel:
		return msg, nil
	}
}

func (c *Channels) Close() {
	c.cancel()
	// Don't close the channel to avoid "send on closed channel" panics
	// The context cancellation is sufficient to prevent further operations
}
