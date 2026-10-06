package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"vecai/internal/httpx"
	"vecai/internal/middleware"
)

func timeNowMs() int64         { return time.Now().UnixMilli() }
func mustInt64(s string) int64 { v, _ := strconv.ParseInt(s, 10, 64); return v }

var documentCategories = []string{"Architectural Drawing", "Structural Drawing", "BOQ", "Contract", "Permit", "Invoice", "Site Report", "Inspection Report", "Photo"}

var allowedExtensions = map[string]bool{
	"pdf": true, "xls": true, "xlsx": true, "csv": true, "doc": true, "docx": true,
	"png": true, "jpg": true, "jpeg": true, "webp": true, "zip": true, "dwg": true,
	"ifc": true, "txt": true,
}

const docSelect = `SELECT d.*, p.name AS project_name, u.name AS uploaded_by_name
	FROM documents d JOIN projects p ON p.id = d.project_id LEFT JOIN users u ON u.id = d.uploaded_by`

func (s *Server) ListDocuments(w http.ResponseWriter, r *http.Request) {
	if middleware.RequireUser(w, r) == nil {
		return
	}
	q := docSelect
	var args []any
	var conds []string
	if v := r.URL.Query().Get("projectId"); v != "" && v != "All" {
		conds = append(conds, "d.project_id = ?")
		args = append(args, v)
	}
	if v := r.URL.Query().Get("category"); v != "" && v != "All" {
		conds = append(conds, "d.category = ?")
		args = append(args, v)
	}
	if v := r.URL.Query().Get("q"); v != "" {
		conds = append(conds, "d.name LIKE ?")
		args = append(args, "%"+v+"%")
	}
	if len(conds) > 0 {
		q += " WHERE " + strings.Join(conds, " AND ")
	}
	q += " ORDER BY d.created_at DESC"
	rows, err := httpx.QueryMaps(s.DB, q, args...)
	if err != nil {
		httpx.Error(w, err)
		return
	}
	if rows == nil {
		rows = []map[string]any{}
	}
	httpx.JSON(w, 200, map[string]any{"documents": rows})
}

func sanitizeFilename(name string) (base, ext string) {
	ext = strings.ToLower(strings.TrimPrefix(filepath.Ext(name), "."))
	base = strings.TrimSuffix(name, filepath.Ext(name))
	var b strings.Builder
	for _, r := range base {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '-', r == '_', r == ' ':
			b.WriteRune(r)
		}
	}
	base = strings.TrimSpace(b.String())
	if len(base) > 80 {
		base = base[:80]
	}
	if base == "" {
		base = "file"
	}
	return base, ext
}

func (s *Server) UploadDocument(w http.ResponseWriter, r *http.Request) {
	u := middleware.RequireUser(w, r)
	if u == nil || !middleware.NotViewer(w, u, "You do not have permission to upload documents.") {
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 25<<20)
	if err := r.ParseMultipartForm(25 << 20); err != nil {
		httpx.Error(w, httpx.Err(400, "Could not parse the uploaded file."))
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		httpx.Error(w, httpx.Err(400, "Select a file to upload."))
		return
	}
	defer file.Close()
	name := strings.TrimSpace(r.FormValue("name"))
	category, err := httpx.AsOneOf(r.FormValue("category"), "Category", documentCategories)
	if err != nil {
		httpx.Error(w, err)
		return
	}
	version := strings.TrimSpace(r.FormValue("version"))
	if version == "" {
		version = "1.0"
	}
	projectID := r.FormValue("projectId")
	if name == "" {
		httpx.Error(w, httpx.Err(400, "Document name is required."))
		return
	}
	if header.Size == 0 {
		httpx.Error(w, httpx.Err(400, "The selected file is empty."))
		return
	}
	if header.Size > 25<<20 {
		httpx.Error(w, httpx.Err(400, "Files are limited to 25 MB."))
		return
	}
	_, ext := sanitizeFilename(header.Filename)
	if ext != "" && !allowedExtensions[ext] {
		httpx.Error(w, httpx.Err(400, "File type “."+ext+"” is not allowed."))
		return
	}
	var c int
	s.DB.QueryRow("SELECT COUNT(*) FROM projects WHERE id = ?", projectID).Scan(&c)
	if c == 0 {
		httpx.Error(w, httpx.Err(400, "Select a valid project."))
		return
	}
	if err := os.MkdirAll(s.Cfg.UploadsDir, 0o755); err != nil {
		httpx.Error(w, err)
		return
	}
	randBytes := make([]byte, 6)
	rand.Read(randBytes)
	storedName := strconv.FormatInt(timeNowMs(), 10) + "-" + hex.EncodeToString(randBytes)
	if ext != "" {
		storedName += "." + ext
	}
	dst, err := os.Create(filepath.Join(s.Cfg.UploadsDir, storedName))
	if err != nil {
		httpx.Error(w, err)
		return
	}
	if _, err := io.Copy(dst, file); err != nil {
		dst.Close()
		httpx.Error(w, err)
		return
	}
	dst.Close()
	mime := header.Header.Get("Content-Type")
	if mime == "" {
		mime = "application/octet-stream"
	}
	res, err := s.DB.Exec(`INSERT INTO documents (project_id, uploaded_by, name, category, version, original_name, stored_name, mime_type, size_bytes, format) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		projectID, u.ID, name, category, version, header.Filename, storedName, mime, header.Size, strings.ToUpper(ext))
	if err != nil {
		httpx.Error(w, err)
		return
	}
	id, _ := res.LastInsertId()
	httpx.LogActivity(s.DB, projectID, &httpx.Actor{ID: u.ID, Name: u.Name}, "document", "Uploaded “"+name+"”")
	httpx.TouchProject(s.DB, mustInt64(projectID))
	row, err := httpx.QueryMap(s.DB, docSelect+" WHERE d.id = ?", id)
	if err != nil {
		httpx.Error(w, err)
		return
	}
	httpx.JSON(w, 201, map[string]any{"document": row})
}

func (s *Server) DownloadDocument(w http.ResponseWriter, r *http.Request) {
	if middleware.RequireUser(w, r) == nil {
		return
	}
	doc, err := httpx.QueryMap(s.DB, "SELECT * FROM documents WHERE id = ?", r.PathValue("id"))
	if err != nil {
		httpx.Error(w, httpx.Err(404, "Document not found."))
		return
	}
	storedName, _ := doc["stored_name"].(string)
	if strings.HasPrefix(storedName, "seeded/") {
		httpx.Error(w, httpx.Err(404, "This document was imported as a record only — no file is attached."))
		return
	}
	filePath := filepath.Join(s.Cfg.UploadsDir, filepath.Base(storedName))
	f, err := os.Open(filePath)
	if err != nil {
		httpx.Error(w, httpx.Err(404, "The file is no longer available."))
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
