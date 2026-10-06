// Package documents handles upload, listing and download of project documents.
package documents

import (
	"errors"
	"regexp"
	"strings"
)

var (
	ErrDocumentNotFound = errors.New("Document not found.")
	ErrFileUnavailable  = errors.New("The file is no longer available.")
	ErrRecordOnly       = errors.New("This document was imported as a record only — no file is attached.")
)

var Categories = []string{"Architectural Drawing", "Structural Drawing", "BOQ", "Contract", "Permit", "Invoice", "Site Report", "Inspection Report", "Photo"}

var allowedExtensions = map[string]bool{
	"pdf": true, "xls": true, "xlsx": true, "csv": true, "doc": true, "docx": true,
	"png": true, "jpg": true, "jpeg": true, "webp": true, "zip": true, "dwg": true,
	"ifc": true, "txt": true,
}

var unsafeFilenameChars = regexp.MustCompile(`[^a-zA-Z0-9\-_ ]+`)

func sanitizeFilename(name string) (base, ext string) {
	idx := strings.LastIndex(name, ".")
	if idx >= 0 {
		ext = strings.ToLower(name[idx+1:])
		base = name[:idx]
	} else {
		base = name
	}
	base = strings.TrimSpace(unsafeFilenameChars.ReplaceAllString(base, ""))
	if len(base) > 80 {
		base = base[:80]
	}
	if base == "" {
		base = "file"
	}
	return base, ext
}
