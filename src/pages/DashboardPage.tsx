import { useCallback, useEffect, useState } from "react"
import {
  AlertTriangle,
  Calendar,
  CheckSquare,
  ChevronRight,
  Clock,
  FileEdit,
  FileText,
  MapPin,
  MessageSquare,
  Package,
} from "lucide-react"
import type { NavigateFn } from "../App"
import { api } from "../lib/api"
import type { ActivityItem, DashboardData } from "../lib/api"
import { fmtDate, kesShort, relTime } from "../lib/format"
import {
  ErrorBanner,
  LoadingBlock,
  Progress,
  StatusBadge,
} from "../components/ui"
import { useAuth } from "../lib/auth"

interface Props {
  onNavigate: NavigateFn
}

// Normalise the API payload so a partially-formed response (missing stats,
// null stats, missing arrays) degrades to zeros/empty lists instead of
// crashing the page. Backend errors still surface through `error`.
function normalizeDashboard(raw: unknown): DashboardData {
  const d = (raw ?? {}) as Partial<DashboardData>
  const stats = (d.stats ?? {}) as Partial<DashboardData["stats"]>
  return {
    stats: {
      active_projects: stats.active_projects ?? 0,
      tasks_due: stats.tasks_due ?? 0,
      blocked_tasks: stats.blocked_tasks ?? 0,
      material_alerts: stats.material_alerts ?? 0,
    },
    projects: (d.projects ?? []).map((p) => ({
      ...p,
      team: p.team ?? [],
      stats:
        p.stats ??
        ({
          tasksTotal: 0,
          tasksCompleted: 0,
          tasksDueThisWeek: 0,
          tasksBlocked: 0,
          materialsAlert: 0,
          documentsRecent: 0,
        } satisfies DashboardData["projects"][number]["stats"]),
    })),
    materialAlerts: d.materialAlerts ?? [],
    deadlines: d.deadlines ?? [],
    activity: d.activity ?? [],
  }
}

function ActivityIcon({ type }: { type: string }) {
  const props = { size: 14, strokeWidth: 1.8 }
  const icons: Record<string, React.ReactNode> = {
    "site-update": <FileEdit {...props} />,
    document: <FileText {...props} />,
    task: <CheckSquare {...props} />,
    material: <Package {...props} />,
    project: <MapPin {...props} />,
    comment: <MessageSquare {...props} />,
  }
  return (
    <span
      className="flex h-7 w-7 shrink-0 items-center justify-center rounded-full bg-slate-100 text-slate-500"
      aria-hidden="true"
    >
      {icons[type] ?? <FileText {...props} />}
    </span>
  )
}

export default function DashboardPage({ onNavigate }: Props) {
  const { user } = useAuth()
  const [data, setData] = useState<DashboardData | null>(null)
  const [error, setError] = useState("")
  const [loading, setLoading] = useState(true)

  const load = useCallback(async () => {
    setLoading(true)
    setError("")
    try {
      setData(normalizeDashboard(await api.get<DashboardData>("/api/dashboard")))
    } catch (err) {
      setData(null)
      setError(
        err instanceof Error ? err.message : "Could not load the dashboard.",
      )
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    void load()
  }, [load])

  if (loading) return <LoadingBlock label="Loading dashboard…" />
  if (error)
    return (
      <div className="p-6">
        <ErrorBanner message={error} onRetry={load} />
      </div>
    )
  if (!data) return null

  const firstName = user?.name.split(" ")[0] ?? "there"
  const dateStr = new Date().toLocaleDateString("en-KE", {
    weekday: "long",
    day: "numeric",
    month: "long",
    year: "numeric",
  })
  const activeProjects = data.projects.filter((p) => p.status !== "Completed")

  return (
    <div>
      {/* Page header */}
      <div className="border-b border-slate-200 bg-white px-4 py-5 sm:px-8">
        <div className="flex flex-wrap items-start justify-between gap-3">
          <div>
            <h1 className="text-xl font-bold tracking-tight text-slate-900">
              Hello, {firstName}.
            </h1>
            <p className="mt-0.5 text-sm text-slate-500">{dateStr}</p>
          </div>
          <button
            onClick={() => onNavigate("site-update")}
            className="inline-flex items-center gap-2 rounded bg-slate-900 px-4 py-2 text-sm font-semibold text-white transition-colors hover:bg-slate-700 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-slate-800"
          >
            <FileEdit size={15} strokeWidth={2} />
            Add site update
          </button>
        </div>

        {/* Quick stats */}
        <div className="mt-5 flex flex-wrap gap-x-8 gap-y-2 border-t border-slate-100 pt-4">
          {[
            { label: "Active projects", value: data.stats.active_projects },
            { label: "Tasks due this week", value: data.stats.tasks_due },
            {
              label: "Material alerts",
              value: data.stats.material_alerts,
              urgent: data.stats.material_alerts > 0,
            },
            {
              label: "Blocked tasks",
              value: data.stats.blocked_tasks,
              urgent: data.stats.blocked_tasks > 0,
            },
          ].map((s) => (
            <div key={s.label} className="flex items-baseline gap-2">
              <span
                className={`text-lg font-bold ${
                  s.urgent ? "text-orange-700" : "text-slate-900"
                }`}
              >
                {s.value}
              </span>
              <span
                className={`text-sm ${
                  s.urgent ? "font-medium text-orange-700" : "text-slate-500"
                }`}
              >
                {s.label}
              </span>
            </div>
          ))}
        </div>
      </div>

      <div className="space-y-6 px-4 py-6 sm:px-8">
        {/* Projects */}
        <section aria-labelledby="dash-projects">
          <div className="mb-3 flex items-center justify-between">
            <h2
              id="dash-projects"
              className="text-base font-semibold text-slate-900"
            >
              Projects
            </h2>
            <button
              onClick={() => onNavigate("projects")}
              className="inline-flex items-center gap-1 text-sm font-medium text-slate-600 hover:text-slate-900"
            >
              All projects <ChevronRight size={15} aria-hidden="true" />
            </button>
          </div>

          {activeProjects.length === 0 ? (
            <div className="rounded border border-slate-200 bg-white py-14 text-center">
              <p className="text-sm font-semibold text-slate-700">
                No active projects
              </p>
              <p className="mt-1 text-sm text-slate-500">
                Projects appear here once work begins.
              </p>
            </div>
          ) : (
            <div className="grid gap-4 lg:grid-cols-2">
              {activeProjects.map((project) => (
                <button
                  key={project.id}
                  onClick={() => onNavigate("project-detail", project.id)}
                  className="rounded border border-slate-200 bg-white text-left transition-colors hover:border-slate-400 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-slate-800"
                >
                  <div className="p-4">
                    <div className="mb-3 flex items-start justify-between gap-3">
                      <div className="min-w-0">
                        <h3 className="truncate text-sm font-semibold text-slate-900">
                          {project.name}
                        </h3>
                        <div className="mt-0.5 flex items-center gap-1 text-xs text-slate-400">
                          <MapPin size={12} aria-hidden="true" />
                          {project.location}
                        </div>
                      </div>
                      <div className="shrink-0 text-right">
                        <span className="text-sm font-bold text-slate-900">
                          {project.progress}%
                        </span>
                        <div className="text-xs text-slate-400">complete</div>
                      </div>
                    </div>

                    <Progress value={project.progress} className="mb-4" />

                    <div className="mb-3 grid grid-cols-2 gap-3 border-t border-slate-100 pt-3 text-xs sm:grid-cols-4">
                      <div>
                        <div className="font-medium text-slate-400">Budget</div>
                        <div className="mt-0.5 font-semibold text-slate-800">
                          {kesShort(project.budgetSpent)}{" "}
                          <span className="font-normal text-slate-400">
                            / {kesShort(project.budgetTotal)}
                          </span>
                        </div>
                      </div>
                      <div>
                        <div className="font-medium text-slate-400">
                          Tasks due
                        </div>
                        <div
                          className={`mt-0.5 font-semibold ${
                            project.stats.tasksDueThisWeek > 5
                              ? "text-orange-700"
                              : "text-slate-800"
                          }`}
                        >
                          {project.stats.tasksDueThisWeek} this week
                        </div>
                      </div>
                      <div>
                        <div className="font-medium text-slate-400">
                          Materials
                        </div>
                        <div
                          className={`mt-0.5 font-semibold ${
                            project.stats.materialsAlert > 0
                              ? "text-orange-700"
                              : "text-slate-800"
                          }`}
                        >
                          {project.stats.materialsAlert > 0
                            ? `${project.stats.materialsAlert} alert${
                                project.stats.materialsAlert > 1 ? "s" : ""
                              }`
                            : "OK"}
                        </div>
                      </div>
                      <div>
                        <div className="font-medium text-slate-400">Due</div>
                        <div className="mt-0.5 font-semibold text-slate-800">
                          {fmtDate(project.expectedCompletion)}
                        </div>
                      </div>
                    </div>

                    <div className="flex items-center justify-between">
                      <div className="flex -space-x-1.5">
                        {project.team.slice(0, 5).map((m) => (
                          <span
                            key={m.id}
                            title={m.name}
                            className="flex h-6 w-6 items-center justify-center rounded-full text-[9px] font-semibold text-white ring-2 ring-white"
                            style={{ background: m.color }}
                          >
                            {m.initials}
                          </span>
                        ))}
                        {project.team.length > 5 && (
                          <span className="flex h-6 w-6 items-center justify-center rounded-full bg-slate-100 text-[9px] font-semibold text-slate-500 ring-2 ring-white">
                            +{project.team.length - 5}
                          </span>
                        )}
                      </div>
                      <StatusBadge status={project.status} />
                    </div>
                  </div>
                </button>
              ))}
            </div>
          )}
        </section>

        {/* Bottom two-column */}
        <div className="grid gap-4 xl:grid-cols-2">
          {/* Recent activity */}
          <div className="rounded border border-slate-200 bg-white">
            <div className="flex items-center justify-between border-b border-slate-100 px-5 py-4">
              <h3 className="text-sm font-semibold text-slate-900">
                Recent activity
              </h3>
              <span className="text-xs text-slate-400">All projects</span>
            </div>
            {data.activity.length === 0 ? (
              <p className="px-5 py-8 text-center text-sm text-slate-500">
                No activity recorded yet.
              </p>
            ) : (
              <ul className="divide-y divide-slate-50">
                {data.activity.map((item: ActivityItem) => (
                  <li
                    key={item.id}
                    className="flex items-start gap-3 px-5 py-3"
                  >
                    <ActivityIcon type={item.type} />
                    <div className="min-w-0 flex-1">
                      <p className="text-sm leading-snug text-slate-700">
                        {item.text}
                      </p>
                      <p className="mt-1 text-xs text-slate-400">
                        {item.actor_name} · {relTime(item.created_at)}
                      </p>
                    </div>
                  </li>
                ))}
              </ul>
            )}
          </div>

          <div className="space-y-4">
            {/* Material alerts */}
            <div className="rounded border border-slate-200 bg-white">
              <div className="flex items-center gap-2 border-b border-slate-100 px-5 py-4">
                <AlertTriangle
                  size={15}
                  className="text-orange-600"
                  aria-hidden="true"
                />
                <h3 className="text-sm font-semibold text-slate-900">
                  Material alerts
                </h3>
                {data.materialAlerts.length > 0 && (
                  <span className="ml-auto rounded bg-orange-50 px-1.5 py-0.5 text-xs font-medium text-orange-700">
                    {data.materialAlerts.length}
                  </span>
                )}
              </div>
              {data.materialAlerts.length === 0 ? (
                <p className="px-5 py-6 text-center text-sm text-slate-500">
                  All stock levels are healthy.
                </p>
              ) : (
                <ul className="divide-y divide-slate-50">
                  {data.materialAlerts.map((m) => (
                    <li
                      key={m.id}
                      className="flex items-center justify-between gap-3 px-5 py-2.5"
                    >
                      <div className="min-w-0">
                        <p className="truncate text-sm font-medium text-slate-700">
                          {m.name}
                        </p>
                        <p className="mt-0.5 text-xs text-slate-400">
                          {m.project_name} · {m.available_qty} {m.unit} left
                        </p>
                      </div>
                      <span
                        className={`shrink-0 rounded px-2 py-0.5 text-xs font-medium ${
                          m.status === "critical"
                            ? "bg-red-50 text-red-700"
                            : "bg-orange-50 text-orange-700"
                        }`}
                      >
                        {m.status === "critical" ? "Critical" : "Low"}
                      </span>
                    </li>
                  ))}
                </ul>
              )}
              <div className="border-t border-slate-100 px-5 py-3">
                <button
                  onClick={() => onNavigate("materials")}
                  className="text-xs font-medium text-slate-600 hover:text-slate-900"
                >
                  View all materials →
                </button>
              </div>
            </div>

            {/* Upcoming deadlines */}
            <div className="rounded border border-slate-200 bg-white">
              <div className="flex items-center gap-2 border-b border-slate-100 px-5 py-4">
                <Calendar
                  size={15}
                  className="text-slate-500"
                  aria-hidden="true"
                />
                <h3 className="text-sm font-semibold text-slate-900">
                  Upcoming deadlines
                </h3>
              </div>
              {data.deadlines.length === 0 ? (
                <p className="px-5 py-6 text-center text-sm text-slate-500">
                  Nothing due in the next 14 days.
                </p>
              ) : (
                <ul className="divide-y divide-slate-50">
                  {data.deadlines.map((t) => (
                    <li
                      key={t.id}
                      className="flex items-start justify-between gap-3 px-5 py-2.5"
                    >
                      <div className="min-w-0">
                        <p className="truncate text-sm leading-snug text-slate-700">
                          {t.name}
                        </p>
                        <p className="mt-0.5 text-xs text-slate-400">
                          {t.project_name}
                        </p>
                      </div>
                      <span className="shrink-0 text-xs font-semibold text-slate-600">
                        {fmtDate(t.due_date)}
                      </span>
                    </li>
                  ))}
                </ul>
              )}
              <div className="border-t border-slate-100 px-5 py-3">
                <button
                  onClick={() => onNavigate("tasks")}
                  className="text-xs font-medium text-slate-600 hover:text-slate-900"
                >
                  View all tasks →
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}
