package reports

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// Create records a daily site report for a project.
func (s *Service) Create(ctx context.Context, projectID, userID string, req CreateRequest) (*Report, error) {
	if err := validate(req); err != nil {
		return nil, err
	}

	report := &Report{
		ID:             uuid.NewString(),
		ProjectID:      projectID,
		UserID:         userID,
		ReportDate:     req.ReportDate,
		CompletedWork:  strings.TrimSpace(req.CompletedWork),
		WorkersPresent: req.WorkersPresent,
		Remarks:        strings.TrimSpace(req.Remarks),
	}

	if err := s.repo.Create(ctx, report); err != nil {
		return nil, fmt.Errorf("failed to create daily report: %w", err)
	}
	return report, nil
}

func (s *Service) Get(ctx context.Context, projectID, reportID string) (*Report, error) {
	return s.repo.GetByID(ctx, projectID, reportID)
}

func (s *Service) List(ctx context.Context, projectID string) ([]*Report, error) {
	return s.repo.ListByProject(ctx, projectID)
}

func (s *Service) Update(ctx context.Context, projectID, reportID string, req CreateRequest) (*Report, error) {
	if err := validate(req); err != nil {
		return nil, err
	}
	report, err := s.repo.GetByID(ctx, projectID, reportID)
	if err != nil {
		return nil, err
	}
	report.ReportDate = req.ReportDate
	report.CompletedWork = strings.TrimSpace(req.CompletedWork)
	report.WorkersPresent = req.WorkersPresent
	report.Remarks = strings.TrimSpace(req.Remarks)
	if err := s.repo.Update(ctx, report); err != nil {
		return nil, fmt.Errorf("failed to update daily report: %w", err)
	}
	return report, nil
}

func (s *Service) Delete(ctx context.Context, projectID, reportID string) error {
	return s.repo.Delete(ctx, projectID, reportID)
}

// Summary aggregates workforce figures for the project dashboard.
type Summary struct {
	TotalReports     int     `json:"total_reports"`
	AverageWorkers   float64 `json:"average_workers"`
	PeakWorkers      int     `json:"peak_workers"`
	LatestReportDate string  `json:"latest_report_date,omitempty"`
}

func (s *Service) Summarize(items []*Report) Summary {
	summary := Summary{TotalReports: len(items)}
	if len(items) == 0 {
		return summary
	}

	totalWorkers := 0
	latest := items[0].ReportDate
	for _, r := range items {
		totalWorkers += r.WorkersPresent
		if r.WorkersPresent > summary.PeakWorkers {
			summary.PeakWorkers = r.WorkersPresent
		}
		if r.ReportDate > latest {
			latest = r.ReportDate
		}
	}
	summary.AverageWorkers = float64(totalWorkers) / float64(len(items))
	summary.LatestReportDate = latest
	return summary
}

func validate(req CreateRequest) error {
	if req.ReportDate == "" {
		return fmt.Errorf("%w: report_date is required", ErrInvalidInput)
	}
	if _, err := time.Parse("2006-01-02", req.ReportDate); err != nil {
		return fmt.Errorf("%w: report_date must be YYYY-MM-DD", ErrInvalidInput)
	}
	if strings.TrimSpace(req.CompletedWork) == "" {
		return fmt.Errorf("%w: completed_work is required", ErrInvalidInput)
	}
	if req.WorkersPresent < 0 {
		return fmt.Errorf("%w: workers_present cannot be negative", ErrInvalidInput)
	}
	return nil
}
