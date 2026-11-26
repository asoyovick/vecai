package documents

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"

	"vecai/backend/internal/storage"
)

// Service manages document metadata and file storage via a Storer.
type Service struct {
	repo      Repository
	storer    storage.Storer
	uploadDir string // kept for backward compatibility in logs; storage may ignore
}

// NewService constructs a document service.
func NewService(repo Repository, storer storage.Storer) *Service {
	return &Service{repo: repo, storer: storer}
}

var allowedContentTypes = map[string]bool{
	"application/pdf":    true,
	"image/png":          true,
	"image/jpeg":         true,
	"image/jpg":          true,
	"application/msword": true,
	"application/vnd.openxmlformats-officedocument.wordprocessingml.document": true,
	"application/vnd.ms-excel": true,
	"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet": true,
}

const sniffSize = 512

// Upload saves a multipart file for a project and records its metadata.
func (s *Service) Upload(ctx context.Context, projectID, userID string, header *multipart.FileHeader, category string) (*Document, error) {
	if header.Size <= 0 {
		return nil, fmt.Errorf("%w: uploaded file is empty", ErrInvalidInput)
	}
	if header.Size > MaxUploadSize {
		return nil, ErrFileTooLarge
	}
	category = strings.TrimSpace(category)

	src, err := header.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open uploaded file: %w", err)
	}
	defer src.Close()

	// Sniff content type from first bytes
	buf := make([]byte, sniffSize)
	n, err := src.Read(buf)
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("failed to read uploaded file: %w", err)
	}
	sniffed := http.DetectContentType(buf[:n])
	if !allowedContentTypes[sniffed] {
		return nil, fmt.Errorf("%w: file type %s is not allowed", ErrInvalidInput, sniffed)
	}

	// Reconstruct reader: first the sniffed bytes we read, then the rest
	combined := io.MultiReader(strings.NewReader(string(buf[:n])), src)

	docID := uuid.NewString()
	filename := sanitizeFilename(header.Filename)
	key := projectID + "/" + docID + "_" + filename

	obj, err := s.storer.Save(ctx, key, combined, sniffed)
	if err != nil {
		return nil, fmt.Errorf("failed to store uploaded file: %w", err)
	}

	doc := &Document{
		ID:          docID,
		ProjectID:   projectID,
		UploadedBy:  userID,
		Filename:    filename,
		FileType:    sniffed,
		FileSize:    obj.Size,
		StoragePath: obj.Key,
		Category:    category,
		UploadedAt:  time.Now().UTC(),
	}

	if err := s.repo.Create(ctx, doc); err != nil {
		_ = s.storer.Delete(ctx, key)
		return nil, fmt.Errorf("failed to record document: %w", err)
	}
	return doc, nil
}

// Get retrieves document by project and document IDs.
func (s *Service) Get(ctx context.Context, projectID, documentID string) (*Document, error) {
	return s.repo.GetByID(ctx, projectID, documentID)
}

// List fetches all documents metadata records belonging to a specific project.
func (s *Service) List(ctx context.Context, projectID string) ([]*Document, error) {
	return s.repo.ListByProject(ctx, projectID)
}

// Delete removes a document from the database and purges its stored file.
func (s *Service) Delete(ctx context.Context, projectID, documentID string) error {
	doc, err := s.repo.GetByID(ctx, projectID, documentID)
	if err != nil {
		return err
	}
	if err := s.repo.Delete(ctx, projectID, documentID); err != nil {
		return err
	}
	if doc.StoragePath != "" {
		if err := s.storer.Delete(ctx, doc.StoragePath); err != nil {
			// log but don't fail the delete? best effort
			_ = err
		}
	}
	return nil
}

// OpenFile returns a reader over the stored bytes for streaming downloads.
func (s *Service) OpenFile(ctx context.Context, doc *Document) (io.ReadCloser, error) {
	return s.storer.Open(ctx, doc.StoragePath)
}

// sanitizeFilename strips directory components and characters that could escape the upload directory.
func sanitizeFilename(name string) string {
	name = filepath.Base(name)
	name = strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9':
			return r
		case r == '.', r == '-', r == '_':
			return r
		default:
			return '_'
		}
	}, name)
	if name == "" || name == "." {
		name = "upload"
	}
	return name
}
