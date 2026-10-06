import type { RouteDef } from "../router"
import { asNumber, asString, json, readJsonBody, requireUser } from "../http"
import { getDb } from "../db/db"

export type Tone = "default" | "danger" | "warn" | "success"
export interface AssistantLine {
  text: string
  tone?: Tone
}
export interface AssistantSection {
  heading?: string
  lines: AssistantLine[]
}
export interface AssistantAnswer {
  title: string
  intro: string
  sections: AssistantSection[]
  footnote: string
}

interface MaterialRow {
  id: number
  name: string
  unit: string
  required_qty: number
  available_qty: number
  used_qty: number
  status: string
  project_name: string
  start_date: string | null
}

function daysBetween(from: string, to: Date): number {
  return Math.max(
    1,
    Math.round(
      (to.getTime() - new Date(from + "T00:00:00Z").getTime()) / 86400000,
    ),
  )
}

function materialDaysLeft(m: MaterialRow, now: Date): number | null {
  if (!m.start_date || m.used_qty <= 0) return null
  const rate = m.used_qty / daysBetween(m.start_date, now)
  return rate > 0 ? Math.floor(m.available_qty / rate) : null
}

function fmtQty(n: number): string {
  return Number.isInteger(n)
    ? n.toLocaleString("en-KE")
    : n.toFixed(n < 1 ? 2 : 1)
}

function getProject(db: ReturnType<typeof getDb>, id: number | null) {
  if (!id) return null
  return (
    db
      .prepare(
        "SELECT id, name, status, progress, start_date, expected_completion FROM projects WHERE id = ?",
      )
      .get(id) as {
      id: number
      name: string
      status: string
      progress: number
      start_date: string | null
      expected_completion: string | null
    } | undefined ?? null
  )
}

function materialsForecast(
  db: ReturnType<typeof getDb>,
  projectId: number | null,
  now: Date,
): AssistantAnswer {
  const where = projectId ? "WHERE m.project_id = ?" : ""
  const rows = db
    .prepare(
      `SELECT m.*, p.name AS project_name, p.start_date FROM materials m JOIN projects p ON p.id = m.project_id
     ${where} ORDER BY CASE m.status WHEN 'critical' THEN 0 WHEN 'low' THEN 1 WHEN 'ordered' THEN 2 ELSE 3 END, m.name`,
    )
    .all(...(projectId ? [projectId] : [])) as unknown as MaterialRow[]

  const critical = rows.filter((m) => m.status === "critical")
  const low = rows.filter((m) => m.status === "low")
  const ordered = rows.filter((m) => m.status === "ordered")

  const sections: AssistantSection[] = []
  if (critical.length) {
    sections.push({
      heading: "Critical — reorder immediately",
      lines: critical.map((m) => {
        const days = materialDaysLeft(m, now)
        return {
          text: `${m.name} (${m.project_name}) — ${fmtQty(m.available_qty)} ${m.unit} remaining${
            days !== null
              ? `, roughly ${days} day${
                  days === 1 ? "" : "s"
                } of supply at the current usage rate`
              : ""
          }.`,
          tone: "danger" as Tone,
        }
      }),
    })
  }
  if (low.length) {
    sections.push({
      heading: "Low stock — order within the week",
      lines: low.map((m) => {
        const days = materialDaysLeft(m, now)
        return {
          text: `${m.name} (${m.project_name}) — ${fmtQty(m.available_qty)} of ${fmtQty(m.required_qty)} ${m.unit}${
            days !== null ? `, roughly ${days} days of supply` : ""
          }.`,
          tone: "warn" as Tone,
        }
      }),
    })
  }
  if (ordered.length) {
    sections.push({
      heading: "On order",
      lines: ordered.map((m) => ({
        text: `${m.name} (${m.project_name}) — ${fmtQty(m.available_qty)} ${m.unit} expected; not yet consumed on site.`,
      })),
    })
  }
  if (!sections.length) {
    sections.push({
      lines: [
        {
          text: "No materials are at critical or low stock levels right now.",
          tone: "success",
        },
      ],
    })
  }
  return {
    title: "Material Stock Forecast",
    intro: projectId
      ? "Computed from tracked stock levels and average consumption since each project started."
      : "Computed across all projects from tracked stock levels.",
    sections,
    footnote: "Computed from live project data.",
  }
}

function tasksAtRisk(
  db: ReturnType<typeof getDb>,
  projectId: number | null,
  now: Date,
): AssistantAnswer {
  const today = now.toISOString().slice(0, 10)
  const soon = new Date(now.getTime() + 7 * 86400000).toISOString().slice(0, 10)
  const where = projectId ? "AND t.project_id = ?" : ""
  const values = projectId ? [projectId] : []
  const blocked = db
    .prepare(
      `SELECT t.name, t.due_date, p.name AS project_name, u.name AS assignee FROM tasks t
     JOIN projects p ON p.id = t.project_id LEFT JOIN users u ON u.id = t.assignee_id
     WHERE t.status = 'blocked' ${where} ORDER BY t.due_date`,
    )
    .all(...values) as Array<Record<string, unknown>>
  const overdue = db
    .prepare(
      `SELECT t.name, t.due_date, p.name AS project_name, u.name AS assignee FROM tasks t
     JOIN projects p ON p.id = t.project_id LEFT JOIN users u ON u.id = t.assignee_id
     WHERE t.status != 'completed' AND t.status != 'blocked' AND t.due_date IS NOT NULL AND t.due_date < ? ${where} ORDER BY t.due_date`,
    )
    .all(today, ...values) as Array<Record<string, unknown>>
  const dueSoon = db
    .prepare(
      `SELECT t.name, t.due_date, p.name AS project_name, t.priority, u.name AS assignee FROM tasks t
     JOIN projects p ON p.id = t.project_id LEFT JOIN users u ON u.id = t.assignee_id
     WHERE t.status = 'in-progress' AND t.due_date BETWEEN ? AND ? ${where} ORDER BY t.due_date LIMIT 10`,
    )
    .all(today, soon, ...values) as Array<Record<string, unknown>>

  const sections: AssistantSection[] = []
  if (blocked.length) {
    sections.push({
      heading: "Blocked — needs intervention",
      lines: blocked.map((t) => ({
        text: `${t.name} (${t.project_name})${
          t.assignee ? ` — assigned to ${t.assignee}` : ""
        }${t.due_date ? ` · was due ${t.due_date}` : ""}.`,
        tone: "danger",
      })),
    })
  }
  if (overdue.length) {
    sections.push({
      heading: "Overdue",
      lines: overdue.map((t) => ({
        text: `${t.name} (${t.project_name}) — due ${t.due_date}.`,
        tone: "warn",
      })),
    })
  }
  if (dueSoon.length) {
    sections.push({
      heading: "In progress — due within 7 days",
      lines: dueSoon.map((t) => ({
        text: `${t.name} (${t.project_name}) — due ${t.due_date}.`,
      })),
    })
  }
  if (!sections.length) {
    sections.push({
      lines: [
        {
          text: "No blocked, overdue, or soon-due tasks found.",
          tone: "success",
        },
      ],
    })
  }
  return {
    title: "Tasks Behind Schedule or Blocked",
    intro: "Live task status across the selected scope.",
    sections,
    footnote: "Computed from live project data.",
  }
}

function budgetVariance(
  db: ReturnType<typeof getDb>,
  projectId: number | null,
): AssistantAnswer {
  const scope = projectId ? "WHERE bm.project_id = ?" : ""
  const values = projectId ? [projectId] : []
  const months = db
    .prepare(
      `SELECT bm.ym, SUM(bm.planned) AS planned, SUM(bm.actual) AS actual FROM budget_months bm ${scope} GROUP BY bm.ym ORDER BY bm.ym`,
    )
    .all(...values) as Array<{ ym: string; planned: number; actual: number }>

  const totalPlanned = months.reduce((s, m) => s + m.planned, 0)
  const totalActual = months.reduce((s, m) => s + m.actual, 0)
  const variance = totalActual - totalPlanned
  const pct = totalPlanned ? Math.round((variance / totalPlanned) * 100) : 0

  const sections: AssistantSection[] = [
    {
      heading: "Position to date",
      lines: [
        { text: `Planned: KES ${totalPlanned.toLocaleString("en-KE")}` },
        { text: `Actual: KES ${totalActual.toLocaleString("en-KE")}` },
        {
          text:
            variance > 0
              ? `Over budget by KES ${variance.toLocaleString("en-KE")} (${pct}%)`
              : `Under budget by KES ${Math.abs(variance).toLocaleString("en-KE")} (${Math.abs(pct)}%)`,
          tone: variance > 0 ? "danger" : "success",
        },
      ],
    },
  ]
  const worst = [...months]
    .map((m) => ({ ...m, variance: m.actual - m.planned }))
    .sort((a, b) => Math.abs(b.variance) - Math.abs(a.variance))
    .slice(0, 3)
  if (worst.length) {
    sections.push({
      heading: "Largest monthly variances",
      lines: worst.map((m) => ({
        text: `${m.ym}: ${
          m.variance > 0 ? "+" : ""
        }KES ${m.variance.toLocaleString("en-KE")} against plan.`,
        tone: m.variance > 0 ? "warn" as Tone : "success" as Tone,
      })),
    })
  }
  return {
    title: "Budget Variance Analysis",
    intro: "Planned vs. actual monthly expenditure from the budget ledger.",
    sections,
    footnote: "Computed from live project data.",
  }
}

function siteSummary(
  db: ReturnType<typeof getDb>,
  projectId: number | null,
): AssistantAnswer {
  const where = projectId ? "WHERE su.project_id = ?" : ""
  const rows = db
    .prepare(
      `SELECT su.*, p.name AS project_name, u.name AS submitted_by FROM site_updates su
     JOIN projects p ON p.id = su.project_id LEFT JOIN users u ON u.id = su.submitted_by
     ${where} ORDER BY su.date DESC, su.id DESC LIMIT 2`,
    )
    .all(...(projectId ? [projectId] : [])) as Array<Record<string, unknown>>

  if (!rows.length) {
    return {
      title: "Site Report Summary",
      intro: "No site updates have been submitted for this scope yet.",
      sections: [
        {
          lines: [
            {
              text: "Submit a site update from the Site Update page to start building the daily record.",
            },
          ],
        },
      ],
      footnote: "Computed from live project data.",
    }
  }
  const avgWorkers = Math.round(
    rows.reduce((s, r) => s + Number(r.workers_on_site), 0) / rows.length,
  )
  const sections: AssistantSection[] = rows.map((su) => {
    const lines: AssistantLine[] = [
      {
        text: `Weather: ${su.weather || "not recorded"}${
          su.temperature ? ` · ${su.temperature}°C` : ""
        } · ${su.workers_on_site} workers on site`,
      },
      ...String(su.work_completed)
        .split(/(?<=\.)\s+/)
        .slice(0, 4)
        .map((s) => ({ text: s })),
    ]
    if (su.issues) lines.push({ text: `Issues: ${su.issues}`, tone: "warn" })
    return {
      heading: `${su.project_name} — ${su.date} (submitted by ${su.submitted_by ?? "unknown"})`,
      lines,
    }
  })
  return {
    title: "Site Report Summary",
    intro: `Latest ${rows.length} update${
      rows.length === 1 ? "" : "s"
    } · average ${avgWorkers} workers on site.`,
    sections,
    footnote: "Computed from live project data.",
  }
}

function weeklyReport(
  db: ReturnType<typeof getDb>,
  now: Date,
): AssistantAnswer {
  const projects = db
    .prepare("SELECT * FROM projects ORDER BY status != 'Completed' DESC, name")
    .all() as Array<Record<string, unknown>>
  const today = now.toISOString().slice(0, 10)
  const sections: AssistantSection[] = projects.map((p) => {
    const blocked = (db
      .prepare(
        "SELECT COUNT(*) AS c FROM tasks WHERE project_id = ? AND status = 'blocked'",
      )
      .get(Number(p.id)) as { c: number }).c
    const overdue = (db
      .prepare(
        "SELECT COUNT(*) AS c FROM tasks WHERE project_id = ? AND status != 'completed' AND due_date < ?",
      )
      .get(Number(p.id), today) as { c: number }).c
    const alerts = (db
      .prepare(
        "SELECT COUNT(*) AS c FROM materials WHERE project_id = ? AND status IN ('critical','low')",
      )
      .get(Number(p.id)) as { c: number }).c
    const spentPct = Number(p.budget_total)
      ? Math.round((Number(p.budget_spent) / Number(p.budget_total)) * 100)
      : 0
    const lines: AssistantLine[] = [
      {
        text: `${p.status} · ${p.progress}% complete · budget ${spentPct}% spent (KES ${(Number(p.budget_spent) / 1e6).toFixed(1)}M of ${(Number(p.budget_total) / 1e6).toFixed(1)}M)`,
      },
    ]
    if (blocked)
      lines.push({
        text: `${blocked} blocked task${
          blocked === 1 ? "" : "s"
        } need attention.`,
        tone: "danger",
      })
    if (overdue)
      lines.push({
        text: `${overdue} overdue task${overdue === 1 ? "" : "s"}.`,
        tone: "warn",
      })
    if (alerts)
      lines.push({
        text: `${alerts} material alert${
          alerts === 1 ? "" : "s"
        } (critical or low stock).`,
        tone: "warn",
      })
    if (!blocked && !overdue && !alerts)
      lines.push({
        text: "No blocked tasks, overdue items, or stock alerts.",
        tone: "success",
      })
    return { heading: `${p.name} — ${p.location}`, lines }
  })
  return {
    title: "Weekly Project Status Report",
    intro: `All projects · week ending ${now.toISOString().slice(0, 10)}.`,
    sections,
    footnote: "Computed from live project data.",
  }
}

function riskAnalysis(
  db: ReturnType<typeof getDb>,
  projectId: number | null,
  now: Date,
): AssistantAnswer {
  const project = getProject(db, projectId)
  if (!project || !project.start_date || !project.expected_completion) {
    return {
      title: "Project Risk Analysis",
      intro:
        "Select a project with start and completion dates to run a schedule risk check.",
      sections: [
        {
          lines: [
            {
              text: "No project selected, or the project has no programme dates recorded.",
            },
          ],
        },
      ],
      footnote: "Computed from live project data.",
    }
  }
  const totalDays = daysBetween(
    project.start_date,
    new Date(project.expected_completion + "T00:00:00Z"),
  )
  const elapsed = daysBetween(project.start_date, now)
  const elapsedPct = Math.min(100, Math.round((elapsed / totalDays) * 100))
  const progress = project.progress as number

  const criticalMaterials = (db
    .prepare(
      "SELECT COUNT(*) AS c FROM materials WHERE project_id = ? AND status = 'critical'",
    )
    .get(project.id) as { c: number }).c
  const blocked = (db
    .prepare(
      "SELECT COUNT(*) AS c FROM tasks WHERE project_id = ? AND status = 'blocked'",
    )
    .get(project.id) as { c: number }).c
  const overdue = (db
    .prepare(
      "SELECT COUNT(*) AS c FROM tasks WHERE project_id = ? AND status != 'completed' AND due_date IS NOT NULL AND due_date < date('now')",
    )
    .get(project.id) as { c: number }).c
  const varianceRow = db
    .prepare(
      "SELECT SUM(actual) AS a, SUM(planned) AS p FROM budget_months WHERE project_id = ?",
    )
    .get(project.id) as { a: number | null; p: number | null }
  const overBudget =
    varianceRow.a !== null &&
    varianceRow.p !== null &&
    varianceRow.a > varianceRow.p

  const lines: AssistantLine[] = [
    {
      text: `Programme: ${elapsedPct}% of the schedule elapsed, ${progress}% physical progress reported.`,
    },
  ]
  if (progress < elapsedPct - 10) {
    lines.push({
      text: `Progress is trailing the programme by ${elapsedPct - progress} points — the completion date is at risk if the trend continues.`,
      tone: "warn",
    })
  } else {
    lines.push({
      text: "Physical progress is broadly in line with the programme.",
      tone: "success",
    })
  }
  if (criticalMaterials)
    lines.push({
      text: `${criticalMaterials} critical material${
        criticalMaterials === 1 ? "" : "s"
      } could halt site work — reorder immediately.`,
      tone: "danger",
    })
  if (blocked)
    lines.push({
      text: `${blocked} blocked task${
        blocked === 1 ? "" : "s"
      } need management intervention.`,
      tone: "danger",
    })
  if (overdue)
    lines.push({
      text: `${overdue} task${
        overdue === 1 ? " is" : "s are"
      } already past the due date.`,
      tone: "warn",
    })
  if (overBudget)
    lines.push({
      text: "Actual expenditure is running ahead of the plan to date.",
      tone: "warn",
    })
  if (
    !criticalMaterials &&
    !blocked &&
    !overdue &&
    !overBudget &&
    progress >= elapsedPct - 10
  ) {
    lines.push({
      text: "No material schedule or cost risks detected in the current data.",
      tone: "success",
    })
  }
  return {
    title: `Risk Analysis — ${project.name}`,
    intro: `Completion date ${project.expected_completion} · ${Math.max(0, totalDays - elapsed)} days remaining.`,
    sections: [{ lines }],
    footnote: "Computed from live project data.",
  }
}

function overview(db: ReturnType<typeof getDb>, now: Date): AssistantAnswer {
  return weeklyReport(db, now)
}

export const assistantRoutes: RouteDef[] = [
  {
    method: "POST",
    path: "/api/assistant",
    handler: async (ctx) => {
      requireUser(ctx)
      const body = (await readJsonBody(ctx.req)) as Record<string, unknown>
      const question = asString(body.question, "Question", {
        required: true,
        max: 500,
      })
      const projectId = body.projectId
        ? asNumber(body.projectId, "Project")
        : null
      const db = getDb()
      const now = new Date()
      const q = question.toLowerCase()

      let answer: AssistantAnswer
      if (/weekly|status report|generate.*report|all projects/.test(q)) {
        answer = weeklyReport(db, now)
      } else if (/summar|site (report|update)|diary/.test(q)) {
        answer = siteSummary(db, projectId)
      } else if (/budget|spend|cost|variance|money|expenditure/.test(q)) {
        answer = budgetVariance(db, projectId)
      } else if (/material|stock|run low|reorder|supply|deliver/.test(q)) {
        answer = materialsForecast(db, projectId, now)
      } else if (/task|blocked|behind|overdue|schedule/.test(q)) {
        answer = tasksAtRisk(db, projectId, now)
      } else if (/why|risk|delay|miss|completion date|on track/.test(q)) {
        answer = riskAnalysis(db, projectId, now)
      } else {
        answer = overview(db, now)
      }
      json(ctx.res, 200, { answer })
    },
  },
]
