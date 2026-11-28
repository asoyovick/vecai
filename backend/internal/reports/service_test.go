package reports

import (
	"context"
	"errors"
	"testing"
	"time"
)

type fakeRepo struct {
	reports map[string]*Report
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{reports: make(map[string]*Report)}
}

func (f *fakeRepo) Create(ctx context.Context, report *Report) error {
	f.reports[report.ProjectID+"|"+report.ID] = report
	return nil
}

func (f *fakeRepo) GetByID(ctx context.Context, projectID, reportID string) (*Report, error) {
	r, ok := f.reports[projectID+"|"+reportID]
	if !ok {
		return nil, ErrReportNotFound
	}
	return r, nil
}

func (f *fakeRepo) ListByProject(ctx context.Context, projectID string) ([]*Report, error) {
	var out []*Report
	for k, r := range f.reports {
		if len(k) > len(projectID)+1 && k[:len(projectID)+1] == projectID+"|" {
			out = append(out, r)
		}
	}
	return out, nil
}

func (f *fakeRepo) Update(ctx context.Context, report *Report) error {
	key := report.ProjectID + "|" + report.ID
	if _, ok := f.reports[key]; !ok {
		return ErrReportNotFound
	}
	f.reports[key] = report
	return nil
}

func (f *fakeRepo) Delete(ctx context.Context, projectID, reportID string) error {
	key := projectID + "|" + reportID
	if _, ok := f.reports[key]; !ok {
		return ErrReportNotFound
	}
	delete(f.reports, key)
	return nil
}

func TestServiceUpdateReport(t *testing.T) {
	repo := newFakeRepo()
	svc := NewService(repo)

	// Seed a report
	repo.reports["p1|r1"] = &Report{
		ID:             "r1",
		ProjectID:      "p1",
		UserID:         "u1",
		ReportDate:     "2026-01-01",
		CompletedWork:  "old",
		WorkersPresent: 5,
		Remarks:        "old",
		CreatedAt:      time.Now().UTC(),
	}

	updated, err := svc.Update(context.Background(), "p1", "r1", "u1", CreateRequest{
		ReportDate:     "2026-01-02",
		CompletedWork:  "new work",
		WorkersPresent: 10,
		Remarks:        "new remarks",
	})
	if err != nil {
		t.Fatalf("update failed: %v", err)
	}
	if updated.ReportDate != "2026-01-02" {
		t.Fatalf("expected date 2026-01-02, got %s", updated.ReportDate)
	}
	if updated.CompletedWork != "new work" {
		t.Fatalf("expected updated work, got %s", updated.CompletedWork)
	}
}

func TestServiceUpdateReportNotFound(t *testing.T) {
	repo := newFakeRepo()
	svc := NewService(repo)

	_, err := svc.Update(context.Background(), "p1", "missing", "u1", CreateRequest{
		ReportDate:     "2026-01-02",
		CompletedWork:  "work",
		WorkersPresent: 1,
	})
	if !errors.Is(err, ErrReportNotFound) {
		t.Fatalf("expected ErrReportNotFound, got %v", err)
	}
}
