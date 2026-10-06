import type { RouteDef } from "../router"
import {
  ApiError,
  asDate,
  asOneOf,
  asString,
  json,
  logActivity,
  readJsonBody,
  requireUser,
  touchProject,
  type SqlParam,
} from "../http"
import { getDb } from "../db/db"

const STATUSES = ["not-started", "in-progress", "blocked", "completed"] as const
const PRIORITIES = ["high", "medium", "low"] as const

const TASK_SELECT = `
  SELECT t.*, p.name AS project_name, u.name AS assignee_name, u.initials AS assignee_initials, u.color AS assignee_color
  FROM tasks t
  JOIN projects p ON p.id = t.project_id
  LEFT JOIN users u ON u.id = t.assignee_id`

export const taskRoutes: RouteDef[] = [
  {
    method: "GET",
    path: "/api/tasks",
    handler: async (ctx) => {
      requireUser(ctx)
      const db = getDb()
      const conditions: string[] = []
      const values: SqlParam[] = []
      const projectId = ctx.query.get("projectId")
      if (projectId) {
        conditions.push("t.project_id = ?")
        values.push(Number(projectId))
      }
      const status = ctx.query.get("status")
      if (status && status !== "All") {
        conditions.push("t.status = ?")
        values.push(status)
      }
      const q = ctx.query.get("q")
      if (q) {
        conditions.push("t.name LIKE ?")
        values.push(`%${q}%`)
      }
      const sql =
        TASK_SELECT +
        (conditions.length ? " WHERE " + conditions.join(" AND ") : "") +
        " ORDER BY t.due_date IS NULL, t.due_date, t.id"
      json(ctx.res, 200, { tasks: db.prepare(sql).all(...values) })
    },
  },
  {
    method: "POST",
    path: "/api/tasks",
    handler: async (ctx) => {
      const user = requireUser(ctx)
      if (user.role === "viewer")
        throw new ApiError(403, "You do not have permission to create tasks.")
      const body = (await readJsonBody(ctx.req)) as Record<string, unknown>
      const projectId = Number(body.projectId)
      const name = asString(body.name, "Task name", {
        required: true,
        max: 200,
      })
      const description = asString(body.description, "Description", {
        max: 2000,
      })
      const assigneeId =
        body.assigneeId === null ||
        body.assigneeId === undefined ||
        body.assigneeId === ""
          ? null
          : Number(body.assigneeId)
      const startDate = asDate(body.startDate, "Start date")
      const dueDate = asDate(body.dueDate, "Due date", { required: true })
      const status = asOneOf(body.status ?? "not-started", "Status", STATUSES)
      const priority = asOneOf(
        body.priority ?? "medium",
        "Priority",
        PRIORITIES,
      )

      const db = getDb()
      const project = db
        .prepare("SELECT id, name FROM projects WHERE id = ?")
        .get(projectId) as { id: number; name: string } | undefined
      if (!project) throw new ApiError(400, "Select a valid project.")
      if (assigneeId !== null) {
        const assignee = db
          .prepare("SELECT id FROM users WHERE id = ?")
          .get(assigneeId)
        if (!assignee) throw new ApiError(400, "Select a valid assignee.")
      }

      const result = db
        .prepare(
          `INSERT INTO tasks (project_id, assignee_id, name, description, start_date, due_date, status, priority)
         VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
        )
        .run(
          projectId,
          assigneeId,
          name,
          description,
          startDate,
          dueDate,
          status,
          priority,
        )
      logActivity(projectId, user, "task", `Created task “${name}”`)
      touchProject(projectId)
      const row = db
        .prepare(TASK_SELECT + " WHERE t.id = ?")
        .get(Number(result.lastInsertRowid))
      json(ctx.res, 201, { task: row })
    },
  },
  {
    method: "PATCH",
    path: "/api/tasks/:id",
    handler: async (ctx) => {
      const user = requireUser(ctx)
      if (user.role === "viewer")
        throw new ApiError(403, "You do not have permission to update tasks.")
      const db = getDb()
      const id = Number(ctx.params.id)
      const existing = db
        .prepare("SELECT * FROM tasks WHERE id = ?")
        .get(id) as Record<string, unknown> | undefined
      if (!existing) throw new ApiError(404, "Task not found.")

      const body = (await readJsonBody(ctx.req)) as Record<string, unknown>
      const updates: string[] = []
      const values: SqlParam[] = []

      if (body.status !== undefined) {
        updates.push("status = ?")
        values.push(asOneOf(body.status, "Status", STATUSES))
      }
      if (body.priority !== undefined) {
        updates.push("priority = ?")
        values.push(asOneOf(body.priority, "Priority", PRIORITIES))
      }
      if (body.name !== undefined) {
        updates.push("name = ?")
        values.push(
          asString(body.name, "Task name", { required: true, max: 200 }),
        )
      }
      if (body.description !== undefined) {
        updates.push("description = ?")
        values.push(asString(body.description, "Description", { max: 2000 }))
      }
      if (body.assigneeId !== undefined) {
        const assigneeId =
          body.assigneeId === null || body.assigneeId === ""
            ? null
            : Number(body.assigneeId)
        if (assigneeId !== null) {
          const assignee = db
            .prepare("SELECT id FROM users WHERE id = ?")
            .get(assigneeId)
          if (!assignee) throw new ApiError(400, "Select a valid assignee.")
        }
        updates.push("assignee_id = ?")
        values.push(assigneeId)
      }
      if (body.dueDate !== undefined) {
        updates.push("due_date = ?")
        values.push(asDate(body.dueDate, "Due date", { required: true }))
      }
      if (updates.length === 0) throw new ApiError(400, "No fields to update.")

      values.push(id)
      db.prepare(
        `UPDATE tasks SET ${updates.join(", ")}, updated_at = datetime('now') WHERE id = ?`,
      ).run(...values)

      const statusChanged =
        body.status !== undefined && body.status !== existing.status
      if (statusChanged) {
        const labels: Record<string, string> = {
          "not-started": "Not started",
          "in-progress": "In progress",
          blocked: "Blocked",
          completed: "Completed",
        }
        logActivity(
          existing.project_id as number,
          user,
          "task",
          `Marked “${existing.name}” as ${labels[(body.status as string)]}`,
        )
      }
      touchProject(existing.project_id as number)
      const row = db.prepare(TASK_SELECT + " WHERE t.id = ?").get(id)
      json(ctx.res, 200, { task: row })
    },
  },
]
