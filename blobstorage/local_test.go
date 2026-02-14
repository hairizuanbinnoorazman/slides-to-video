package blobstorage

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/hairizuanbinnoorazman/slides-to-video-manager/logger"
)

func TestNewLocalStorage(t *testing.T) {
	testLogger := logger.LoggerForTests{Tester: t}

	t.Run("Valid path succeeds", func(t *testing.T) {
		tmpDir := t.TempDir()
		storage, err := NewLocalStorage(testLogger, tmpDir)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if storage.BasePath != tmpDir {
			t.Errorf("Expected basePath %s, got %s", tmpDir, storage.BasePath)
		}
	})

	t.Run("Empty path returns error", func(t *testing.T) {
		_, err := NewLocalStorage(testLogger, "")
		if err == nil {
			t.Error("Expected error for empty basePath, got nil")
		}
	})

	t.Run("Read-only directory returns error", func(t *testing.T) {
		tmpDir := t.TempDir()
		readOnlyDir := filepath.Join(tmpDir, "readonly")
		if err := os.Mkdir(readOnlyDir, 0555); err != nil {
			t.Fatalf("Failed to create read-only dir: %v", err)
		}

		_, err := NewLocalStorage(testLogger, readOnlyDir)
		if err == nil {
			t.Error("Expected error for read-only directory, got nil")
		}
	})

	t.Run("Non-existent path is created", func(t *testing.T) {
		tmpDir := t.TempDir()
		newPath := filepath.Join(tmpDir, "new", "nested", "path")

		storage, err := NewLocalStorage(testLogger, newPath)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		// Verify directory was created
		if _, err := os.Stat(storage.BasePath); os.IsNotExist(err) {
			t.Error("Expected directory to be created, but it doesn't exist")
		}
	})
}

func TestLocalStorage_Save(t *testing.T) {
	testLogger := logger.LoggerForTests{Tester: t}
	tmpDir := t.TempDir()
	storage, err := NewLocalStorage(testLogger, tmpDir)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}

	ctx := context.Background()

	t.Run("Save file successfully", func(t *testing.T) {
		fileName := "test.txt"
		content := []byte("test content")

		err := storage.Save(ctx, fileName, content)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		// Verify file exists and has correct content
		fullPath := filepath.Join(tmpDir, fileName)
		savedContent, err := os.ReadFile(fullPath)
		if err != nil {
			t.Errorf("Failed to read saved file: %v", err)
		}
		if string(savedContent) != string(content) {
			t.Errorf("Expected content %s, got %s", content, savedContent)
		}
	})

	t.Run("Save file with nested path", func(t *testing.T) {
		fileName := "pdf/project123/file.pdf"
		content := []byte("nested content")

		err := storage.Save(ctx, fileName, content)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		// Verify file exists
		fullPath := filepath.Join(tmpDir, fileName)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			t.Error("Expected nested file to be created, but it doesn't exist")
		}
	})

	t.Run("Path traversal attempt returns error", func(t *testing.T) {
		fileName := "../../../etc/passwd"
		content := []byte("malicious content")

		err := storage.Save(ctx, fileName, content)
		if err == nil {
			t.Error("Expected error for path traversal attempt, got nil")
		}
	})

	t.Run("Save large file", func(t *testing.T) {
		fileName := "large.bin"
		content := make([]byte, 1024*1024) // 1MB
		for i := range content {
			content[i] = byte(i % 256)
		}

		err := storage.Save(ctx, fileName, content)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		// Verify file size
		fullPath := filepath.Join(tmpDir, fileName)
		info, err := os.Stat(fullPath)
		if err != nil {
			t.Errorf("Failed to stat saved file: %v", err)
		}
		if info.Size() != int64(len(content)) {
			t.Errorf("Expected file size %d, got %d", len(content), info.Size())
		}
	})
}

func TestLocalStorage_Load(t *testing.T) {
	testLogger := logger.LoggerForTests{Tester: t}
	tmpDir := t.TempDir()
	storage, err := NewLocalStorage(testLogger, tmpDir)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}

	ctx := context.Background()

	t.Run("Load existing file successfully", func(t *testing.T) {
		fileName := "existing.txt"
		expectedContent := []byte("existing content")

		// Create the file first
		fullPath := filepath.Join(tmpDir, fileName)
		if err := os.WriteFile(fullPath, expectedContent, 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		// Load the file
		content, err := storage.Load(ctx, fileName)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if string(content) != string(expectedContent) {
			t.Errorf("Expected content %s, got %s", expectedContent, content)
		}
	})

	t.Run("Load non-existent file returns error", func(t *testing.T) {
		fileName := "nonexistent.txt"

		_, err := storage.Load(ctx, fileName)
		if err == nil {
			t.Error("Expected error for non-existent file, got nil")
		}
	})

	t.Run("Path traversal attempt returns error", func(t *testing.T) {
		fileName := "../../../etc/passwd"

		_, err := storage.Load(ctx, fileName)
		if err == nil {
			t.Error("Expected error for path traversal attempt, got nil")
		}
	})

	t.Run("Load nested file", func(t *testing.T) {
		fileName := "images/project456/slide1.png"
		expectedContent := []byte("image data")

		// Create nested directories and file
		fullPath := filepath.Join(tmpDir, fileName)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			t.Fatalf("Failed to create nested directories: %v", err)
		}
		if err := os.WriteFile(fullPath, expectedContent, 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		// Load the file
		content, err := storage.Load(ctx, fileName)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if string(content) != string(expectedContent) {
			t.Errorf("Expected content %s, got %s", expectedContent, content)
		}
	})
}

func TestLocalStorage_Integration(t *testing.T) {
	testLogger := logger.LoggerForTests{Tester: t}
	tmpDir := t.TempDir()
	storage, err := NewLocalStorage(testLogger, tmpDir)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}

	ctx := context.Background()

	t.Run("Save and Load roundtrip", func(t *testing.T) {
		fileName := "roundtrip.bin"
		originalContent := []byte("roundtrip test content with special chars: \n\t\x00\xff")

		// Save the content
		if err := storage.Save(ctx, fileName, originalContent); err != nil {
			t.Fatalf("Failed to save: %v", err)
		}

		// Load it back
		loadedContent, err := storage.Load(ctx, fileName)
		if err != nil {
			t.Fatalf("Failed to load: %v", err)
		}

		// Verify content matches exactly
		if string(loadedContent) != string(originalContent) {
			t.Errorf("Content mismatch. Original: %v, Loaded: %v", originalContent, loadedContent)
		}
	})

	t.Run("Multiple files in different directories", func(t *testing.T) {
		files := map[string][]byte{
			"pdf/project1/file.pdf":       []byte("pdf content 1"),
			"pdf/project2/file.pdf":       []byte("pdf content 2"),
			"images/project1/slide1.png":  []byte("image 1"),
			"images/project1/slide2.png":  []byte("image 2"),
			"videos/project1/snippet1.mp4": []byte("video 1"),
		}

		// Save all files
		for fileName, content := range files {
			if err := storage.Save(ctx, fileName, content); err != nil {
				t.Fatalf("Failed to save %s: %v", fileName, err)
			}
		}

		// Load and verify all files
		for fileName, expectedContent := range files {
			loadedContent, err := storage.Load(ctx, fileName)
			if err != nil {
				t.Fatalf("Failed to load %s: %v", fileName, err)
			}
			if string(loadedContent) != string(expectedContent) {
				t.Errorf("Content mismatch for %s", fileName)
			}
		}
	})
}
