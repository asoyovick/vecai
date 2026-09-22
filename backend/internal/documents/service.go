package documents

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Service manages file document storage It coordinates persisting raw files
// bytes onto local disk and saving their associated meta data.
type Service struct {
	repo      Repository
	uploadDir string
}

// NewService constructs a new document management service with the provided repository implementation and root upload directory on the disk.
func NewService(repo Repository, uploadDir string) *Service {
	return &Service{repo: repo, uploadDir: uploadDir}
}

// Upload saves a multipart file for a project and records its metadata.
func (s *Service) Upload(ctx context.Context, projectID, userID string, header *multipart.FileHeader, category string) (*Document, error) {
	// Guard clause: ensue the uploaded file contains data.
	if header.Size <= 0 {
		return nil, fmt.Errorf("%w: uploaded file is empty", ErrInvalidInput)
	}
	if header.Size > MaxUploadSize {
		return nil, ErrFileTooLarge
	}
	category = strings.TrimSpace(category)
	//open the uploaded file header stream to read bytes
	src, err := header.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open uploaded file: %w", err)
	}
	defer src.Close()

	docID := uuid.NewString()
	filename := sanitizeFilename(header.Filename)
	storagePath := filepath.Join(s.uploadDir, projectID, docID+"_"+filename)

	if err := os.MkdirAll(filepath.Dir(storagePath), 0o755); err != nil {
		return nil, fmt.Errorf("failed to create upload directory: %w", err)
	}

	dst, err := os.Create(storagePath)
	if err != nil {
		return nil, fmt.Errorf("failed to store uploaded file: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		os.Remove(storagePath)
		return nil, fmt.Errorf("failed to write uploaded file: %w", err)
	}
	// Build the document metadata record.
	doc := &Document{
		ID:          docID,
		ProjectID:   projectID,
		UploadedBy:  userID,
		Filename:    filename,
		FileType:    filepath.Ext(filename),
		FileSize:    header.Size,
		StoragePath: storagePath,
		Category:    category,
		UploadedAt:  time.Now().UTC(),
	}

	if err := s.repo.Create(ctx, doc); err != nil {
		os.Remove(storagePath) // keep disk and database consistent
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

// Delete removes a document from the database and purges its stored file from local disk.
func (s *Service) Delete(ctx context.Context, projectID, documentID string) error {
	doc, err := s.repo.GetByID(ctx, projectID, documentID)
	if err != nil {
		return err
	}
	// delete metadata from the database
	if err := s.repo.Delete(ctx, projectID, documentID); err != nil {
		return err
	}

	if doc.StoragePath != "" {
		if err := os.Remove(doc.StoragePath); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to remove stored file: %w", err)
		}
	}
	return nil
}

// OpenFile returns a reader over the stored bytes for streaming downloads.
func (s *Service) OpenFile(doc *Document) (*os.File, error) {
	return os.Open(doc.StoragePath)
}

// sanitizeFilename strips directory components and characters that could
// escape the upload directory.
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
	//Provides a feedback name if sanitization results in an empty or current-dir string.
	if name == "" || name == "." {
		name = "upload"
	}
	return name
}
