import { apiDelete, apiGet, apiPost, apiPut } from "./api";

export type BoqItem = {
  id: string;
  project_id: string;
  item_name: string;
  quantity: number;
  unit: string;
  estimated_cost: number;
  actual_cost: number;
  created_at: string;
  updated_at: string;
};

export type BoqSummary = {
  total_items: number;
  total_estimated_cost: number;
  total_actual_cost: number;
  variance: number;
};

export type BoqItemInput = {
  item_name: string;
  quantity: number;
  unit?: string;
  estimated_cost?: number;
  actual_cost?: number;
};

export function listBoqItems(token: string, projectId: string): Promise<BoqItem[]> {
  return apiGet<BoqItem[]>(`/api/v1/projects/${projectId}/boq`, token);
}

export function compareBoq(token: string, projectId: string): Promise<BoqSummary> {
  return apiGet<BoqSummary>(`/api/v1/projects/${projectId}/boq/compare`, token);
}

export function createBoqItem(token: string, projectId: string, input: BoqItemInput): Promise<BoqItem> {
  return apiPost<BoqItem>(`/api/v1/projects/${projectId}/boq`, input, token);
}

export function updateBoqItem(token: string, projectId: string, itemId: string, input: BoqItemInput): Promise<BoqItem> {
  return apiPut<BoqItem>(`/api/v1/projects/${projectId}/boq/${itemId}`, input, token);
}

export function deleteBoqItem(token: string, projectId: string, itemId: string): Promise<null> {
  return apiDelete<null>(`/api/v1/projects/${projectId}/boq/${itemId}`, token);
}
