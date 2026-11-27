import { apiDelete, apiGet, apiPost, apiPut } from "./api";

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

/** GET /api/v1/projects/:id — fetch a single project (owner or member). */
export function getProject(token: string, projectId: string): Promise<Project> {
  return apiGet<Project>(`/api/v1/projects/${projectId}`, token);
}

export type UpdateProjectInput = Partial<CreateProjectInput>;

/** PUT /api/v1/projects/:id — update project (owner only in practice). */
export function updateProject(
  token: string,
  projectId: string,
  input: UpdateProjectInput,
): Promise<Project> {
  return apiPut<Project>(`/api/v1/projects/${projectId}`, input, token);
}

/** DELETE /api/v1/projects/:id — archive/delete project (owner only). */
export function deleteProject(token: string, projectId: string): Promise<null> {
  return apiDelete<null>(`/api/v1/projects/${projectId}`, token);
}

/** Status values accepted by the backend (capitalized enum). */
export const PROJECT_STATUSES = [
  "Planning",
  "Active",
  "Paused",
  "Completed",
  "Archived",
] as const;
