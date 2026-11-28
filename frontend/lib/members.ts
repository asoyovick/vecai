import { apiDelete, apiGet, apiPost } from "./api";

export type ProjectMember = {
  id: string;
  project_id: string;
  user_id: string;
  name?: string;
  email?: string;
  role: string;
  joined_at: string;
};

export function listProjectMembers(
  token: string,
  projectId: string,
): Promise<ProjectMember[]> {
  return apiGet<ProjectMember[]>(`/api/v1/projects/${projectId}/members`, token);
}

export function inviteProjectMember(
  token: string,
  projectId: string,
  userId: string,
  role: string,
): Promise<ProjectMember> {
  return apiPost<ProjectMember>(
    `/api/v1/projects/${projectId}/members`,
    { user_id: userId, role },
    token,
  );
}

export function removeProjectMember(
  token: string,
  projectId: string,
  memberId: string,
): Promise<null> {
  return apiDelete<null>(
    `/api/v1/projects/${projectId}/members/${memberId}`,
    token,
  );
}
