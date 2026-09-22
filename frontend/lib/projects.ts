import { apiGet, apiPost } from "./api";

/** Mirrors backend/internal/project Project JSON (docs/api.md). */
export type Project = {
  id: string;
  owner_id: string;
  name: string;
  description: string;
  location: string;
  client_name: string;
  status: string;
  start_date: string;
  expected_end_date: string;
  created_at: string;
  updated_at: string;
};

export type CreateProjectInput = {
  name: string;
  description?: string;
  location: string;
  client_name?: string;
  status?: string;
  start_date?: string;
  expected_end_date?: string;
};

/** GET /api/v1/projects — projects owned by the authenticated user. */
export function listProjects(token: string): Promise<Project[]> {
  return apiGet<Project[]>("/api/v1/projects", token);
}

/** POST /api/v1/projects — create a project owned by the authenticated user. */
export function createProject(
  token: string,
  input: CreateProjectInput,
): Promise<Project> {
  return apiPost<Project>("/api/v1/projects", input, token);
}

/** Status values accepted by the backend (capitalized enum). */
export const PROJECT_STATUSES = [
  "Planning",
  "Active",
  "Paused",
  "Completed",
  "Archived",
] as const;
