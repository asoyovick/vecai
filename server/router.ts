import type { Ctx } from "./http"

export interface RouteDef {
  method: string
  path: string
  handler: (ctx: Ctx) => Promise<void> | void
}
