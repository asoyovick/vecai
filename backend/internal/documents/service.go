package documents

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"vecai/internal/api"
	"vecai/internal/auth"
)

// Service contains the document upload/list/download rules.
type Service struct {
	repo       Repository
	uploadsDir string
}

func NewService(repo Repository, uploadsDir string) *Service {
	return &Service{repo: repo, uploadsDir: uploadsDir}
}

func (s *Service) ListDocuments(projectID, category, q any) ([]map[string]any, error) {
	rows, err := s.repo.ListDocuments(projectID, category, q)
	if rows == nil {
		rows = []map[string]any{}
	}
	return rows, err
}

// Upload stores a multipart file and records its metadata.
func (s *Service) Upload(u *auth.User, file multipart.File, header *multipart.FileHeader, projectIDRaw, name, category, version string) (map[string]any, error) {
	if u.Role == "viewer" {
		return nil, errors.New("You do not have permission to upload documents.")
	}
	defer file.Close()
	projectID, err := api.AsNumber(projectIDRaw, "Project", true, 0, true)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(name) == "" {
		return nil, errors.New("Document name is required.")
	}
	if version == "" {
		version = "1.0"
	}
	validCategory, err := api.AsOneOf(category, "Category", Categories)
	if err != nil {
		return nil, err
	}
	if header.Size == 0 {
		return nil, errors.New("The selected file is empty.")
	}
	if header.Size > 25<<20 {
		return nil, errors.New("Files are limited to 25 MB.")
	}
	_, ext := sanitizeFilename(header.Filename)
	if ext != "" && !allowedExtensions[ext] {
		return nil, fmt.Errorf("File type “.%s” is not allowed.", ext)
	}
	if ok, _ := s.repo.ProjectExists(int64(projectID)); !ok {
		return nil, errors.New("Select a valid project.")
	}
	if err := os.MkdirAll(s.uploadsDir, 0o755); err != nil {
		return nil, err
	}
	randBytes := make([]byte, 6)
	rand.Read(randBytes)
	storedName := fmt.Sprintf("%d-%s", time.Now().UnixMilli(), hex.EncodeToString(randBytes))
	if ext != "" {
		storedName += "." + ext
	}
	dst, err := os.Create(filepath.Join(s.uploadsDir, storedName))
	if err != nil {
		return nil, err
	}
	if _, err := io.Copy(dst, file); err != nil {
		dst.Close()
		return nil, err
	}
	dst.Close()
	mimeType := header.Header.Get("Content-Type")
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}
	id, err := s.repo.CreateDocument(int64(projectID), u.ID, name, validCategory, version, header.Filename, storedName, mimeType, header.Size, strings.ToUpper(ext))
	if err != nil {
		return nil, err
	}
	s.repo.LogActivity(int64(projectID), u.ID, u.Name, "document", "Uploaded “"+name+"”")
	s.repo.TouchProject(int64(projectID))
	return s.repo.GetDocument(id)
}

// Locate resolves the on-disk file and metadata for a document download.
func (s *Service) Locate(id int64) (map[string]any, string, error) {
	doc, err := s.repo.GetDocument(id)
	if err != nil {
		return nil, "", ErrDocumentNotFound
	}
	storedName, _ := doc["stored_name"].(string)
	if strings.HasPrefix(storedName, "seeded/") {
		return nil, "", ErrRecordOnly
	}
	filePath := filepath.Join(s.uploadsDir, filepath.Base(storedName))
	if _, err := os.Stat(filePath); err != nil {
		return nil, "", ErrFileUnavailable
	}
	return doc, filePath, nil
}
