import { apiDelete, apiGet, API_BASE_URL, ApiError, type Envelope } from "./api";

export type ProjectDocument = {
  id: string;
  project_id: string;
  uploaded_by: string;
  filename: string;
  file_type: string;
  file_size: number;
  category?: string;
  uploaded_at: string;
};

export function listDocuments(token: string, projectId: string): Promise<ProjectDocument[]> {
  return apiGet<ProjectDocument[]>(`/api/v1/projects/${projectId}/documents`, token);
}

export async function uploadDocument(
  token: string,
  projectId: string,
  file: File,
  category?: string,
): Promise<ProjectDocument> {
  const form = new FormData();
  form.append("file", file);
  if (category) form.append("category", category);

  const res = await fetch(`${API_BASE_URL}/api/v1/projects/${projectId}/documents`, {
    method: "POST",
    headers: { Authorization: `Bearer ${token}` },
    body: form,
  });
  const json = (await res.json().catch(() => null)) as Envelope<ProjectDocument> | null;
  if (!res.ok || !json?.success) {
    throw new ApiError(res.status, json?.error ?? "Failed to upload document.");
  }
  return json.data as ProjectDocument;
}

export async function downloadDocument(
  token: string,
  projectId: string,
  documentId: string,
  filename: string,
): Promise<void> {
  const res = await fetch(`${API_BASE_URL}/api/v1/projects/${projectId}/documents/${documentId}`, {
    headers: { Authorization: `Bearer ${token}` },
  });
  if (!res.ok) {
    throw new ApiError(res.status, "Failed to download document.");
  }
  const blob = await res.blob();
  const url = URL.createObjectURL(blob);
  const a = document.createElement("a");
  a.href = url;
  a.download = filename;
  document.body.appendChild(a);
  a.click();
  a.remove();
  URL.revokeObjectURL(url);
}

export function deleteDocument(token: string, projectId: string, documentId: string): Promise<null> {
  return apiDelete<null>(`/api/v1/projects/${projectId}/documents/${documentId}`, token);
}
