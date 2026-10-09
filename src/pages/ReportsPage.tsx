import { useCallback, useEffect, useMemo, useState } from "react"
import {
  AreaChart,
  Area,
  BarChart,
  Bar,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  Legend,
} from "recharts"
import type { NavigateFn } from "../App"
import { api } from "../lib/api"
import type { BudgetMonth, Material, Project, Task } from "../lib/api"
import { fmtDate, kes } from "../lib/format"
import {
  ErrorBanner,
  LoadingBlock,
  MaterialBadge,
  Progress,
  StatusBadge,
} from "../components/ui"

interface Props {
  onNavigate: NavigateFn
}

const REPORT_TYPES = [
  { id: "project-status", label: "Project status" },
  { id: "budget", label: "Budget" },
  { id: "materials", label: "Materials" },
  { id: "tasks", label: "Tasks" },
] as const
type ReportType = typeof REPORT_TYPES[number]["id"]

export default function ReportsPage(_props: Props) {
  const [projects, setProjects] = useState<Project[]>([])
  const [tasks, setTasks] = useState<Task[]>([])
  const [materials, setMaterials] = useState<Material[]>([])
  const [budget, setBudget] = useState<BudgetMonth[]>([])
  const [selectedProject, setSelectedProject] = useState<string>("all")
  const [reportType, setReportType] = useState<ReportType>("project-status")
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState("")

  const load = useCallback(async () => {
    setLoading(true)
    setError("")
    try {
      const [p, t, m] = await Promise.all([
        api.get<{ projects: Project[] }>("/api/projects"),
        api.get<{ tasks: Task[] }>("/api/tasks"),
        api.get<{ materials: Material[] }>("/api/materials"),
      ])
      setProjects(p.projects)
      setTasks(t.tasks)
      setMaterials(m.materials)
    } catch (err) {
      setError(
        err instanceof Error ? err.message : "Could not load report data.",
      )
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    void load()
  }, [load])

  useEffect(() => {
    const pid = selectedProject === "all" ? null : Number(selectedProject)
    if (pid === null) {
      setBudget([])
      return
    }
    api
      .get<{ budget: BudgetMonth[] }>(`/api/projects/${pid}/budget`)
      .then((d) => setBudget(d.budget))
      .catch(() => setBudget([]))
  }, [selectedProject])

  const scoped = useMemo(() => {
    const pid = selectedProject === "all" ? null : Number(selectedProject)
    return {
      projects: pid ? projects.filter((p) => p.id === pid) : projects,
      tasks: pid ? tasks.filter((t) => t.project_id === pid) : tasks,
      materials: pid
        ? materials.filter((m) => m.project_id === pid)
        : materials,
    }
  }, [projects, tasks, materials, selectedProject])

  const totalBudget = scoped.projects.reduce((s, p) => s + p.budgetTotal, 0)
  const totalSpent = scoped.projects.reduce((s, p) => s + p.budgetSpent, 0)
  const avgProgress = scoped.projects.length
    ? Math.round(
        scoped.projects.reduce((s, p) => s + p.progress, 0) /
          scoped.projects.length,
      )
    : 0

  const progressData = scoped.projects.map((p) => ({
    name: p.name.length > 22 ? p.name.slice(0, 22) + "…" : p.name,
    Progress: p.progress,
    "Budget spent": p.budgetTotal
      ? Math.round((p.budgetSpent / p.budgetTotal) * 100)
      : 0,
  }))

  const budgetChart = budget.map((b) => ({
    month: b.ym.slice(5) + "/" + b.ym.slice(2, 4),
    Planned: b.planned,
    Actual: b.actual,
  }))

  const taskCompletionData = scoped.projects.map((p) => ({
    name: p.name.length > 18 ? p.name.slice(0, 18) + "…" : p.name,
    Completed: p.stats.tasksCompleted,
    Open: p.stats.tasksTotal - p.stats.tasksCompleted,
  }))

  const taskStatusCounts = STATUSES.map(
    (s) => scoped.tasks.filter((t) => t.status === s).length,
  )

  if (loading) return <LoadingBlock label="Preparing reports…" />

  return (
    <div>
      <div className="border-b border-slate-200 bg-white px-4 py-5 sm:px-8">
        <h1 className="text-xl font-bold tracking-tight text-slate-900">
          Reports
        </h1>
        <p className="mt-0.5 text-sm text-slate-500">
          Analytics computed from live project data
        </p>

        <div
          className="mt-4 flex gap-1 overflow-x-auto"
          role="tablist"
          aria-label="Report type"
        >
          {REPORT_TYPES.map((r) => (
            <button
              key={r.id}
              role="tab"
              aria-selected={reportType === r.id}
              onClick={() => setReportType(r.id)}
              className={`whitespace-nowrap border-b-2 px-3.5 py-2.5 text-sm font-medium transition-colors ${
                reportType === r.id
                  ? "border-slate-900 text-slate-900"
                  : "border-transparent text-slate-500 hover:text-slate-800"
              }`}
            >
              {r.label}
            </button>
          ))}
        </div>
      </div>

      <div className="space-y-5 px-4 py-6 sm:px-8">
        {error && <ErrorBanner message={error} onRetry={load} />}

        <div className="flex flex-wrap items-center gap-3">
          <label
            htmlFor="rep-project"
            className="text-sm font-medium text-slate-600"
          >
            Project
          </label>
          <select
            id="rep-project"
            value={selectedProject}
            onChange={(e) => setSelectedProject(e.target.value)}
            className="rounded border border-slate-300 bg-white px-3 py-2 text-sm outline-none focus:border-slate-800"
          >
            <option value="all">All projects</option>
            {projects.map((p) => (
              <option key={p.id} value={p.id}>
                {p.name}
              </option>
            ))}
          </select>
        </div>

        {/* ── PROJECT STATUS ── */}
        {reportType === "project-status" && (
          <div className="space-y-4">
            <dl className="grid grid-cols-2 gap-3 xl:grid-cols-4">
              {[
                {
                  label: "Projects in scope",
                  value: String(scoped.projects.length),
                },
                { label: "Average progress", value: `${avgProgress}%` },
                { label: "Total budget", value: kes(totalBudget) },
                { label: "Total spent", value: kes(totalSpent) },
              ].map((s) => (
                <div
                  key={s.label}
                  className="rounded border border-slate-200 bg-white p-4"
                >
                  <dt className="text-xs font-semibold uppercase tracking-wide text-slate-400">
                    {s.label}
                  </dt>
                  <dd className="mt-1.5 text-xl font-bold text-slate-900">
                    {s.value}
                  </dd>
                </div>
              ))}
            </dl>

            {progressData.length > 0 && (
              <div className="rounded border border-slate-200 bg-white p-5">
                <h3 className="mb-4 text-sm font-semibold text-slate-900">
                  Progress overview (%)
                </h3>
                <ResponsiveContainer width="100%" height={220}>
                  <BarChart
                    data={progressData}
                    margin={{ top: 4, right: 4, left: 0, bottom: 0 }}
                  >
                    <CartesianGrid
                      strokeDasharray="3 3"
                      stroke="#f1f5f9"
                      vertical={false}
                    />
                    <XAxis
                      dataKey="name"
                      tick={{ fontSize: 10, fill: "#94a3b8" }}
                      axisLine={false}
                      tickLine={false}
                    />
                    <YAxis
                      domain={[0, 100]}
                      tick={{ fontSize: 11, fill: "#94a3b8" }}
                      axisLine={false}
                      tickLine={false}
                      tickFormatter={(v) => `${v}%`}
                    />
                    <Tooltip
                      contentStyle={{
                        border: "1px solid #e2e8f0",
                        borderRadius: 4,
                        fontSize: 12,
                      }}
                      formatter={(v) => `${v}%`}
                    />
                    <Legend wrapperStyle={{ fontSize: 12 }} />
                    <Bar
                      dataKey="Progress"
                      fill="#0f172a"
                      radius={[2, 2, 0, 0]}
                      maxBarSize={48}
                    />
                    <Bar
                      dataKey="Budget spent"
                      fill="#ea580c"
                      radius={[2, 2, 0, 0]}
                      maxBarSize={48}
                    />
                  </BarChart>
                </ResponsiveContainer>
              </div>
            )}

            <div className="overflow-hidden rounded border border-slate-200 bg-white">
              <div className="overflow-x-auto">
                <table className="w-full">
                  <thead>
                    <tr className="border-b border-slate-100 bg-slate-50">
                      {[
                        "Project",
                        "Status",
                        "Progress",
                        "Budget spent",
                        "Due date",
                        "Tasks",
                      ].map((h) => (
                        <th
                          key={h}
                          className="px-4 py-3 text-left text-xs font-semibold uppercase tracking-wide text-slate-500"
                        >
                          {h}
                        </th>
                      ))}
                    </tr>
                  </thead>
                  <tbody>
                    {scoped.projects.map((p) => (
                      <tr
                        key={p.id}
                        className="border-b border-slate-50 last:border-0"
                      >
                        <td className="px-4 py-3 text-sm font-medium text-slate-700">
                          {p.name}
                        </td>
                        <td className="px-4 py-3">
                          <StatusBadge status={p.status} />
                        </td>
                        <td className="px-4 py-3">
                          <div className="flex items-center gap-2">
                            <Progress value={p.progress} />
                            <span className="w-8 text-right text-xs font-semibold text-slate-600">
                              {p.progress}%
                            </span>
                          </div>
                        </td>
                        <td className="px-4 py-3 text-sm text-slate-600">
                          {kes(p.budgetSpent)}
                          <span className="text-xs text-slate-400">
                            {" "}
                            / {kes(p.budgetTotal)}
                          </span>
                        </td>
                        <td className="px-4 py-3 text-xs text-slate-500">
                          {fmtDate(p.expectedCompletion)}
                        </td>
                        <td className="px-4 py-3 text-xs text-slate-600">
                          {p.stats.tasksCompleted}/{p.stats.tasksTotal}
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
              {scoped.projects.length === 0 && (
                <p className="py-10 text-center text-sm text-slate-500">
                  No projects in scope.
                </p>
              )}
            </div>
          </div>
        )}

        {/* ── BUDGET ── */}
        {reportType === "budget" && (
          <div className="space-y-4">
            {selectedProject === "all" ? (
              <div className="rounded border border-slate-200 bg-white p-8 text-center">
                <p className="text-sm font-semibold text-slate-700">
                  Select a project
                </p>
                <p className="mt-1 text-sm text-slate-500">
                  Monthly budget data is tracked per project. Choose one above.
                </p>
              </div>
            ) : budgetChart.length === 0 ? (
              <div className="rounded border border-slate-200 bg-white p-8 text-center">
                <p className="text-sm font-semibold text-slate-700">
                  No budget data recorded
                </p>
                <p className="mt-1 text-sm text-slate-500">
                  Planned and actual figures have not been entered for this
                  project yet.
                </p>
              </div>
            ) : (
              <>
                <div className="rounded border border-slate-200 bg-white p-5">
                  <h3 className="mb-4 text-sm font-semibold text-slate-900">
                    Monthly expenditure (KES)
                  </h3>
                  <ResponsiveContainer width="100%" height={260}>
                    <AreaChart
                      data={budgetChart}
                      margin={{ top: 4, right: 4, left: 0, bottom: 0 }}
                    >
                      <defs>
                        <linearGradient id="rp-p" x1="0" y1="0" x2="0" y2="1">
                          <stop
                            offset="5%"
                            stopColor="#0f172a"
                            stopOpacity={0.15}
                          />
                          <stop
                            offset="95%"
                            stopColor="#0f172a"
                            stopOpacity={0}
                          />
                        </linearGradient>
                        <linearGradient id="rp-a" x1="0" y1="0" x2="0" y2="1">
                          <stop
                            offset="5%"
                            stopColor="#ea580c"
                            stopOpacity={0.15}
                          />
                          <stop
                            offset="95%"
                            stopColor="#ea580c"
                            stopOpacity={0}
                          />
                        </linearGradient>
                      </defs>
                      <CartesianGrid
                        strokeDasharray="3 3"
                        stroke="#f1f5f9"
                        vertical={false}
                      />
                      <XAxis
                        dataKey="month"
                        tick={{ fontSize: 11, fill: "#94a3b8" }}
                        axisLine={false}
                        tickLine={false}
                      />
                      <YAxis
                        tick={{ fontSize: 10, fill: "#94a3b8" }}
                        axisLine={false}
                        tickLine={false}
                        tickFormatter={(v) => `${Math.round(v / 1000)}K`}
                      />
                      <Tooltip
                        contentStyle={{
                          border: "1px solid #e2e8f0",
                          borderRadius: 4,
                          fontSize: 12,
                        }}
                        formatter={(v) =>
                          `KES ${Number(v).toLocaleString("en-KE")}`
                        }
                      />
                      <Legend wrapperStyle={{ fontSize: 12 }} />
                      <Area
                        type="monotone"
                        dataKey="Planned"
                        stroke="#0f172a"
                        strokeWidth={2}
                        fill="url(#rp-p)"
                        dot={false}
                      />
                      <Area
                        type="monotone"
                        dataKey="Actual"
                        stroke="#ea580c"
                        strokeWidth={2}
                        fill="url(#rp-a)"
                        dot={false}
                      />
                    </AreaChart>
                  </ResponsiveContainer>
                </div>

                <div className="overflow-hidden rounded border border-slate-200 bg-white">
                  <div className="overflow-x-auto">
                    <table className="w-full">
                      <thead>
                        <tr className="border-b border-slate-100 bg-slate-50">
                          {["Month", "Planned", "Actual", "Variance"].map(
                            (h) => (
                              <th
                                key={h}
                                className="px-4 py-3 text-left text-xs font-semibold uppercase tracking-wide text-slate-500"
                              >
                                {h}
                              </th>
                            ),
                          )}
                        </tr>
                      </thead>
                      <tbody>
                        {budget.map((row) => {
                          const variance = row.actual - row.planned
                          return (
                            <tr
                              key={row.ym}
                              className="border-b border-slate-50 last:border-0"
                            >
                              <td className="px-4 py-2.5 text-sm font-medium text-slate-700">
                                {row.ym}
                              </td>
                              <td className="px-4 py-2.5 text-sm text-slate-600">
                                {row.planned.toLocaleString("en-KE")}
                              </td>
                              <td className="px-4 py-2.5 text-sm text-slate-600">
                                {row.actual.toLocaleString("en-KE")}
                              </td>
                              <td
                                className={`px-4 py-2.5 text-sm font-medium ${
                                  variance > 0
                                    ? "text-red-700"
                                    : "text-green-700"
                                }`}
                              >
                                {variance > 0 ? "+" : ""}
                                {variance.toLocaleString("en-KE")}
                              </td>
                            </tr>
                          )
                        })}
                      </tbody>
                    </table>
                  </div>
                </div>
              </>
            )}
          </div>
        )}

        {/* ── MATERIALS ── */}
        {reportType === "materials" && (
          <div className="overflow-hidden rounded border border-slate-200 bg-white">
            <div className="overflow-x-auto">
              <table className="w-full">
                <thead>
                  <tr className="border-b border-slate-100 bg-slate-50">
                    {[
                      "Material",
                      "Project",
                      "Required",
                      "Used",
                      "Available",
                      "Usage",
                      "Status",
                    ].map((h) => (
                      <th
                        key={h}
                        className="px-4 py-3 text-left text-xs font-semibold uppercase tracking-wide text-slate-500"
                      >
                        {h}
                      </th>
                    ))}
                  </tr>
                </thead>
                <tbody>
                  {scoped.materials.map((m) => {
                    const usagePct = m.required_qty
                      ? Math.round((m.used_qty / m.required_qty) * 100)
                      : 0
                    return (
                      <tr
                        key={m.id}
                        className="border-b border-slate-50 last:border-0"
                      >
                        <td className="px-4 py-2.5 text-sm font-medium text-slate-700">
                          {m.name}
                        </td>
                        <td className="px-4 py-2.5 text-xs text-slate-500">
                          {m.project_name}
                        </td>
                        <td className="px-4 py-2.5 text-sm text-slate-600">
                          {m.required_qty.toLocaleString()} {m.unit}
                        </td>
                        <td className="px-4 py-2.5 text-sm text-slate-600">
                          {m.used_qty.toLocaleString()} {m.unit}
                        </td>
                        <td
                          className={`px-4 py-2.5 text-sm font-medium ${
                            m.status === "critical"
                              ? "text-red-700"
                              : "text-slate-700"
                          }`}
                        >
                          {m.available_qty.toLocaleString()} {m.unit}
                        </td>
                        <td className="px-4 py-2.5">
                          <div className="flex items-center gap-2">
                            <Progress value={usagePct} />
                            <span className="w-8 text-right text-xs text-slate-600">
                              {usagePct}%
                            </span>
                          </div>
                        </td>
                        <td className="px-4 py-2.5">
                          <MaterialBadge status={m.status} />
                        </td>
                      </tr>
                    )
                  })}
                </tbody>
              </table>
            </div>
            {scoped.materials.length === 0 && (
              <p className="py-10 text-center text-sm text-slate-500">
                No materials in scope.
              </p>
            )}
          </div>
        )}

        {/* ── TASKS ── */}
        {reportType === "tasks" && (
          <div className="space-y-4">
            <dl className="grid grid-cols-2 gap-3 xl:grid-cols-4">
              {["Completed", "In progress", "Not started", "Blocked"].map(
                (label, i) => (
                  <div
                    key={label}
                    className="rounded border border-slate-200 bg-white p-4"
                  >
                    <dt className="text-xs font-semibold uppercase tracking-wide text-slate-400">
                      {label}
                    </dt>
                    <dd
                      className={`mt-1.5 text-xl font-bold ${
                        i === 3 && taskStatusCounts[i] > 0
                          ? "text-red-700"
                          : "text-slate-900"
                      }`}
                    >
                      {taskStatusCounts[i]}
                    </dd>
                  </div>
                ),
              )}
            </dl>
            <div className="rounded border border-slate-200 bg-white p-5">
              <h3 className="mb-4 text-sm font-semibold text-slate-900">
                Task completion by project
              </h3>
              {taskCompletionData.length === 0 ? (
                <p className="py-8 text-center text-sm text-slate-500">
                  No tasks in scope.
                </p>
              ) : (
                <ResponsiveContainer width="100%" height={220}>
                  <BarChart
                    data={taskCompletionData}
                    margin={{ top: 4, right: 4, left: 0, bottom: 0 }}
                  >
                    <CartesianGrid
                      strokeDasharray="3 3"
                      stroke="#f1f5f9"
                      vertical={false}
                    />
                    <XAxis
                      dataKey="name"
                      tick={{ fontSize: 10, fill: "#94a3b8" }}
                      axisLine={false}
                      tickLine={false}
                    />
                    <YAxis
                      tick={{ fontSize: 11, fill: "#94a3b8" }}
                      axisLine={false}
                      tickLine={false}
                    />
                    <Tooltip
                      contentStyle={{
                        border: "1px solid #e2e8f0",
                        borderRadius: 4,
                        fontSize: 12,
                      }}
                    />
                    <Legend wrapperStyle={{ fontSize: 12 }} />
                    <Bar dataKey="Completed" stackId="a" fill="#16a34a" />
                    <Bar dataKey="Open" stackId="a" fill="#e2e8f0" />
                  </BarChart>
                </ResponsiveContainer>
              )}
            </div>
          </div>
        )}
      </div>
    </div>
  )
}

const STATUSES = ["completed", "in-progress", "not-started", "blocked"] as const
