package workers

import (
	"context"
	"testing"
)

func TestQueueWorker_Start_ValidatesRequiredFields(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name          string
		worker        *QueueWorker
		expectedError string
	}{
		{
			name: "nil Queue",
			worker: &QueueWorker{
				Queue:         nil,
				ProcessorFunc: func(context.Context, []byte) error { return nil },
				Logger:        &mockLogger{},
				WorkerName:    "test-worker",
			},
			expectedError: "QueueWorker.Queue is nil",
		},
		{
			name: "nil ProcessorFunc",
			worker: &QueueWorker{
				Queue:         &mockQueue{},
				ProcessorFunc: nil,
				Logger:        &mockLogger{},
				WorkerName:    "test-worker",
			},
			expectedError: "QueueWorker.ProcessorFunc is nil",
		},
		{
			name: "nil Logger",
			worker: &QueueWorker{
				Queue:         &mockQueue{},
				ProcessorFunc: func(context.Context, []byte) error { return nil },
				Logger:        nil,
				WorkerName:    "test-worker",
			},
			expectedError: "QueueWorker.Logger is nil",
		},
		{
			name: "empty WorkerName",
			worker: &QueueWorker{
				Queue:         &mockQueue{},
				ProcessorFunc: func(context.Context, []byte) error { return nil },
				Logger:        &mockLogger{},
				WorkerName:    "",
			},
			expectedError: "QueueWorker.WorkerName is empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.worker.Start(ctx)
			if err == nil {
				t.Errorf("Expected error but got nil")
			} else if err.Error() != tt.expectedError && !contains(err.Error(), tt.expectedError) {
				t.Errorf("Expected error containing '%s', got '%s'", tt.expectedError, err.Error())
			}
		})
	}
}

func TestQueueWorker_Start_ValidWorkerStartsSuccessfully(t *testing.T) {
	// Create a context that we'll cancel to stop the worker
	ctx, cancel := context.WithCancel(context.Background())

	worker := &QueueWorker{
		Queue:         &mockQueue{},
		ProcessorFunc: func(context.Context, []byte) error { return nil },
		Logger:        &mockLogger{},
		WorkerName:    "test-worker",
	}

	// Start the worker in a goroutine
	errChan := make(chan error, 1)
	go func() {
		errChan <- worker.Start(ctx)
	}()

	// Cancel the context to stop the worker
	cancel()

	// Wait for the worker to stop
	err := <-errChan

	// Worker should return context.Canceled error
	if err != context.Canceled {
		t.Errorf("Expected context.Canceled error, got: %v", err)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr))
}

// Mock implementations for testing
type mockQueue struct{}

func (m *mockQueue) Add(ctx context.Context, message []byte) error {
	return nil
}

func (m *mockQueue) Pop(ctx context.Context) ([]byte, error) {
	// Block forever to avoid actual processing
	<-ctx.Done()
	return nil, ctx.Err()
}

type mockLogger struct{}

func (m *mockLogger) Debug(args ...interface{})                   {}
func (m *mockLogger) Debugf(format string, args ...interface{})   {}
func (m *mockLogger) Info(args ...interface{})                    {}
func (m *mockLogger) Infof(format string, args ...interface{})    {}
func (m *mockLogger) Warning(args ...interface{})                 {}
func (m *mockLogger) Warningf(format string, args ...interface{}) {}
func (m *mockLogger) Error(args ...interface{})                   {}
func (m *mockLogger) Errorf(format string, args ...interface{})   {}
