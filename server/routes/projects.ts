import type { RouteDef } from "../router"
import {
  ApiError,
  asDate,
  asNumber,
  asOneOf,
  asString,
  json,
  logActivity,
  readJsonBody,
  requireUser,
  requireRole,
  touchProject,
  type SqlParam,
} from "../http"
import { getDb } from "../db/db"

const PROJECT_STATUSES = [
  "Tendering",
  "Design",
  "Foundation",
  "Construction",
  "Roofing",
  "Finishing",
  "Handover",
  "Completed",
] as const

interface ProjectStats {
  tasksTotal: number
  tasksCompleted: number
  tasksDueThisWeek: number
  tasksBlocked: number
  materialsAlert: number
  documentsRecent: number
}

function projectStats(
  db: ReturnType<typeof getDb>,
  projectId: number,
): ProjectStats {
  const tasks = db
    .prepare(
      `SELECT
       COUNT(*) AS total,
       SUM(CASE WHEN status = 'completed' THEN 1 ELSE 0 END) AS completed,
       SUM(CASE WHEN status != 'completed' AND due_date IS NOT NULL AND due_date <= date('now', '+7 days') THEN 1 ELSE 0 END) AS dueWeek,
       SUM(CASE WHEN status = 'blocked' THEN 1 ELSE 0 END) AS blocked
     FROM tasks WHERE project_id = ?`,
    )
    .get(projectId) as {
    total: number
    completed: number | null
    dueWeek: number | null
    blocked: number | null
  }
  const alerts = (db
    .prepare(
      "SELECT COUNT(*) AS c FROM materials WHERE project_id = ? AND status IN ('critical','low')",
    )
    .get(projectId) as { c: number }).c
  const recentDocs = (db
    .prepare(
      "SELECT COUNT(*) AS c FROM documents WHERE project_id = ? AND created_at >= datetime('now', '-14 days')",
    )
    .get(projectId) as { c: number }).c
  return {
    tasksTotal: tasks.total,
    tasksCompleted: tasks.completed ?? 0,
    tasksDueThisWeek: tasks.dueWeek ?? 0,
    tasksBlocked: tasks.blocked ?? 0,
    materialsAlert: alerts,
    documentsRecent: recentDocs,
  }
}

function mapProject(
  db: ReturnType<typeof getDb>,
  row: Record<string, unknown>,
) {
  return {
    id: row.id,
    name: row.name,
    location: row.location,
    status: row.status,
    progress: row.progress,
    startDate: row.start_date,
    expectedCompletion: row.expected_completion,
    client: row.client,
    contractor: row.contractor,
    budgetTotal: row.budget_total,
    budgetSpent: row.budget_spent,
    description: row.description,
    imageUrl: row.image_url,
    createdAt: row.created_at,
    updatedAt: row.updated_at,
    stats: projectStats(db, row.id as number),
    team: db
      .prepare(
        `SELECT u.id, u.name, u.role, u.job_title AS roleLabel, u.initials, u.color, u.email, u.phone
       FROM project_members pm JOIN users u ON u.id = pm.user_id WHERE pm.project_id = ? ORDER BY u.name`,
      )
      .all(Number(row.id)),
  }
}

export const projectRoutes: RouteDef[] = [
  {
    method: "GET",
    path: "/api/projects",
    handler: async (ctx) => {
      requireUser(ctx)
      const db = getDb()
      const rows = db
        .prepare("SELECT * FROM projects ORDER BY created_at DESC")
        .all() as Record<string, unknown>[]
      json(ctx.res, 200, { projects: rows.map((r) => mapProject(db, r)) })
    },
  },
  {
    method: "POST",
    path: "/api/projects",
    handler: async (ctx) => {
      const user = requireRole(ctx, "admin")
      const body = (await readJsonBody(ctx.req)) as Record<string, unknown>
      const name = asString(body.name, "Project name", {
        required: true,
        max: 160,
      })
      const location = asString(body.location, "Location", { max: 160 })
      const status = asOneOf(
        body.status ?? "Tendering",
        "Status",
        PROJECT_STATUSES,
      )
      const startDate = asDate(body.startDate, "Start date")
      const completion = asDate(body.expectedCompletion, "Expected completion")
      const client = asString(body.client, "Client", { max: 160 })
      const contractor = asString(body.contractor, "Contractor", { max: 160 })
      const budgetTotal = asNumber(body.budgetTotal, "Total budget", {
        min: 0,
        integer: true,
      })
      const description = asString(body.description, "Description", {
        max: 4000,
      })

      const db = getDb()
      const result = db
        .prepare(
          `INSERT INTO projects (name, location, status, start_date, expected_completion, client, contractor, budget_total, description)
         VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
        )
        .run(
          name,
          location,
          status,
          startDate,
          completion,
          client,
          contractor,
          budgetTotal,
          description,
        )
      const id = Number(result.lastInsertRowid)

      // Add the creating admin to the project team.
      db.prepare(
        "INSERT INTO project_members (project_id, user_id) VALUES (?, ?)",
      ).run(id, user.id)
      logActivity(id, user, "project", `Created project “${name}”`)
      json(ctx.res, 201, {
        project: mapProject(
          db,
          db
            .prepare("SELECT * FROM projects WHERE id = ?")
            .get(id) as Record<string, unknown>,
        ),
      })
    },
  },
  {
    method: "GET",
    path: "/api/projects/:id",
    handler: async (ctx) => {
      requireUser(ctx)
      const db = getDb()
      const row = db
        .prepare("SELECT * FROM projects WHERE id = ?")
        .get(Number(ctx.params.id)) as Record<string, unknown> | undefined
      if (!row) throw new ApiError(404, "Project not found.")

      const project = mapProject(db, row)
      const tasks = db
        .prepare(
          `SELECT t.*, u.name AS assignee_name, u.initials AS assignee_initials, u.color AS assignee_color
         FROM tasks t LEFT JOIN users u ON u.id = t.assignee_id
         WHERE t.project_id = ? ORDER BY t.due_date IS NULL, t.due_date, t.id`,
        )
        .all(Number(ctx.params.id))
      const materials = db
        .prepare(
          `SELECT m.*, s.name AS supplier_name FROM materials m LEFT JOIN suppliers s ON s.id = m.supplier_id
         WHERE m.project_id = ? ORDER BY m.name`,
        )
        .all(Number(ctx.params.id))
      const documents = db
        .prepare(
          `SELECT d.*, u.name AS uploaded_by_name FROM documents d LEFT JOIN users u ON u.id = d.uploaded_by
         WHERE d.project_id = ? ORDER BY d.created_at DESC`,
        )
        .all(Number(ctx.params.id))
      const activity = db
        .prepare(
          "SELECT * FROM activity_log WHERE project_id = ? ORDER BY created_at DESC, id DESC LIMIT 30",
        )
        .all(Number(ctx.params.id))
      const budget = db
        .prepare(
          "SELECT ym, planned, actual FROM budget_months WHERE project_id = ? ORDER BY ym",
        )
        .all(Number(ctx.params.id))

      json(ctx.res, 200, {
        project,
        tasks,
        materials,
        documents,
        activity,
        budget,
      })
    },
  },
  {
    method: "PATCH",
    path: "/api/projects/:id",
    handler: async (ctx) => {
      const user = requireRole(ctx, "admin")
      const db = getDb()
      const id = Number(ctx.params.id)
      const existing = db
        .prepare("SELECT * FROM projects WHERE id = ?")
        .get(id) as Record<string, unknown> | undefined
      if (!existing) throw new ApiError(404, "Project not found.")

      const body = (await readJsonBody(ctx.req)) as Record<string, unknown>
      const updates: string[] = []
      const values: SqlParam[] = []

      if (body.name !== undefined) {
        updates.push("name = ?")
        values.push(
          asString(body.name, "Project name", { required: true, max: 160 }),
        )
      }
      if (body.location !== undefined) {
        updates.push("location = ?")
        values.push(asString(body.location, "Location", { max: 160 }))
      }
      if (body.status !== undefined) {
        updates.push("status = ?")
        values.push(asOneOf(body.status, "Status", PROJECT_STATUSES))
      }
      if (body.progress !== undefined) {
        updates.push("progress = ?")
        values.push(
          asNumber(body.progress, "Progress", { min: 0, integer: true }),
        )
      }
      if (body.budgetSpent !== undefined) {
        updates.push("budget_spent = ?")
        values.push(
          asNumber(body.budgetSpent, "Budget spent", { min: 0, integer: true }),
        )
      }
      if (body.description !== undefined) {
        updates.push("description = ?")
        values.push(asString(body.description, "Description", { max: 4000 }))
      }
      if (updates.length === 0) throw new ApiError(400, "No fields to update.")

      values.push(id)
      db.prepare(
        `UPDATE projects SET ${updates.join(", ")}, updated_at = datetime('now') WHERE id = ?`,
      ).run(...values)

      if (body.progress !== undefined) {
        logActivity(
          id,
          user,
          "project",
          `Updated progress to ${Number(body.progress)}%`,
        )
      } else {
        logActivity(id, user, "project", `Updated project details`)
      }
      touchProject(id)
      json(ctx.res, 200, {
        project: mapProject(
          db,
          db
            .prepare("SELECT * FROM projects WHERE id = ?")
            .get(id) as Record<string, unknown>,
        ),
      })
    },
  },
  {
    method: "GET",
    path: "/api/projects/:id/budget",
    handler: async (ctx) => {
      requireUser(ctx)
      const db = getDb()
      const rows = db
        .prepare(
          "SELECT ym, planned, actual FROM budget_months WHERE project_id = ? ORDER BY ym",
        )
        .all(Number(ctx.params.id))
      json(ctx.res, 200, { budget: rows })
    },
  },
]
