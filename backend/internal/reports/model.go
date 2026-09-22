package reports

import (
	"errors"
	"time"
)

var (
	ErrReportNotFound = errors.New("daily report not found")
	ErrInvalidInput   = errors.New("invalid report input")
)

// Report mirrors the daily_reports table from docs/database.md.
type Report struct {
	ID             string    `json:"id"`
	ProjectID      string    `json:"project_id"`
	UserID         string    `json:"user_id"`
	ReportDate     string    `json:"report_date"`
	CompletedWork  string    `json:"completed_work"`
	WorkersPresent int       `json:"workers_present"`
	Remarks        string    `json:"remarks"`
	CreatedAt      time.Time `json:"created_at"`
}
