package blobstorage

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hairizuanbinnoorazman/slides-to-video-manager/logger"
)

type LocalStorage struct {
	Logger   logger.Logger
	BasePath string
}

func NewLocalStorage(logger logger.Logger, basePath string) (LocalStorage, error) {
	if basePath == "" {
		return LocalStorage{}, fmt.Errorf("basePath cannot be empty")
	}

	// Clean the path to normalize it
	basePath = filepath.Clean(basePath)

	// Create basePath directory if it doesn't exist
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return LocalStorage{}, fmt.Errorf("failed to create base path directory %s: %w", basePath, err)
	}

	// Verify basePath is writable
	testFile := filepath.Join(basePath, ".write_test")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		return LocalStorage{}, fmt.Errorf("base path %s is not writable: %w", basePath, err)
	}
	os.Remove(testFile)

	return LocalStorage{
		Logger:   logger,
		BasePath: basePath,
	}, nil
}

func (l LocalStorage) Save(ctx context.Context, fileName string, content []byte) error {
	// Clean and validate the file path to prevent path traversal
	cleanFileName := filepath.Clean(fileName)

	// Prevent path traversal by ensuring the file stays within basePath
	fullPath := filepath.Join(l.BasePath, cleanFileName)

	// Verify the resolved path is still within basePath
	relPath, err := filepath.Rel(l.BasePath, fullPath)
	if err != nil || strings.HasPrefix(relPath, "..") {
		return fmt.Errorf("invalid file path: path traversal attempt detected for %s", fileName)
	}

	// Create parent directories if they don't exist
	dirPath := filepath.Dir(fullPath)
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return fmt.Errorf("failed to create parent directories for %s: %w", fileName, err)
	}

	// Write the file
	if err := os.WriteFile(fullPath, content, 0644); err != nil {
		return fmt.Errorf("failed to write file %s: %w", fileName, err)
	}

	l.Logger.Infof("Successfully saved file: %s", fileName)
	return nil
}

func (l LocalStorage) Load(ctx context.Context, fileName string) (content []byte, err error) {
	// Clean and validate the file path to prevent path traversal
	cleanFileName := filepath.Clean(fileName)

	// Prevent path traversal by ensuring the file stays within basePath
	fullPath := filepath.Join(l.BasePath, cleanFileName)

	// Verify the resolved path is still within basePath
	relPath, err := filepath.Rel(l.BasePath, fullPath)
	if err != nil || strings.HasPrefix(relPath, "..") {
		return []byte{}, fmt.Errorf("invalid file path: path traversal attempt detected for %s", fileName)
	}

	// Read the file
	content, err = os.ReadFile(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []byte{}, fmt.Errorf("file not found: %s", fileName)
		}
		return []byte{}, fmt.Errorf("failed to read file %s: %w", fileName, err)
	}

	l.Logger.Infof("Successfully loaded file: %s", fileName)
	return content, nil
}
