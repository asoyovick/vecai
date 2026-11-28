import { apiDelete, apiGet, apiPost, apiPut } from "./api";

export type Material = {
  id: string;
  project_id: string;
  material_name: string;
  quantity: number;
  unit: string;
  unit_cost: number;
  supplier: string;
  delivery_date: string;
  recorded_by: string;
  created_at: string;
};

export type MaterialSummary = {
  total_items: number;
  total_quantity: number;
  total_cost: number;
  [key: string]: unknown;
};

export type CreateMaterialInput = {
  material_name: string;
  quantity: number;
  unit: string;
  unit_cost?: number;
  supplier?: string;
  delivery_date?: string;
};

export type UpdateMaterialInput = Partial<CreateMaterialInput>;

export function listMaterials(token: string, projectId: string): Promise<Material[]> {
  return apiGet<Material[]>(`/api/v1/projects/${projectId}/materials`, token);
}

export function getMaterialsSummary(token: string, projectId: string): Promise<MaterialSummary> {
  return apiGet<MaterialSummary>(`/api/v1/projects/${projectId}/materials/summary`, token);
}

export function createMaterial(
  token: string,
  projectId: string,
  input: CreateMaterialInput,
): Promise<Material> {
  return apiPost<Material>(`/api/v1/projects/${projectId}/materials`, input, token);
}

export function updateMaterial(
  token: string,
  projectId: string,
  materialId: string,
  input: UpdateMaterialInput,
): Promise<Material> {
  return apiPut<Material>(`/api/v1/projects/${projectId}/materials/${materialId}`, input, token);
}

export function deleteMaterial(
  token: string,
  projectId: string,
  materialId: string,
): Promise<null> {
  return apiDelete<null>(`/api/v1/projects/${projectId}/materials/${materialId}`, token);
}
