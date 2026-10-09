package documents

// Repository is the persistence contract for project documents.
type Repository interface {
	ListDocuments(projectID, category, q any) ([]map[string]any, error)
	GetDocument(id int64) (map[string]any, error)
	CreateDocument(projectID int64, uploadedBy int64, name, category, version, originalName, storedName, mimeType string, sizeBytes int64, format string) (int64, error)
	ProjectExists(id int64) (bool, error)
	TouchProject(projectID int64) error
	LogActivity(projectID any, actorID any, name, typ, text string)
}
