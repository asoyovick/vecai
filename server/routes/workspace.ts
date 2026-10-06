import crypto from "node:crypto"
import type { RouteDef } from "../router"
import {
  ApiError,
  asNumber,
  asOneOf,
  asString,
  json,
  logActivity,
  readJsonBody,
  requireRole,
  requireUser,
  type SqlParam,
} from "../http"
import { getDb } from "../db/db"
import { hashPassword, verifyPassword } from "../auth"

const DEFAULT_PREFS = {
  "material-alerts": true,
  "task-reminders": true,
  "site-updates": true,
  "document-uploads": false,
  "budget-thresholds": true,
  "weekly-summary": true,
} as const

function initialsFor(name: string): string {
  return (
    name
      .split(/\s+/)
      .filter(Boolean)
      .slice(0, 2)
      .map((p) => p[0]?.toUpperCase() ?? "")
      .join("") || "U"
  )
}

export const workspaceRoutes: RouteDef[] = [
  {
    method: "GET",
    path: "/api/dashboard",
    handler: async (ctx) => {
      requireUser(ctx)
      const db = getDb()

      const projects = db
        .prepare("SELECT * FROM projects ORDER BY created_at DESC")
        .all() as Array<Record<string, unknown>>
      const stats = db
        .prepare(
          `SELECT
           (SELECT COUNT(*) FROM projects WHERE status != 'Completed') AS active_projects,
           (SELECT COUNT(*) FROM tasks WHERE status != 'completed' AND due_date IS NOT NULL AND due_date <= date('now', '+7 days')) AS tasks_due,
           (SELECT COUNT(*) FROM tasks WHERE status = 'blocked') AS blocked_tasks,
           (SELECT COUNT(*) FROM materials WHERE status IN ('critical','low')) AS material_alerts`,
        )
        .get() as {
        active_projects: number
        tasks_due: number
        blocked_tasks: number
        material_alerts: number
      }

      const materialAlerts = db
        .prepare(
          `SELECT m.id, m.name, m.available_qty, m.required_qty, m.unit, m.status, p.name AS project_name
         FROM materials m JOIN projects p ON p.id = m.project_id
         WHERE m.status IN ('critical','low')
         ORDER BY CASE m.status WHEN 'critical' THEN 0 ELSE 1 END, m.name LIMIT 8`,
        )
        .all()

      const deadlines = db
        .prepare(
          `SELECT t.id, t.name, t.due_date, t.status, p.name AS project_name
         FROM tasks t JOIN projects p ON p.id = t.project_id
         WHERE t.status != 'completed' AND t.due_date IS NOT NULL AND t.due_date <= date('now', '+14 days')
         ORDER BY t.due_date LIMIT 6`,
        )
        .all()

      const activity = db
        .prepare(
          "SELECT * FROM activity_log ORDER BY created_at DESC, id DESC LIMIT 8",
        )
        .all()

      json(ctx.res, 200, {
        stats,
        projects,
        materialAlerts,
        deadlines,
        activity,
      })
    },
  },
  {
    method: "GET",
    path: "/api/activity",
    handler: async (ctx) => {
      requireUser(ctx)
      const db = getDb()
      const projectId = ctx.query.get("projectId")
      const rows = projectId
        ? db
            .prepare(
              "SELECT * FROM activity_log WHERE project_id = ? ORDER BY created_at DESC, id DESC LIMIT 50",
            )
            .all(Number(projectId))
        : db
            .prepare(
              "SELECT * FROM activity_log ORDER BY created_at DESC, id DESC LIMIT 50",
            )
            .all()
      json(ctx.res, 200, { activity: rows })
    },
  },
  {
    method: "GET",
    path: "/api/team",
    handler: async (ctx) => {
      requireUser(ctx)
      const db = getDb()
      const members = db
        .prepare(
          `SELECT u.id, u.name, u.email, u.role, u.job_title, u.initials, u.color, u.phone, u.created_at,
                (SELECT COUNT(*) FROM project_members pm WHERE pm.user_id = u.id) AS project_count
         FROM users u ORDER BY u.name`,
        )
        .all()
      const assignments = db
        .prepare(
          `SELECT pm.user_id, p.id AS project_id, p.name AS project_name
         FROM project_members pm JOIN projects p ON p.id = pm.project_id`,
        )
        .all()
      json(ctx.res, 200, { members, assignments })
    },
  },
  {
    method: "POST",
    path: "/api/team/invite",
    handler: async (ctx) => {
      const admin = requireRole(ctx, "admin")
      const body = (await readJsonBody(ctx.req)) as Record<string, unknown>
      const name = asString(body.name, "Full name", {
        required: true,
        max: 120,
      })
      const email = asString(body.email, "Email", {
        required: true,
        max: 200,
      }).toLowerCase()
      const jobTitle = asString(body.jobTitle, "Job title", {
        required: true,
        max: 120,
      })
      const role = asOneOf(body.role ?? "member", "Role", [
        "member",
        "viewer",
      ] as const)
      const projectId = body.projectId ? Number(body.projectId) : null
      if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email))
        throw new ApiError(400, "Enter a valid email address.")

      const db = getDb()
      if (db.prepare("SELECT id FROM users WHERE email = ?").get(email)) {
        throw new ApiError(409, "A team member with that email already exists.")
      }

      const tempPassword = `vecai-${crypto.randomBytes(4).toString("hex")}`
      const colors = [
        "#1e3a5f",
        "#0891b2",
        "#7c3aed",
        "#db2777",
        "#059669",
        "#d97706",
        "#dc2626",
        "#0d9488",
      ]
      const color = colors[Math.floor(Math.random() * colors.length)]
      const result = db
        .prepare(
          `INSERT INTO users (email, password_hash, name, role, job_title, initials, color, phone)
         VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
        )
        .run(
          email,
          hashPassword(tempPassword),
          name,
          role,
          jobTitle,
          initialsFor(name),
          color,
          asString(body.phone, "Phone", { max: 40 }),
        )

      const newId = Number(result.lastInsertRowid)
      if (
        projectId &&
        db.prepare("SELECT id FROM projects WHERE id = ?").get(projectId)
      ) {
        db.prepare(
          "INSERT OR IGNORE INTO project_members (project_id, user_id) VALUES (?, ?)",
        ).run(projectId, newId)
        logActivity(
          projectId,
          admin,
          "project",
          `Invited ${name} to the project team`,
        )
      }
      json(ctx.res, 201, {
        member: db
          .prepare(
            "SELECT id, name, email, role, job_title, initials, color, phone FROM users WHERE id = ?",
          )
          .get(newId),
        // Returned once so the admin can hand it over; no email service is configured.
        temporaryPassword: tempPassword,
      })
    },
  },
  {
    method: "GET",
    path: "/api/suppliers",
    handler: async (ctx) => {
      requireUser(ctx)
      const db = getDb()
      const conditions: string[] = []
      const values: SqlParam[] = []
      const category = ctx.query.get("category")
      if (category && category !== "All") {
        conditions.push("category = ?")
        values.push(category)
      }
      const q = ctx.query.get("q")
      if (q) {
        conditions.push("(name LIKE ? OR category LIKE ?)")
        values.push(`%${q}%`, `%${q}%`)
      }
      const sql =
        "SELECT * FROM suppliers" +
        (conditions.length ? " WHERE " + conditions.join(" AND ") : "") +
        " ORDER BY name"
      json(ctx.res, 200, { suppliers: db.prepare(sql).all(...values) })
    },
  },
  {
    method: "POST",
    path: "/api/suppliers",
    handler: async (ctx) => {
      const user = requireUser(ctx)
      if (user.role === "viewer")
        throw new ApiError(403, "You do not have permission to add suppliers.")
      const body = (await readJsonBody(ctx.req)) as Record<string, unknown>
      const name = asString(body.name, "Supplier name", {
        required: true,
        max: 160,
      })
      const category = asString(body.category, "Category", {
        required: true,
        max: 80,
      })
      const contact = asString(body.contact, "Contact person", { max: 120 })
      const phone = asString(body.phone, "Phone", { max: 40 })
      const email = asString(body.email, "Email", { max: 200 })
      const location = asString(body.location, "Location", { max: 160 })
      const rating = asNumber(body.rating, "Rating", { min: 0 })
      if (rating > 5) throw new ApiError(400, "Rating must be between 0 and 5.")

      const db = getDb()
      const result = db
        .prepare(
          `INSERT INTO suppliers (name, category, contact, phone, email, location, rating) VALUES (?, ?, ?, ?, ?, ?, ?)`,
        )
        .run(name, category, contact, phone, email, location, rating)
      json(ctx.res, 201, {
        supplier: db
          .prepare("SELECT * FROM suppliers WHERE id = ?")
          .get(Number(result.lastInsertRowid)),
      })
    },
  },
  {
    method: "PATCH",
    path: "/api/me",
    handler: async (ctx) => {
      const user = requireUser(ctx)
      const body = (await readJsonBody(ctx.req)) as Record<string, unknown>
      const name = asString(body.name, "Name", { required: true, max: 120 })
      const email = asString(body.email, "Email", {
        required: true,
        max: 200,
      }).toLowerCase()
      const phone = asString(body.phone, "Phone", { max: 40 })
      const jobTitle = asString(body.jobTitle, "Job title", { max: 120 })
      if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email))
        throw new ApiError(400, "Enter a valid email address.")

      const db = getDb()
      const clash = db
        .prepare("SELECT id FROM users WHERE email = ? AND id != ?")
        .get(email, user.id)
      if (clash)
        throw new ApiError(409, "That email address is already in use.")

      db.prepare(
        "UPDATE users SET name = ?, email = ?, phone = ?, job_title = ?, initials = ?, updated_at = datetime('now') WHERE id = ?",
      ).run(name, email, phone, jobTitle, initialsFor(name), user.id)
      const updated = db
        .prepare(
          "SELECT id, email, name, role, job_title, initials, color, phone FROM users WHERE id = ?",
        )
        .get(user.id) as Record<string, unknown>
      json(ctx.res, 200, {
        user: {
          id: updated.id,
          email: updated.email,
          name: updated.name,
          role: updated.role,
          jobTitle: updated.job_title,
          initials: updated.initials,
          color: updated.color,
          phone: updated.phone,
        },
      })
    },
  },
  {
    method: "POST",
    path: "/api/me/password",
    handler: async (ctx) => {
      const user = requireUser(ctx)
      const body = (await readJsonBody(ctx.req)) as Record<string, unknown>
      const currentPassword = asString(
        body.currentPassword,
        "Current password",
        { required: true },
      )
      const newPassword = asString(body.newPassword, "New password", {
        required: true,
        max: 200,
      })
      if (newPassword.length < 8)
        throw new ApiError(400, "New password must be at least 8 characters.")
      if (!/[a-zA-Z]/.test(newPassword) || !/[0-9]/.test(newPassword)) {
        throw new ApiError(
          400,
          "New password must include at least one letter and one number.",
        )
      }

      const db = getDb()
      const row = db
        .prepare("SELECT password_hash FROM users WHERE id = ?")
        .get(user.id) as { password_hash: string }
      if (!verifyPassword(currentPassword, row.password_hash)) {
        throw new ApiError(400, "Your current password is incorrect.")
      }
      db.prepare(
        "UPDATE users SET password_hash = ?, updated_at = datetime('now') WHERE id = ?",
      ).run(hashPassword(newPassword), user.id)
      // Invalidate every other session for this account.
      db.prepare("DELETE FROM sessions WHERE user_id = ?").run(user.id)
      ctx.res.setHeader(
        "Set-Cookie",
        "vecai_session=; Path=/; HttpOnly; SameSite=Lax; Max-Age=0",
      )
      json(ctx.res, 200, { ok: true })
    },
  },
  {
    method: "GET",
    path: "/api/me/preferences",
    handler: async (ctx) => {
      const user = requireUser(ctx)
      const db = getDb()
      const row = db
        .prepare("SELECT prefs FROM user_preferences WHERE user_id = ?")
        .get(user.id) as { prefs: string } | undefined
      const prefs = row
        ? {
            ...DEFAULT_PREFS,
            ...JSON.parse(row.prefs) as Record<string, boolean>,
          }
        : { ...DEFAULT_PREFS }
      json(ctx.res, 200, { preferences: prefs })
    },
  },
  {
    method: "PATCH",
    path: "/api/me/preferences",
    handler: async (ctx) => {
      const user = requireUser(ctx)
      const body = (await readJsonBody(ctx.req)) as Record<string, unknown>
      const db = getDb()
      const row = db
        .prepare("SELECT prefs FROM user_preferences WHERE user_id = ?")
        .get(user.id) as { prefs: string } | undefined
      const current = row
        ? JSON.parse(row.prefs) as Record<string, boolean>
        : {}
      const next: Record<string, boolean> = { ...current }
      for (const key of Object.keys(DEFAULT_PREFS)) {
        if (typeof body[key] === "boolean") next[key] = (body[key] as boolean)
      }
      db.prepare(
        `INSERT INTO user_preferences (user_id, prefs) VALUES (?, ?)
         ON CONFLICT (user_id) DO UPDATE SET prefs = excluded.prefs`,
      ).run(user.id, JSON.stringify(next))
      json(ctx.res, 200, { preferences: { ...DEFAULT_PREFS, ...next } })
    },
  },
  {
    method: "GET",
    path: "/api/organisation",
    handler: async (ctx) => {
      requireUser(ctx)
      const row = getDb()
        .prepare(
          "SELECT name, business_type, country, currency FROM organisation WHERE id = 1",
        )
        .get() as Record<string, unknown>
      json(ctx.res, 200, {
        organisation: {
          name: row.name,
          businessType: row.business_type,
          country: row.country,
          currency: row.currency,
        },
      })
    },
  },
  {
    method: "PATCH",
    path: "/api/organisation",
    handler: async (ctx) => {
      requireRole(ctx, "admin")
      const body = (await readJsonBody(ctx.req)) as Record<string, unknown>
      const name = asString(body.name, "Organisation name", {
        required: true,
        max: 160,
      })
      const businessType = asString(body.businessType, "Business type", {
        required: true,
        max: 80,
      })
      const country = asString(body.country, "Country", {
        required: true,
        max: 80,
      })
      const currency = asString(body.currency, "Currency", {
        required: true,
        max: 80,
      })
      getDb()
        .prepare(
          "UPDATE organisation SET name = ?, business_type = ?, country = ?, currency = ? WHERE id = 1",
        )
        .run(name, businessType, country, currency)
      json(ctx.res, 200, {
        organisation: { name, businessType, country, currency },
      })
    },
  },
]
