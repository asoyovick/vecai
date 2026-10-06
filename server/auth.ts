import crypto from "node:crypto"
import type { DatabaseSync } from "node:sqlite"
import { getDb } from "./db/db"

export const SESSION_COOKIE = "vecai_session"
const SESSION_TTL_DAYS = 7

export interface UserRow {
  id: number
  email: string
  name: string
  role: "admin" | "member" | "viewer"
  job_title: string
  initials: string
  color: string
  phone: string
}

export function hashPassword(password: string): string {
  const salt = crypto.randomBytes(16).toString("hex")
  const hash = crypto.scryptSync(password, salt, 64).toString("hex")
  return `scrypt:${salt}:${hash}`
}

export function verifyPassword(password: string, stored: string): boolean {
  const [scheme, salt, hash] = stored.split(":")
  if (scheme !== "scrypt" || !salt || !hash) return false
  const candidate = crypto.scryptSync(password, salt, 64)
  const expected = Buffer.from(hash, "hex")
  return (
    candidate.length === expected.length &&
    crypto.timingSafeEqual(candidate, expected)
  )
}

export function createSession(
  userId: number,
): { token: string; expiresAt: Date } {
  const db = getDb()
  const token = crypto.randomBytes(32).toString("hex")
  const tokenHash = crypto.createHash("sha256").update(token).digest("hex")
  const expiresAt = new Date(
    Date.now() + SESSION_TTL_DAYS * 24 * 60 * 60 * 1000,
  )
  db.prepare(
    "INSERT INTO sessions (token_hash, user_id, expires_at) VALUES (?, ?, ?)",
  ).run(tokenHash, userId, expiresAt.toISOString())
  return { token, expiresAt }
}

export function destroySession(token: string): void {
  const db = getDb()
  const tokenHash = crypto.createHash("sha256").update(token).digest("hex")
  db.prepare("DELETE FROM sessions WHERE token_hash = ?").run(tokenHash)
}

export function parseCookies(
  header: string | undefined,
): Record<string, string> {
  const out: Record<string, string> = {}
  if (!header) return out
  for (const part of header.split(";")) {
    const idx = part.indexOf("=")
    if (idx === -1) continue
    out[part.slice(0, idx).trim()] = decodeURIComponent(
      part.slice(idx + 1).trim(),
    )
  }
  return out
}

export function getSessionUser(
  req: import("http").IncomingMessage,
): UserRow | null {
  const db = getDb()
  const cookies = parseCookies(req.headers.cookie)
  const token = cookies[SESSION_COOKIE]
  if (!token) return null
  const tokenHash = crypto.createHash("sha256").update(token).digest("hex")
  const row = db
    .prepare(
      `SELECT u.id, u.email, u.name, u.role, u.job_title, u.initials, u.color, u.phone
       FROM sessions s JOIN users u ON u.id = s.user_id
       WHERE s.token_hash = ? AND s.expires_at > datetime('now')`,
    )
    .get(tokenHash) as UserRow | undefined
  return row ?? null
}

export function sessionCookieHeader(token: string, expiresAt: Date): string {
  return `${SESSION_COOKIE}=${token}; Path=/; HttpOnly; SameSite=Lax; Expires=${expiresAt.toUTCString()}`
}

export function clearSessionCookieHeader(): string {
  return `${SESSION_COOKIE}=; Path=/; HttpOnly; SameSite=Lax; Max-Age=0`
}
