import type { RouteDef } from "../router"
import { ApiError, asString, json, readJsonBody } from "../http"
import {
  clearSessionCookieHeader,
  createSession,
  destroySession,
  parseCookies,
  SESSION_COOKIE,
  sessionCookieHeader,
  verifyPassword,
} from "../auth"
import { getDb } from "../db/db"

export const authRoutes: RouteDef[] = [
  {
    method: "POST",
    path: "/api/auth/login",
    handler: async (ctx) => {
      const body = (await readJsonBody(ctx.req)) as Record<string, unknown>
      const email = asString(body.email, "Email", { required: true, max: 200 })
      const password = asString(body.password, "Password", {
        required: true,
        max: 200,
      })

      const db = getDb()
      const user = db
        .prepare(
          "SELECT id, email, password_hash, name, role, job_title, initials, color, phone FROM users WHERE email = ?",
        )
        .get(email) as Record<string, unknown> & { id: number } | undefined

      if (!user || !verifyPassword(password, String(user.password_hash))) {
        // Uniform message: do not reveal whether the account exists.
        throw new ApiError(401, "Incorrect email or password.")
      }

      const { token, expiresAt } = createSession(user.id)
      ctx.res.setHeader("Set-Cookie", sessionCookieHeader(token, expiresAt))
      json(ctx.res, 200, {
        user: {
          id: user.id,
          email: user.email,
          name: user.name,
          role: user.role,
          jobTitle: user.job_title,
          initials: user.initials,
          color: user.color,
          phone: user.phone,
        },
      })
    },
  },
  {
    method: "POST",
    path: "/api/auth/logout",
    handler: async (ctx) => {
      const token = parseCookies(ctx.req.headers.cookie)[SESSION_COOKIE]
      if (token) destroySession(token)
      ctx.res.setHeader("Set-Cookie", clearSessionCookieHeader())
      json(ctx.res, 200, { ok: true })
    },
  },
  {
    method: "GET",
    path: "/api/auth/me",
    handler: async (ctx) => {
      if (!ctx.user) throw new ApiError(401, "Not signed in.")
      const { id, email, name, role, job_title, initials, color, phone } =
        ctx.user
      json(ctx.res, 200, {
        user: {
          id,
          email,
          name,
          role,
          jobTitle: job_title,
          initials,
          color,
          phone,
        },
      })
    },
  },
]
