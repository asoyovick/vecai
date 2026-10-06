// Types mirrored from the REST API (internal/handlers/*).

export type Role = "admin" | "member" | "viewer"

export interface User {
  id: number
  email: string
  name: string
  role: Role
  jobTitle: string
  initials: string
  color: string
  phone: string
}

export interface TeamMember {
  id: number
  name: string
  role: Role
  roleLabel: string
  job_title: string
  initials: string
  color: string
  email: string
  phone: string
  project_count?: number
}

export interface ProjectStats {
  tasksTotal: number
  tasksCompleted: number
  tasksDueThisWeek: number
  tasksBlocked: number
  materialsAlert: number
  documentsRecent: number
}

export interface Project {
  id: number
  name: string
  location: string
  status: string
  progress: number
  startDate: string | null
  expectedCompletion: string | null
  client: string
  contractor: string
  budgetTotal: number
  budgetSpent: number
  description: string
  imageUrl: string
  stats: ProjectStats
  team: TeamMember[]
}

export interface Task {
  id: number
  project_id: number
  assignee_id: number | null
  name: string
  description: string
  start_date: string | null
  due_date: string | null
  status: "not-started" | "in-progress" | "blocked" | "completed"
  priority: "high" | "medium" | "low"
  project_name?: string
  assignee_name?: string | null
  assignee_initials?: string | null
  assignee_color?: string | null
}

export interface Material {
  id: number
  project_id: number
  supplier_id: number | null
  name: string
  unit: string
  required_qty: number
  available_qty: number
  used_qty: number
  status: "ok" | "low" | "critical" | "ordered"
  unit_cost: number
  last_delivery: string | null
  supplier_name?: string | null
  project_name?: string
}

export interface DocumentItem {
  id: number
  project_id: number
  uploaded_by: number | null
  name: string
  category: string
  version: string
  original_name: string
  mime_type: string
  size_bytes: number
  format: string
  created_at: string
  project_name?: string
  uploaded_by_name?: string | null
}

export interface SiteUpdate {
  id: number
  project_id: number
  submitted_by: number | null
  date: string
  weather: string
  temperature: string
  workers_on_site: number
  work_completed: string
  issues: string
  notes: string
  photo_count: number
  project_name?: string
  submitted_by_name?: string | null
  materialsUsed?: Array<{ name: string; quantity: number; unit: string }>
}

export interface Supplier {
  id: number
  name: string
  category: string
  contact: string
  phone: string
  email: string
  rating: number
  status: "active" | "inactive"
  location: string
  last_order: string | null
}

export interface ActivityItem {
  id: number
  project_id: number | null
  actor_id: number | null
  actor_name: string
  type: string
  text: string
  created_at: string
}

export interface BudgetMonth {
  ym: string
  planned: number
  actual: number
}

export interface DashboardData {
  stats: {
    active_projects: number
    tasks_due: number
    blocked_tasks: number
    material_alerts: number
  }
  projects: Project[]
  materialAlerts: Array<Pick<Material, "id" | "name" | "available_qty" | "required_qty" | "unit" | "status"> & {
    project_name: string
  }>
  deadlines: Array<Pick<Task, "id" | "name" | "due_date" | "status"> & {
    project_name: string
  }>
  activity: ActivityItem[]
}

export interface AssistantAnswer {
  title: string
  intro: string
  sections: Array<{
    heading?: string
    lines: Array<{
      text: string
      tone?: "default" | "danger" | "warn" | "success"
    }>
  }>
  footnote: string
}

// ── Fetch helper ──────────────────────────────────────────────────────────────

export class ApiRequestError extends Error {
  status: number
  constructor(status: number, message: string) {
    super(message)
    this.status = status
  }
}

async function request<T>(path: string, options: RequestInit = {}): Promise<T> {
  const hasBody = options.body !== undefined
  const res = await fetch(path, {
    ...options,
    headers: {
      ...(hasBody ? { "Content-Type": "application/json" } : {}),
      ...options.headers,
    },
    credentials: "same-origin",
  })
  let data: unknown = null
  const text = await res.text()
  if (text) {
    try {
      data = JSON.parse(text)
    } catch {
      // Non-JSON response body; leave data null.
    }
  }
  if (!res.ok) {
    const message =
      (data as { error?: string } | null)?.error ||
      (res.status === 401
        ? "Please sign in to continue."
        : "Something went wrong. Please try again.")
    throw new ApiRequestError(res.status, message)
  }
  return data as T
}

export const api = {
  get: <T>(path: string) => request<T>(path),
  post: <T>(path: string, body?: unknown) =>
    request<T>(path, {
      method: "POST",
      body: body === undefined ? undefined : JSON.stringify(body),
    }),
  patch: <T>(path: string, body: unknown) =>
    request<T>(path, { method: "PATCH", body: JSON.stringify(body) }),
}

export function uploadDocument(
  projectId: number,
  file: File,
  meta: { name: string; category: string; version: string },
): Promise<{ document: DocumentItem }> {
  const form = new FormData()
  form.set("file", file)
  form.set("projectId", String(projectId))
  form.set("name", meta.name)
  form.set("category", meta.category)
  form.set("version", meta.version)
  return fetch("/api/documents", {
    method: "POST",
    body: form,
    credentials: "same-origin",
  }).then(async (res) => {
    const data = await res.json().catch(() => null)
    if (!res.ok)
      throw new ApiRequestError(
        res.status,
        (data as { error?: string })?.error || "Upload failed.",
      )
    return data as { document: DocumentItem }
  })
}
