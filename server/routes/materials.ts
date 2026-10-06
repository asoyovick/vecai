import type { RouteDef } from "../router"
import {
  ApiError,
  asNumber,
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

const MATERIAL_SELECT = `
  SELECT m.*, s.name AS supplier_name, p.name AS project_name
  FROM materials m
  LEFT JOIN suppliers s ON s.id = m.supplier_id
  JOIN projects p ON p.id = m.project_id`

/** Recompute stock status after a delivery raises the on-hand quantity. */
function recomputeStatusAfterDelivery(
  available: number,
  required: number,
  current: string,
): string {
  if (required <= 0) return current === "ordered" ? "ok" : current
  const ratio = available / required
  if (current === "ordered") return "ok"
  if (ratio >= 0.5) return "ok"
  return current
}

export const materialRoutes: RouteDef[] = [
  {
    method: "GET",
    path: "/api/materials",
    handler: async (ctx) => {
      requireUser(ctx)
      const db = getDb()
      const conditions: string[] = []
      const values: SqlParam[] = []
      const projectId = ctx.query.get("projectId")
      if (projectId) {
        conditions.push("m.project_id = ?")
        values.push(Number(projectId))
      }
      const status = ctx.query.get("status")
      if (status && status !== "All") {
        conditions.push("m.status = ?")
        values.push(status)
      }
      const q = ctx.query.get("q")
      if (q) {
        conditions.push("m.name LIKE ?")
        values.push(`%${q}%`)
      }
      const sql =
        MATERIAL_SELECT +
        (conditions.length ? " WHERE " + conditions.join(" AND ") : "") +
        " ORDER BY m.name"
      json(ctx.res, 200, { materials: db.prepare(sql).all(...values) })
    },
  },
  {
    method: "POST",
    path: "/api/materials",
    handler: async (ctx) => {
      const user = requireUser(ctx)
      if (user.role === "viewer")
        throw new ApiError(403, "You do not have permission to add materials.")
      const body = (await readJsonBody(ctx.req)) as Record<string, unknown>
      const projectId = Number(body.projectId)
      const name = asString(body.name, "Material name", {
        required: true,
        max: 160,
      })
      const unit = asString(body.unit, "Unit", { required: true, max: 24 })
      const requiredQty = asNumber(body.requiredQty, "Required quantity", {
        min: 0,
      })
      const availableQty = asNumber(body.availableQty, "Available quantity", {
        min: 0,
      })
      const supplierId = body.supplierId ? Number(body.supplierId) : null
      const unitCost = asNumber(body.unitCost, "Unit cost", {
        min: 0,
        integer: true,
      })
      const status = asOneOf(body.status ?? "ok", "Status", [
        "ok",
        "low",
        "critical",
        "ordered",
      ] as const)

      const db = getDb()
      if (!db.prepare("SELECT id FROM projects WHERE id = ?").get(projectId)) {
        throw new ApiError(400, "Select a valid project.")
      }
      if (
        supplierId !== null &&
        !db.prepare("SELECT id FROM suppliers WHERE id = ?").get(supplierId)
      ) {
        throw new ApiError(400, "Select a valid supplier.")
      }

      const result = db
        .prepare(
          `INSERT INTO materials (project_id, supplier_id, name, unit, required_qty, available_qty, used_qty, status, unit_cost)
         VALUES (?, ?, ?, ?, ?, ?, 0, ?, ?)`,
        )
        .run(
          projectId,
          supplierId,
          name,
          unit,
          requiredQty,
          availableQty,
          status,
          unitCost,
        )
      logActivity(
        projectId,
        user,
        "material",
        `Added material “${name}” to tracking`,
      )
      touchProject(projectId)
      const row = db
        .prepare(MATERIAL_SELECT + " WHERE m.id = ?")
        .get(Number(result.lastInsertRowid))
      json(ctx.res, 201, { material: row })
    },
  },
  {
    method: "POST",
    path: "/api/materials/:id/deliveries",
    handler: async (ctx) => {
      const user = requireUser(ctx)
      if (user.role === "viewer")
        throw new ApiError(403, "You do not have permission to log deliveries.")
      const body = (await readJsonBody(ctx.req)) as Record<string, unknown>
      const quantity = asNumber(body.quantity, "Quantity", { required: true })
      if (quantity <= 0)
        throw new ApiError(400, "Quantity must be greater than zero.")
      const note = asString(body.note, "Note", { max: 500 })

      const db = getDb()
      const id = Number(ctx.params.id)
      const material = db
        .prepare("SELECT * FROM materials WHERE id = ?")
        .get(id) as Record<string, unknown> | undefined
      if (!material) throw new ApiError(404, "Material not found.")

      const today = new Date().toISOString().slice(0, 10)
      db.prepare(
        "INSERT INTO material_deliveries (material_id, quantity, note, recorded_by) VALUES (?, ?, ?, ?)",
      ).run(id, quantity, note, user.id)
      const newAvailable = material.available_qty as number + quantity
      const newStatus = recomputeStatusAfterDelivery(
        newAvailable,
        material.required_qty as number,
        material.status as string,
      )
      db.prepare(
        "UPDATE materials SET available_qty = ?, status = ?, last_delivery = ?, updated_at = datetime('now') WHERE id = ?",
      ).run(newAvailable, newStatus, today, id)

      logActivity(
        material.project_id as number,
        user,
        "material",
        `Logged delivery: ${quantity} ${material.unit} ${material.name}`,
      )
      touchProject(material.project_id as number)
      const row = db.prepare(MATERIAL_SELECT + " WHERE m.id = ?").get(id)
      json(ctx.res, 201, { material: row })
    },
  },
]
