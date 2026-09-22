package documents

import (
	"errors"
	"time"
)

var (
	ErrDocumentNotFound = errors.New("document not found")
	ErrInvalidInput     = errors.New("invalid document input")
	ErrFileTooLarge     = errors.New("file exceeds the maximum upload size")
)

// Document mirrors the documents table from docs/database.md: only metadata
// lives in PostgreSQL; file bytes are stored on disk under the upload dir.
type Document struct {
	ID          string    `json:"id"`
	ProjectID   string    `json:"project_id"`
	UploadedBy  string    `json:"uploaded_by"`
	Filename    string    `json:"filename"`
	FileType    string    `json:"file_type"`
	FileSize    int64     `json:"file_size"`
	StoragePath string    `json:"-"` // never exposed; downloads stream via the API
	Category    string    `json:"category,omitempty"`
	UploadedAt  time.Time `json:"uploaded_at"`
}

// MaxUploadSize is 20 MB, per docs/coding-standards.md constant example.
const MaxUploadSize = 20 << 20
