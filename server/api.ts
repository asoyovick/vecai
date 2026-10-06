import type { IncomingMessage, ServerResponse } from "node:http"
import type { Plugin } from "vite"
import { ApiError, json, type Ctx } from "./http"
import { getSessionUser } from "./auth"
import { seedIfEmpty } from "./db/seed"
import type { RouteDef } from "./router"
import { authRoutes } from "./routes/auth"
import { projectRoutes } from "./routes/projects"
import { taskRoutes } from "./routes/tasks"
import { materialRoutes } from "./routes/materials"
import { documentRoutes } from "./routes/documents"
import { siteUpdateRoutes } from "./routes/siteUpdates"
import { workspaceRoutes } from "./routes/workspace"
import { assistantRoutes } from "./routes/assistant"

interface CompiledRoute extends RouteDef {
  keys: string[]
  regex: RegExp
}

function compile(routes: RouteDef[]): CompiledRoute[] {
  return routes.map((route) => {
    const keys: string[] = []
    const pattern = route.path
      .split("/")
      .map((segment) => {
        if (segment.startsWith(":")) {
          keys.push(segment.slice(1))
          return "([^/]+)"
        }
        return segment.replace(/[.*+?^${}()|[\]\\]/g, "\\$&")
      })
      .join("/")
    return { ...route, keys, regex: new RegExp(`^${pattern}$`) }
  })
}

const allRoutes = [
  ...authRoutes,
  ...projectRoutes,
  ...taskRoutes,
  ...materialRoutes,
  ...documentRoutes,
  ...siteUpdateRoutes,
  ...workspaceRoutes,
  ...assistantRoutes,
]
const compiled = compile(allRoutes)

async function handle(
  req: IncomingMessage,
  res: ServerResponse,
): Promise<void> {
  const url = new URL(req.url || "/", "http://local")
  const pathname = url.pathname

  for (const route of compiled) {
    if (route.method !== req.method) continue
    const match = route.regex.exec(pathname)
    if (!match) continue

    const params: Record<string, string> = {}
    route.keys.forEach((key, i) => {
      params[key] = decodeURIComponent(match[i + 1] ?? "")
    })

    const ctx: Ctx = {
      req,
      res,
      params,
      query: url.searchParams,
      user: getSessionUser(req),
    }
    try {
      await route.handler(ctx)
    } catch (err) {
      if (err instanceof ApiError) {
        json(res, err.status, { error: err.message })
      } else {
        console.error("[vecai-api] unhandled error:", err)
        json(res, 500, {
          error: "Something went wrong on the server. Please try again.",
        })
      }
    }
    return
  }

  json(res, 404, { error: "API endpoint not found." })
}

/** Serves the VECAI REST API on the Vite dev/preview server (same origin, no proxy needed). */
export function vecaiApi(): Plugin {
  const middleware = (
    req: IncomingMessage,
    res: ServerResponse,
    next: () => void,
  ) => {
    if (!req.url || !req.url.startsWith("/api/")) return next()
    // Mutations must come from our own frontend (JSON content type), a lightweight CSRF guard.
    const method = req.method || "GET"
    if (method !== "GET" && method !== "HEAD") {
      const contentType = String(req.headers["content-type"] || "")
      const isJson = contentType.includes("application/json")
      const isMultipart = contentType.includes("multipart/form-data")
      // A bodyless POST (e.g. logout) is still same-origin JSON semantics.
      const bodyless = !contentType && !req.headers["content-length"]
      if (!isJson && !isMultipart && !bodyless) {
        json(res, 415, { error: "Unsupported content type." })
        return
      }
    }
    handle(req, res).catch((err) => {
      console.error("[vecai-api] fatal:", err)
      if (!res.headersSent) json(res, 500, { error: "Internal server error." })
    })
  }

  return {
    name: "vecai-api",
    configureServer(server) {
      seedIfEmpty()
      server.middlewares.use(middleware)
    },
    configurePreviewServer(server) {
      seedIfEmpty()
      server.middlewares.use(middleware)
    },
  }
}
