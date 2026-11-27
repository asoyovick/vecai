package storage

import (
	"context"
	"errors"
	"io"
)

var ErrS3NotConfigured = errors.New("s3 storage is not configured")

// S3Storage is a placeholder for an S3-compatible implementation.
type S3Storage struct{}

// NewS3Storage creates an S3 storage (not yet implemented).
func NewS3Storage() *S3Storage {
	return &S3Storage{}
}

func (s *S3Storage) Save(ctx context.Context, key string, r io.Reader, contentType string) (*Object, error) {
	return nil, ErrS3NotConfigured
}

func (s *S3Storage) Open(ctx context.Context, key string) (io.ReadCloser, error) {
	return nil, ErrS3NotConfigured
}

func (s *S3Storage) Delete(ctx context.Context, key string) error {
	return ErrS3NotConfigured
}
