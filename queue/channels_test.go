package queue

import (
	"context"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
)

func TestChannels_AddAndPop(t *testing.T) {
	logger := logrus.New()
	ch := NewChannels(logger, "test-topic", 10)
	defer ch.Close()

	ctx := context.Background()
	message := []byte("test message")

	// Test Add
	err := ch.Add(ctx, message)
	if err != nil {
		t.Errorf("Add failed: %v", err)
	}

	// Test Pop
	receivedMsg, err := ch.Pop(ctx)
	if err != nil {
		t.Errorf("Pop failed: %v", err)
	}

	if string(receivedMsg) != string(message) {
		t.Errorf("Expected message %s, got %s", string(message), string(receivedMsg))
	}
}

func TestChannels_ContextCancellation(t *testing.T) {
	logger := logrus.New()
	ch := NewChannels(logger, "test-topic", 0) // No buffer to force blocking
	defer ch.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// Test Add with cancelled context - should fail immediately since buffer is 0
	err := ch.Add(ctx, []byte("test"))
	if err != context.Canceled {
		t.Errorf("Expected context.Canceled error, got %v", err)
	}

	// Test Pop with cancelled context - should fail immediately since buffer is empty
	_, err = ch.Pop(ctx)
	if err != context.Canceled {
		t.Errorf("Expected context.Canceled error, got %v", err)
	}
}

func TestChannels_Close(t *testing.T) {
	logger := logrus.New()
	ch := NewChannels(logger, "test-topic", 10)

	// Add a message
	ctx := context.Background()
	err := ch.Add(ctx, []byte("test"))
	if err != nil {
		t.Errorf("Add failed: %v", err)
	}

	// Close the channel
	ch.Close()

	// Give a small delay to ensure context cancellation propagates
	time.Sleep(10 * time.Millisecond)

	// Try to add after close - should fail
	err = ch.Add(ctx, []byte("test2"))
	if err == nil {
		t.Error("Expected error when adding to closed channel")
	}

	// Try to pop after close - should fail
	_, err = ch.Pop(ctx)
	if err == nil {
		t.Error("Expected error when popping from closed channel")
	}
}

func TestChannels_BufferFull(t *testing.T) {
	logger := logrus.New()
	bufferSize := 2
	ch := NewChannels(logger, "test-topic", bufferSize)
	defer ch.Close()

	ctx := context.Background()

	// Fill the buffer
	for i := 0; i < bufferSize; i++ {
		err := ch.Add(ctx, []byte("test"))
		if err != nil {
			t.Errorf("Add failed at %d: %v", i, err)
		}
	}

	// Try to add when buffer is full (should block or timeout)
	ctx2, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err := ch.Add(ctx2, []byte("test"))
	if err != context.DeadlineExceeded {
		t.Errorf("Expected context.DeadlineExceeded when buffer full, got %v", err)
	}

	// Pop one message to free space
	_, err = ch.Pop(ctx)
	if err != nil {
		t.Errorf("Pop failed: %v", err)
	}

	// Now should be able to add again
	err = ch.Add(ctx, []byte("test"))
	if err != nil {
		t.Errorf("Add failed after pop: %v", err)
	}
}

func TestChannels_MultipleMessages(t *testing.T) {
	logger := logrus.New()
	ch := NewChannels(logger, "test-topic", 100)
	defer ch.Close()

	ctx := context.Background()
	messageCount := 10

	// Add multiple messages
	for i := 0; i < messageCount; i++ {
		msg := []byte("message-" + string(rune(i)))
		err := ch.Add(ctx, msg)
		if err != nil {
			t.Errorf("Add failed at %d: %v", i, err)
		}
	}

	// Pop all messages
	for i := 0; i < messageCount; i++ {
		_, err := ch.Pop(ctx)
		if err != nil {
			t.Errorf("Pop failed at %d: %v", i, err)
		}
	}
}
