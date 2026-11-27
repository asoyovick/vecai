package storage

import (
	"context"
	"io"
)

// Object represents a stored file reference.
type Object struct {
	Key         string
	Size        int64
	ContentType string
}

// Storer abstracts file storage (local disk, S3-compatible, etc.).
type Storer interface {
	// Save stores content and returns the storage key/path and metadata.
	Save(ctx context.Context, key string, r io.Reader, contentType string) (*Object, error)
	// Open returns a reader for the stored object.
	Open(ctx context.Context, key string) (io.ReadCloser, error)
	// Delete removes the stored object.
	Delete(ctx context.Context, key string) error
}
