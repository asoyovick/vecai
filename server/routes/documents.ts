import crypto from "node:crypto"
import fs from "node:fs"
import path from "node:path"
import type { RouteDef } from "../router"
import {
  ApiError,
  asOneOf,
  asString,
  json,
  logActivity,
  readMultipart,
  requireUser,
  touchProject,
  type SqlParam,
} from "../http"
import { UPLOADS_DIR } from "../db/db"
import { getDb } from "../db/db"

const CATEGORIES = [
  "Architectural Drawing",
  "Structural Drawing",
  "BOQ",
  "Contract",
  "Permit",
  "Invoice",
  "Site Report",
  "Inspection Report",
  "Photo",
] as const

const DOC_SELECT = `
  SELECT d.*, p.name AS project_name, u.name AS uploaded_by_name
  FROM documents d
  JOIN projects p ON p.id = d.project_id
  LEFT JOIN users u ON u.id = d.uploaded_by`

const ALLOWED_EXTENSIONS = new Set([
  "pdf",
  "xls",
  "xlsx",
  "csv",
  "doc",
  "docx",
  "png",
  "jpg",
  "jpeg",
  "webp",
  "zip",
  "dwg",
  "ifc",
  "txt",
])

function sanitizeFilename(name: string): { base: string; ext: string } {
  const ext = (path.extname(name).slice(1) || "").toLowerCase()
  const base =
    path
      .basename(name, path.extname(name))
      .replace(/[^a-zA-Z0-9-_ ]/g, "")
      .trim()
      .slice(0, 80) || "file"
  return { base, ext }
}

export const documentRoutes: RouteDef[] = [
  {
    method: "GET",
    path: "/api/documents",
    handler: async (ctx) => {
      requireUser(ctx)
      const db = getDb()
      const conditions: string[] = []
      const values: SqlParam[] = []
      const projectId = ctx.query.get("projectId")
      if (projectId && projectId !== "All") {
        conditions.push("d.project_id = ?")
        values.push(Number(projectId))
      }
      const category = ctx.query.get("category")
      if (category && category !== "All") {
        conditions.push("d.category = ?")
        values.push(category)
      }
      const q = ctx.query.get("q")
      if (q) {
        conditions.push("d.name LIKE ?")
        values.push(`%${q}%`)
      }
      const sql =
        DOC_SELECT +
        (conditions.length ? " WHERE " + conditions.join(" AND ") : "") +
        " ORDER BY d.created_at DESC"
      json(ctx.res, 200, { documents: db.prepare(sql).all(...values) })
    },
  },
  {
    method: "POST",
    path: "/api/documents",
    handler: async (ctx) => {
      const user = requireUser(ctx)
      if (user.role === "viewer")
        throw new ApiError(
          403,
          "You do not have permission to upload documents.",
        )
      const form = await readMultipart(ctx.req)

      const file = form.get("file")
      const projectId = Number(form.get("projectId"))
      const name = String(form.get("name") ?? "").trim()
      const category = String(form.get("category") ?? "")
      const version = String(form.get("version") ?? "1.0").trim() || "1.0"

      if (!(file instanceof File))
        throw new ApiError(400, "Select a file to upload.")
      if (!name) throw new ApiError(400, "Document name is required.")
      const validCategory = asOneOf(category, "Category", CATEGORIES)
      if (file.size === 0)
        throw new ApiError(400, "The selected file is empty.")
      if (file.size > 25 * 1024 * 1024)
        throw new ApiError(400, "Files are limited to 25 MB.")

      const { base, ext } = sanitizeFilename(file.name)
      if (ext && !ALLOWED_EXTENSIONS.has(ext)) {
        throw new ApiError(400, `File type “.${ext}” is not allowed.`)
      }

      const db = getDb()
      const project = db
        .prepare("SELECT id FROM projects WHERE id = ?")
        .get(projectId)
      if (!project) throw new ApiError(400, "Select a valid project.")

      const storedName = `${Date.now()}-${crypto.randomBytes(6).toString("hex")}${
        ext ? "." + ext : ""
      }`
      const buffer = Buffer.from(await file.arrayBuffer())
      fs.writeFileSync(path.join(UPLOADS_DIR, storedName), buffer)

      const result = db
        .prepare(
          `INSERT INTO documents (project_id, uploaded_by, name, category, version, original_name, stored_name, mime_type, size_bytes, format)
         VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
        )
        .run(
          projectId,
          user.id,
          name,
          validCategory,
          version,
          file.name,
          storedName,
          file.type || "application/octet-stream",
          file.size,
          ext.toUpperCase(),
        )

      logActivity(projectId, user, "document", `Uploaded “${name}”`)
      touchProject(projectId)
      const row = db
        .prepare(DOC_SELECT + " WHERE d.id = ?")
        .get(Number(result.lastInsertRowid))
      json(ctx.res, 201, { document: row })
    },
  },
  {
    method: "GET",
    path: "/api/documents/:id/download",
    handler: async (ctx) => {
      requireUser(ctx)
      const db = getDb()
      const doc = db
        .prepare("SELECT * FROM documents WHERE id = ?")
        .get(Number(ctx.params.id)) as Record<string, unknown> | undefined
      if (!doc) throw new ApiError(404, "Document not found.")

      const storedName = String(doc.stored_name)
      if (storedName.startsWith("seeded/")) {
        // Seeded metadata records have no binary file behind them.
        throw new ApiError(
          404,
          "This document was imported as a record only — no file is attached.",
        )
      }
      const filePath = path.join(UPLOADS_DIR, path.basename(storedName))
      if (!fs.existsSync(filePath))
        throw new ApiError(404, "The file is no longer available.")

      ctx.res.writeHead(200, {
        "Content-Type": String(doc.mime_type) || "application/octet-stream",
        "Content-Length": fs.statSync(filePath).size,
        "Content-Disposition": `attachment; filename="${String(doc.original_name).replace(/[^\w.\- ]/g, "_")}"`,
      })
      fs.createReadStream(filePath).pipe(ctx.res)
    },
  },
]
