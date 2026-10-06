package documents

import (
	"errors"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	"vecai/internal/api"
	"vecai/internal/auth"
)

// Handler serves the documents endpoints.
type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

// Routes registers the documents endpoints on the mux.
func (h *Handler) Routes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/documents", h.ListDocuments)
	mux.HandleFunc("POST /api/documents", h.UploadDocument)
	mux.HandleFunc("GET /api/documents/{id}/download", h.DownloadDocument)
}

func (h *Handler) ListDocuments(w http.ResponseWriter, r *http.Request) {
	if auth.RequireUser(w, r) == nil {
		return
	}
	var projectID, category, q any
	if v := r.URL.Query().Get("projectId"); v != "" && v != "All" {
		projectID = v
	}
	if v := r.URL.Query().Get("category"); v != "" && v != "All" {
		category = v
	}
	if v := r.URL.Query().Get("q"); v != "" {
		q = "%" + v + "%"
	}
	documents, err := h.svc.ListDocuments(projectID, category, q)
	if err != nil {
		api.Error(w, 500, "Could not load documents.")
		return
	}
	api.JSON(w, 200, map[string]any{"documents": documents})
}

func (h *Handler) UploadDocument(w http.ResponseWriter, r *http.Request) {
	u := auth.RequireUser(w, r)
	if u == nil {
		return
	}
	if !auth.NotViewer(w, u, "You do not have permission to upload documents.") {
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 25<<20)
	if err := r.ParseMultipartForm(25 << 20); err != nil {
		api.Error(w, 400, "Could not parse the uploaded file.")
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		api.Error(w, 400, "Select a file to upload.")
		return
	}
	defer file.Close()
	document, err := h.svc.Upload(u, file, header, r.FormValue("projectId"), strings.TrimSpace(r.FormValue("name")), r.FormValue("category"), strings.TrimSpace(r.FormValue("version")))
	if err != nil {
		api.Error(w, 400, err.Error())
		return
	}
	api.JSON(w, 201, map[string]any{"document": document})
}

func (h *Handler) DownloadDocument(w http.ResponseWriter, r *http.Request) {
	if auth.RequireUser(w, r) == nil {
		return
	}
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		api.Error(w, 400, "Invalid document id.")
		return
	}
	doc, filePath, err := h.svc.Locate(id)
	if err != nil {
		switch {
		case errors.Is(err, ErrDocumentNotFound):
			api.Error(w, 404, err.Error())
		case errors.Is(err, ErrRecordOnly):
			api.Error(w, 404, err.Error())
		default:
			api.Error(w, 404, ErrFileUnavailable.Error())
		}
		return
	}
	f, err := os.Open(filePath)
	if err != nil {
		api.Error(w, 404, ErrFileUnavailable.Error())
		return
	}
	defer f.Close()
	info, _ := f.Stat()
	mime, _ := doc["mime_type"].(string)
	if mime == "" {
		mime = "application/octet-stream"
	}
	original, _ := doc["original_name"].(string)
	safe := strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '.', r == '-', r == '_', r == ' ':
			return r
		default:
			return '_'
		}
	}, original)
	w.Header().Set("Content-Type", mime)
	w.Header().Set("Content-Length", strconv.FormatInt(info.Size(), 10))
	w.Header().Set("Content-Disposition", "attachment; filename=\""+safe+"\"")
	io.Copy(w, f)
}
