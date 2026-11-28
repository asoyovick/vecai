import { apiDelete, apiGet, apiPost } from "./api";

export type DailyReport = {
  id: string;
  project_id: string;
  user_id: string;
  report_date: string;
  completed_work: string;
  workers_present: number;
  remarks: string;
  created_at: string;
};

export type ReportSummary = {
  total_reports: number;
  average_workers: number;
  peak_workers: number;
  latest_report_date?: string;
};

export type CreateReportInput = {
  date: string;
  completed_work: string;
  workers: number;
  remarks?: string;
};

export function listReports(token: string, projectId: string): Promise<DailyReport[]> {
  return apiGet<DailyReport[]>(`/api/v1/projects/${projectId}/reports`, token);
}

export function getReportSummary(token: string, projectId: string): Promise<ReportSummary> {
  return apiGet<ReportSummary>(`/api/v1/projects/${projectId}/reports/summary`, token);
}

export function createReport(
  token: string,
  projectId: string,
  input: CreateReportInput,
): Promise<DailyReport> {
  return apiPost<DailyReport>(`/api/v1/projects/${projectId}/reports`, input, token);
}

export function deleteReport(
  token: string,
  projectId: string,
  reportId: string,
): Promise<null> {
  return apiDelete<null>(`/api/v1/projects/${projectId}/reports/${reportId}`, token);
}
