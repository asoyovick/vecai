import type { IncomingMessage, ServerResponse } from "node:http"
import { getDb } from "./db/db"
import { getSessionUser, type UserRow } from "./auth"

export type SqlParam = string | number | bigint | null | Uint8Array

export class ApiError extends Error {
  status: number
  constructor(status: number, message: string) {
    super(message)
    this.status = status
  }
}

export function json(res: ServerResponse, status: number, data: unknown): void {
  const body = JSON.stringify(data)
  res.writeHead(status, {
    "Content-Type": "application/json; charset=utf-8",
    "Content-Length": Buffer.byteLength(body),
    "Cache-Control": "no-store",
  })
  res.end(body)
}

export function readJsonBody(
  req: IncomingMessage,
  limitBytes = 1024 * 1024,
): Promise<unknown> {
  return new Promise((resolve, reject) => {
    const length = Number(req.headers["content-length"] || 0)
    if (length > limitBytes) {
      reject(new ApiError(413, "Request body too large."))
      req.resume()
      return
    }
    const chunks: Buffer[] = []
    let size = 0
    req.on("data", (chunk: Buffer) => {
      size += chunk.length
      if (size > limitBytes) {
        reject(new ApiError(413, "Request body too large."))
        req.destroy()
        return
      }
      chunks.push(chunk)
    })
    req.on("end", () => {
      if (chunks.length === 0) return resolve({})
      try {
        resolve(JSON.parse(Buffer.concat(chunks).toString("utf8")))
      } catch {
        reject(new ApiError(400, "Invalid JSON body."))
      }
    })
    req.on("error", reject)
  })
}

export async function readMultipart(req: IncomingMessage): Promise<FormData> {
  const length = Number(req.headers["content-length"] || 0)
  if (length > 25 * 1024 * 1024)
    throw new ApiError(413, "File exceeds the 25 MB limit.")
  const request = new Request("http://local/upload", {
    method: "POST",
    headers: { "content-type": req.headers["content-type"] || "" },
    body: req as unknown as ReadableStream,
    duplex: "half",
  } as RequestInit)
  try {
    return await request.formData()
  } catch {
    throw new ApiError(400, "Could not parse the uploaded file.")
  }
}

export interface Ctx {
  req: IncomingMessage
  res: ServerResponse
  params: Record<string, string>
  query: URLSearchParams
  user: UserRow | null
}

export function requireUser(ctx: Ctx): UserRow {
  if (!ctx.user) throw new ApiError(401, "You must be signed in.")
  return ctx.user
}

export function requireRole(ctx: Ctx, ...roles: UserRow["role"][]): UserRow {
  const user = requireUser(ctx)
  if (roles.length && !roles.includes(user.role)) {
    throw new ApiError(
      403,
      "You do not have permission to perform this action.",
    )
  }
  return user
}

// ── Validation helpers ────────────────────────────────────────────────────────

export function asString(
  value: unknown,
  field: string,
  opts: { required?: boolean; max?: number } = {},
): string {
  if (value === undefined || value === null) {
    if (opts.required) throw new ApiError(400, `${field} is required.`)
    return ""
  }
  if (typeof value !== "string")
    throw new ApiError(400, `${field} must be a string.`)
  const trimmed = value.trim()
  if (opts.required && !trimmed)
    throw new ApiError(400, `${field} is required.`)
  if (opts.max && trimmed.length > opts.max)
    throw new ApiError(400, `${field} must be at most ${opts.max} characters.`)
  return trimmed
}

export function asNumber(
  value: unknown,
  field: string,
  opts: { required?: boolean; min?: number; integer?: boolean } = {},
): number {
  if (value === undefined || value === null || value === "") {
    if (opts.required) throw new ApiError(400, `${field} is required.`)
    return 0
  }
  const n = Number(value)
  if (!Number.isFinite(n)) throw new ApiError(400, `${field} must be a number.`)
  if (opts.integer && !Number.isInteger(n))
    throw new ApiError(400, `${field} must be a whole number.`)
  if (opts.min !== undefined && n < opts.min)
    throw new ApiError(400, `${field} must be at least ${opts.min}.`)
  return n
}

export function asOneOf<T extends string>(
  value: unknown,
  field: string,
  allowed: readonly T[],
): T {
  if (typeof value !== "string" || !allowed.includes(value as T)) {
    throw new ApiError(400, `${field} must be one of: ${allowed.join(", ")}.`)
  }
  return value as T
}

export function asDate(
  value: unknown,
  field: string,
  opts: { required?: boolean } = {},
): string | null {
  if (value === undefined || value === null || value === "") {
    if (opts.required) throw new ApiError(400, `${field} is required.`)
    return null
  }
  if (typeof value !== "string" || !/^\d{4}-\d{2}-\d{2}$/.test(value)) {
    throw new ApiError(400, `${field} must be a date (YYYY-MM-DD).`)
  }
  return value
}

export function logActivity(
  projectId: number | null,
  actor: { id: number; name: string } | null,
  type: "task" | "document" | "material" | "site-update" | "project" | "comment" | "system",
  text: string,
): void {
  getDb()
    .prepare(
      "INSERT INTO activity_log (project_id, actor_id, actor_name, type, text) VALUES (?, ?, ?, ?, ?)",
    )
    .run(projectId, actor?.id ?? null, actor?.name ?? "System", type, text)
}

export function touchProject(projectId: number): void {
  getDb()
    .prepare("UPDATE projects SET updated_at = datetime('now') WHERE id = ?")
    .run(projectId)
}
