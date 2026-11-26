package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// LocalStorage stores files on the local filesystem.
type LocalStorage struct {
	baseDir string
}

// NewLocalStorage creates a local disk storage rooted at baseDir.
func NewLocalStorage(baseDir string) *LocalStorage {
	return &LocalStorage{baseDir: baseDir}
}

func (s *LocalStorage) Save(ctx context.Context, key string, r io.Reader, contentType string) (*Object, error) {
	fullPath := filepath.Join(s.baseDir, key)
	if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
		return nil, fmt.Errorf("create upload dir: %w", err)
	}
	f, err := os.Create(fullPath)
	if err != nil {
		return nil, fmt.Errorf("create file: %w", err)
	}
	defer f.Close()

	n, err := io.Copy(f, r)
	if err != nil {
		os.Remove(fullPath)
		return nil, fmt.Errorf("write file: %w", err)
	}
	return &Object{Key: key, Size: n, ContentType: contentType}, nil
}

func (s *LocalStorage) Open(ctx context.Context, key string) (io.ReadCloser, error) {
	fullPath := filepath.Join(s.baseDir, key)
	return os.Open(fullPath)
}

func (s *LocalStorage) Delete(ctx context.Context, key string) error {
	fullPath := filepath.Join(s.baseDir, key)
	if err := os.Remove(fullPath); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

// BaseDir returns the root directory (useful for debugging/migration).
func (s *LocalStorage) BaseDir() string {
	return s.baseDir
}
