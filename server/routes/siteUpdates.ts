import type { RouteDef } from "../router"
import {
  ApiError,
  asDate,
  asNumber,
  asString,
  json,
  logActivity,
  readJsonBody,
  requireUser,
  type SqlParam,
} from "../http"
import { getDb } from "../db/db"

interface MaterialEntry {
  name?: unknown
  quantity?: unknown
  unit?: unknown
}

function parseMaterials(
  raw: unknown,
): Array<{ name: string; quantity: number; unit: string }> {
  if (!Array.isArray(raw)) return []
  const out: Array<{ name: string; quantity: number; unit: string }> = []
  for (const entry of raw as MaterialEntry[]) {
    if (!entry || typeof entry !== "object") continue
    const name = String(entry.name ?? "").trim()
    const quantity = Number(entry.quantity)
    const unit = String(entry.unit ?? "")
      .trim()
      .slice(0, 24)
    if (!name) continue
    if (!Number.isFinite(quantity) || quantity <= 0) {
      throw new ApiError(
        400,
        `Enter a quantity greater than zero for “${name}”.`,
      )
    }
    out.push({ name: name.slice(0, 120), quantity, unit })
  }
  return out
}

export const siteUpdateRoutes: RouteDef[] = [
  {
    method: "GET",
    path: "/api/site-updates",
    handler: async (ctx) => {
      requireUser(ctx)
      const db = getDb()
      const conditions: string[] = []
      const values: SqlParam[] = []
      const projectId = ctx.query.get("projectId")
      if (projectId && projectId !== "All") {
        conditions.push("su.project_id = ?")
        values.push(Number(projectId))
      }
      const sql = `
        SELECT su.*, p.name AS project_name, u.name AS submitted_by_name
        FROM site_updates su
        JOIN projects p ON p.id = su.project_id
        LEFT JOIN users u ON u.id = su.submitted_by
        ${conditions.length ? "WHERE " + conditions.join(" AND ") : ""}
        ORDER BY su.date DESC, su.id DESC LIMIT 50`
      const rows = db
        .prepare(sql)
        .all(...values) as Array<Record<string, unknown>>
      const materialsStmt = db.prepare(
        "SELECT name, quantity, unit FROM site_update_materials WHERE site_update_id = ?",
      )
      json(ctx.res, 200, {
        siteUpdates: rows.map((r) => ({
          ...r,
          materialsUsed: materialsStmt.all(Number(r.id)),
        })),
      })
    },
  },
  {
    method: "POST",
    path: "/api/site-updates",
    handler: async (ctx) => {
      const user = requireUser(ctx)
      if (user.role === "viewer")
        throw new ApiError(
          403,
          "You do not have permission to submit site updates.",
        )
      const body = (await readJsonBody(ctx.req)) as Record<string, unknown>

      const projectId = Number(body.projectId)
      const date = asDate(body.date, "Date", { required: true }) as string
      const weather = asString(body.weather, "Weather", { max: 60 })
      const temperature = asString(body.temperature, "Temperature", { max: 24 })
      const workers = asNumber(body.workersOnSite, "Workers on site", {
        min: 0,
        integer: true,
      })
      const workCompleted = asString(body.workCompleted, "Work completed", {
        required: true,
        max: 5000,
      })
      const issues = asString(body.issues, "Issues", { max: 3000 })
      const notes = asString(body.notes, "Notes", { max: 3000 })
      const materials = parseMaterials(body.materialsUsed)

      const db = getDb()
      const project = db
        .prepare("SELECT id, name FROM projects WHERE id = ?")
        .get(projectId) as { id: number; name: string } | undefined
      if (!project) throw new ApiError(400, "Select a valid project.")

      const result = db
        .prepare(
          `INSERT INTO site_updates (project_id, submitted_by, date, weather, temperature, workers_on_site, work_completed, issues, notes, photo_count)
         VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, 0)`,
        )
        .run(
          projectId,
          user.id,
          date,
          weather,
          temperature,
          workers,
          workCompleted,
          issues,
          notes,
        )
      const updateId = Number(result.lastInsertRowid)

      const insertMaterial = db.prepare(
        "INSERT INTO site_update_materials (site_update_id, name, quantity, unit) VALUES (?, ?, ?, ?)",
      )
      for (const m of materials)
        insertMaterial.run(updateId, m.name, m.quantity, m.unit)

      logActivity(
        projectId,
        user,
        "site-update",
        `Submitted site update for ${project.name} — ${date}`,
      )

      const row = db
        .prepare(
          `SELECT su.*, p.name AS project_name, u.name AS submitted_by_name
         FROM site_updates su JOIN projects p ON p.id = su.project_id
         LEFT JOIN users u ON u.id = su.submitted_by WHERE su.id = ?`,
        )
        .get(updateId) as Record<string, unknown>
      json(ctx.res, 201, { siteUpdate: { ...row, materialsUsed: materials } })
    },
  },
]
