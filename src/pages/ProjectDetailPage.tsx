import { useCallback, useEffect, useState } from "react"
import { AlertTriangle, ChevronLeft, FileEdit, User, Users } from "lucide-react"
import {
  AreaChart,
  Area,
  XAxis,
  YAxis,
  Tooltip,
  ResponsiveContainer,
  CartesianGrid,
} from "recharts"
import type { NavigateFn } from "../App"
import { api, ApiRequestError } from "../lib/api"
import type {
  ActivityItem,
  BudgetMonth,
  DocumentItem,
  Material,
  Project,
  Task,
} from "../lib/api"
import { fmtDate, kes, kesShort, relTime } from "../lib/format"
import {
  ErrorBanner,
  LoadingBlock,
  MaterialBadge,
  Progress,
  StatusBadge,
  TaskBadge,
} from "../components/ui"

interface Props {
  projectId: number
  onNavigate: NavigateFn
}

interface ProjectDetail {
  project: Project
  tasks: Task[]
  materials: Material[]
  documents: DocumentItem[]
  activity: ActivityItem[]
  budget: BudgetMonth[]
}

const TABS = [
  "Overview",
  "Tasks",
  "Materials",
  "Budget",
  "Documents",
  "Team",
  "Activity",
] as const
type Tab = typeof TABS[number]

const budgetChart = (budget: BudgetMonth[]) =>
  budget.map((b) => ({
    month: b.ym.slice(5) + "/" + b.ym.slice(2, 4),
    Planned: b.planned,
    Actual: b.actual,
  }))

export default function ProjectDetailPage({ projectId, onNavigate }: Props) {
  const [tab, setTab] = useState<Tab>("Overview")
  const [data, setData] = useState<ProjectDetail | null>(null)
  const [error, setError] = useState("")
  const [loading, setLoading] = useState(true)
  const [updatingTask, setUpdatingTask] = useState<number | null>(null)

  const load = useCallback(async () => {
    setLoading(true)
    setError("")
    try {
      setData(await api.get<ProjectDetail>(`/api/projects/${projectId}`))
    } catch (err) {
      setError(
        err instanceof Error ? err.message : "Could not load the project.",
      )
    } finally {
      setLoading(false)
    }
  }, [projectId])

  useEffect(() => {
    void load()
  }, [load])

  // Reset to Overview when switching projects.
  useEffect(() => setTab("Overview"), [projectId])

  const setTaskStatus = async (task: Task, status: Task["status"]) => {
    if (status === task.status) return
    setUpdatingTask(task.id)
    try {
      await api.patch(`/api/tasks/${task.id}`, { status })
      await load()
    } catch (err) {
      setError(
        err instanceof ApiRequestError
          ? err.message
          : "Could not update the task.",
      )
    } finally {
      setUpdatingTask(null)
    }
  }

  if (loading) return <LoadingBlock label="Loading project…" />
  if (error && !data)
    return (
      <div className="p-6">
        <ErrorBanner message={error} onRetry={load} />
      </div>
    )
  if (!data) return null

  const { project, tasks, materials, documents, activity, budget } = data
  const spentPct = project.budgetTotal
    ? Math.round((project.budgetSpent / project.budgetTotal) * 100)
    : 0
  const chartData = budgetChart(budget)

  return (
    <div>
      {/* Header */}
      <div className="border-b border-slate-200 bg-white px-4 py-5 sm:px-8">
        <button
          onClick={() => onNavigate("projects")}
          className="mb-3 inline-flex items-center gap-1 text-sm font-medium text-slate-500 hover:text-slate-900"
        >
          <ChevronLeft size={15} aria-hidden="true" /> Back to projects
        </button>
        <div className="flex flex-wrap items-start justify-between gap-3">
          <div className="min-w-0">
            <div className="mb-1.5 flex items-center gap-2">
              <StatusBadge status={project.status} />
              <span className="text-xs text-slate-400">{project.location}</span>
            </div>
            <h1 className="text-2xl font-bold tracking-tight text-slate-900">
              {project.name}
            </h1>
            <p className="mt-1 text-sm text-slate-500">
              Client: {project.client || "—"} · Contractor:{" "}
              {project.contractor || "—"}
            </p>
          </div>
          <button
            onClick={() => onNavigate("site-update")}
            className="inline-flex items-center gap-2 rounded bg-slate-900 px-4 py-2 text-sm font-semibold text-white transition-colors hover:bg-slate-700"
          >
            <FileEdit size={14} strokeWidth={2} aria-hidden="true" />
            Site update
          </button>
        </div>
        <div className="mt-4 flex items-center gap-4">
          <Progress value={project.progress} className="flex-1" />
          <span className="shrink-0 text-sm font-bold text-slate-900">
            {project.progress}% complete
          </span>
        </div>

        {/* Tabs */}
        <div
          className="-mb-px mt-5 flex gap-1 overflow-x-auto"
          role="tablist"
          aria-label="Project sections"
        >
          {TABS.map((t) => (
            <button
              key={t}
              role="tab"
              aria-selected={tab === t}
              onClick={() => setTab(t)}
              className={`whitespace-nowrap border-b-2 px-3.5 py-2.5 text-sm font-medium transition-colors ${
                tab === t
                  ? "border-slate-900 text-slate-900"
                  : "border-transparent text-slate-500 hover:text-slate-800"
              }`}
            >
              {t}
            </button>
          ))}
        </div>
      </div>

      <div className="px-4 py-6 sm:px-8">
        {error && (
          <div className="mb-4">
            <ErrorBanner message={error} />
          </div>
        )}

        {/* ── OVERVIEW ── */}
        {tab === "Overview" && (
          <div className="grid gap-4 xl:grid-cols-3">
            <div className="space-y-4 xl:col-span-2">
              <div className="rounded border border-slate-200 bg-white p-5">
                <h3 className="mb-4 border-b border-slate-100 pb-3 text-sm font-semibold text-slate-900">
                  Project details
                </h3>
                <dl className="grid grid-cols-1 gap-x-6 gap-y-4 sm:grid-cols-2">
                  {[
                    ["Project owner", project.client || "—"],
                    ["Main contractor", project.contractor || "—"],
                    ["Start date", fmtDate(project.startDate)],
                    [
                      "Expected completion",
                      fmtDate(project.expectedCompletion),
                    ],
                    ["Total budget", kes(project.budgetTotal)],
                    [
                      "Budget spent",
                      `${kes(project.budgetSpent)} (${spentPct}%)`,
                    ],
                  ].map(([label, value]) => (
                    <div key={label}>
                      <dt className="text-xs font-medium text-slate-400">
                        {label}
                      </dt>
                      <dd className="mt-0.5 text-sm font-medium text-slate-700">
                        {value}
                      </dd>
                    </div>
                  ))}
                </dl>
                {project.description && (
                  <div className="mt-4 border-t border-slate-100 pt-4">
                    <h4 className="text-xs font-medium text-slate-400">
                      Description
                    </h4>
                    <p className="mt-1.5 text-sm leading-relaxed text-slate-600">
                      {project.description}
                    </p>
                  </div>
                )}
              </div>

              {/* Recent tasks */}
              <div className="rounded border border-slate-200 bg-white">
                <div className="flex items-center justify-between border-b border-slate-100 px-5 py-4">
                  <h3 className="text-sm font-semibold text-slate-900">
                    Tasks
                  </h3>
                  <span className="text-xs text-slate-400">
                    {project.stats.tasksCompleted}/{project.stats.tasksTotal}{" "}
                    complete
                  </span>
                </div>
                {tasks.length === 0 ? (
                  <p className="px-5 py-8 text-center text-sm text-slate-500">
                    No tasks recorded for this project yet.
                  </p>
                ) : (
                  <ul className="divide-y divide-slate-50">
                    {tasks.slice(0, 5).map((task) => (
                      <li
                        key={task.id}
                        className="flex items-center gap-3 px-5 py-3"
                      >
                        <div className="min-w-0 flex-1">
                          <p
                            className={`truncate text-sm font-medium ${
                              task.status === "completed"
                                ? "text-slate-400 line-through"
                                : "text-slate-700"
                            }`}
                          >
                            {task.name}
                          </p>
                          <p className="mt-0.5 text-xs text-slate-400">
                            Due {fmtDate(task.due_date)}
                          </p>
                        </div>
                        <TaskBadge status={task.status} />
                      </li>
                    ))}
                  </ul>
                )}
              </div>

              {/* Budget chart */}
              {chartData.length > 0 && (
                <div className="rounded border border-slate-200 bg-white p-5">
                  <div className="mb-4 flex items-center justify-between">
                    <h3 className="text-sm font-semibold text-slate-900">
                      Budget — planned vs. actual (KES)
                    </h3>
                  </div>
                  <ResponsiveContainer width="100%" height={180}>
                    <AreaChart
                      data={chartData}
                      margin={{ top: 4, right: 4, left: 0, bottom: 0 }}
                    >
                      <defs>
                        <linearGradient
                          id="pd-planned"
                          x1="0"
                          y1="0"
                          x2="0"
                          y2="1"
                        >
                          <stop
                            offset="5%"
                            stopColor="#0f172a"
                            stopOpacity={0.12}
                          />
                          <stop
                            offset="95%"
                            stopColor="#0f172a"
                            stopOpacity={0}
                          />
                        </linearGradient>
                        <linearGradient
                          id="pd-actual"
                          x1="0"
                          y1="0"
                          x2="0"
                          y2="1"
                        >
                          <stop
                            offset="5%"
                            stopColor="#ea580c"
                            stopOpacity={0.12}
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
                      <Area
                        type="monotone"
                        dataKey="Planned"
                        stroke="#0f172a"
                        strokeWidth={2}
                        fill="url(#pd-planned)"
                        dot={false}
                      />
                      <Area
                        type="monotone"
                        dataKey="Actual"
                        stroke="#ea580c"
                        strokeWidth={2}
                        fill="url(#pd-actual)"
                        dot={false}
                      />
                    </AreaChart>
                  </ResponsiveContainer>
                </div>
              )}
            </div>

            {/* Sidebar */}
            <div className="space-y-4">
              <div className="rounded border border-slate-200 bg-white p-5">
                <h3 className="mb-3 text-sm font-semibold text-slate-900">
                  Budget summary
                </h3>
                <div className="mb-1 text-2xl font-bold text-slate-900">
                  {kesShort(project.budgetSpent)}
                </div>
                <div className="mb-3 text-xs text-slate-400">
                  spent of {kes(project.budgetTotal)}
                </div>
                <Progress value={spentPct} className="mb-2" />
                <p className="text-xs font-medium text-green-700">
                  {kesShort(project.budgetTotal - project.budgetSpent)}{" "}
                  remaining
                </p>
              </div>

              <div className="rounded border border-slate-200 bg-white">
                <div className="flex items-center gap-2 border-b border-slate-100 px-4 py-3.5">
                  <AlertTriangle
                    size={14}
                    className="text-orange-600"
                    aria-hidden="true"
                  />
                  <h3 className="text-sm font-semibold text-slate-900">
                    Material alerts
                  </h3>
                </div>
                {materials.filter(
                  (m) => m.status === "critical" || m.status === "low",
                ).length === 0 ? (
                  <p className="px-4 py-6 text-center text-xs text-slate-500">
                    All stock levels healthy.
                  </p>
                ) : (
                  <ul className="divide-y divide-slate-50">
                    {materials
                      .filter(
                        (m) => m.status === "critical" || m.status === "low",
                      )
                      .slice(0, 4)
                      .map((m) => (
                        <li key={m.id} className="px-4 py-3">
                          <div className="flex items-center justify-between gap-2">
                            <span className="min-w-0 truncate text-xs font-medium text-slate-700">
                              {m.name}
                            </span>
                            <MaterialBadge status={m.status} />
                          </div>
                          <p className="mt-0.5 text-xs text-slate-400">
                            {m.available_qty} {m.unit} available
                          </p>
                        </li>
                      ))}
                  </ul>
                )}
              </div>

              <div className="rounded border border-slate-200 bg-white">
                <div className="border-b border-slate-100 px-4 py-3.5">
                  <h3 className="text-sm font-semibold text-slate-900">
                    Recent documents
                  </h3>
                </div>
                {documents.length === 0 ? (
                  <p className="px-4 py-6 text-center text-xs text-slate-500">
                    No documents uploaded.
                  </p>
                ) : (
                  <ul className="divide-y divide-slate-50">
                    {documents.slice(0, 4).map((doc) => (
                      <li key={doc.id} className="px-4 py-2.5">
                        <p className="text-xs font-medium leading-snug text-slate-700">
                          {doc.name}
                        </p>
                        <p className="mt-0.5 text-xs text-slate-400">
                          {doc.category} · {fmtDate(doc.created_at)}
                        </p>
                      </li>
                    ))}
                  </ul>
                )}
              </div>

              <div className="rounded border border-slate-200 bg-white">
                <div className="flex items-center gap-2 border-b border-slate-100 px-4 py-3.5">
                  <User
                    size={14}
                    className="text-slate-500"
                    aria-hidden="true"
                  />
                  <h3 className="text-sm font-semibold text-slate-900">
                    Project team
                  </h3>
                </div>
                {project.team.length === 0 ? (
                  <p className="px-4 py-6 text-center text-xs text-slate-500">
                    No members assigned.
                  </p>
                ) : (
                  <ul className="divide-y divide-slate-50">
                    {project.team.map((m) => (
                      <li
                        key={m.id}
                        className="flex items-center gap-3 px-4 py-2.5"
                      >
                        <span
                          className="flex h-7 w-7 shrink-0 items-center justify-center rounded-full text-[10px] font-bold text-white"
                          style={{ background: m.color }}
                          aria-hidden="true"
                        >
                          {m.initials}
                        </span>
                        <div className="min-w-0">
                          <p className="truncate text-xs font-semibold text-slate-700">
                            {m.name}
                          </p>
                          <p className="truncate text-xs text-slate-400">
                            {m.roleLabel || m.role}
                          </p>
                        </div>
                      </li>
                    ))}
                  </ul>
                )}
              </div>
            </div>
          </div>
        )}

        {/* ── TASKS ── */}
        {tab === "Tasks" && (
          <div className="overflow-hidden rounded border border-slate-200 bg-white">
            <div className="overflow-x-auto">
              <table className="w-full">
                <thead>
                  <tr className="border-b border-slate-100 bg-slate-50">
                    {["Task", "Assignee", "Due date", "Status", "Priority"].map(
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
                  {tasks.map((task) => (
                    <tr
                      key={task.id}
                      className="border-b border-slate-50 last:border-0"
                    >
                      <td className="px-4 py-3">
                        <div
                          className={`text-sm font-medium ${
                            task.status === "completed"
                              ? "text-slate-400 line-through"
                              : "text-slate-700"
                          }`}
                        >
                          {task.name}
                        </div>
                        {task.description && (
                          <div className="mt-0.5 max-w-md truncate text-xs text-slate-400">
                            {task.description}
                          </div>
                        )}
                      </td>
                      <td className="px-4 py-3 text-xs text-slate-600">
                        {task.assignee_name ?? "—"}
                      </td>
                      <td className="px-4 py-3 text-xs text-slate-600">
                        {fmtDate(task.due_date)}
                      </td>
                      <td className="px-4 py-3">
                        <select
                          value={task.status}
                          disabled={updatingTask === task.id}
                          onChange={(e) =>
                            setTaskStatus(
                              task,
                              e.target.value as Task["status"],
                            )
                          }
                          aria-label={`Change status of ${task.name}`}
                          className="rounded border border-slate-300 bg-white px-2 py-1 text-xs text-slate-700 outline-none focus:border-slate-800 disabled:opacity-50"
                        >
                          <option value="not-started">Not started</option>
                          <option value="in-progress">In progress</option>
                          <option value="blocked">Blocked</option>
                          <option value="completed">Completed</option>
                        </select>
                      </td>
                      <td className="px-4 py-3">
                        <span
                          className={`text-xs font-semibold capitalize ${
                            task.priority === "high"
                              ? "text-red-700"
                              : task.priority === "medium"
                                ? "text-orange-700"
                                : "text-slate-500"
                          }`}
                        >
                          {task.priority}
                        </span>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
            {tasks.length === 0 && (
              <p className="py-12 text-center text-sm text-slate-500">
                No tasks recorded for this project yet.
              </p>
            )}
          </div>
        )}

        {/* ── MATERIALS ── */}
        {tab === "Materials" && (
          <div className="overflow-hidden rounded border border-slate-200 bg-white">
            <div className="overflow-x-auto">
              <table className="w-full">
                <thead>
                  <tr className="border-b border-slate-100 bg-slate-50">
                    {[
                      "Material",
                      "Unit",
                      "Required",
                      "Available",
                      "Used",
                      "Supplier",
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
                  {materials.map((m) => (
                    <tr
                      key={m.id}
                      className="border-b border-slate-50 last:border-0"
                    >
                      <td className="px-4 py-3 text-sm font-medium text-slate-700">
                        {m.name}
                      </td>
                      <td className="px-4 py-3 text-xs text-slate-500">
                        {m.unit}
                      </td>
                      <td className="px-4 py-3 text-sm text-slate-600">
                        {m.required_qty.toLocaleString()}
                      </td>
                      <td
                        className={`px-4 py-3 text-sm font-semibold ${
                          m.status === "critical"
                            ? "text-red-700"
                            : m.status === "low"
                              ? "text-orange-700"
                              : "text-slate-700"
                        }`}
                      >
                        {m.available_qty.toLocaleString()}
                      </td>
                      <td className="px-4 py-3 text-sm text-slate-600">
                        {m.used_qty.toLocaleString()}
                      </td>
                      <td className="px-4 py-3 text-xs text-slate-500">
                        {m.supplier_name ?? "—"}
                      </td>
                      <td className="px-4 py-3">
                        <MaterialBadge status={m.status} />
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
            {materials.length === 0 && (
              <p className="py-12 text-center text-sm text-slate-500">
                No materials tracked for this project.
              </p>
            )}
          </div>
        )}

        {/* ── BUDGET ── */}
        {tab === "Budget" && (
          <div className="max-w-4xl space-y-4">
            <div className="grid grid-cols-1 gap-4 sm:grid-cols-3">
              {[
                {
                  label: "Total budget",
                  value: kes(project.budgetTotal),
                  sub: "Original contract value",
                },
                {
                  label: "Amount spent",
                  value: kes(project.budgetSpent),
                  sub: `${spentPct}% of total budget`,
                },
                {
                  label: "Remaining",
                  value: kes(project.budgetTotal - project.budgetSpent),
                  sub: `${100 - spentPct}% remaining`,
                },
              ].map((s) => (
                <div
                  key={s.label}
                  className="rounded border border-slate-200 bg-white p-5"
                >
                  <div className="mb-2 text-xs font-semibold uppercase tracking-wide text-slate-400">
                    {s.label}
                  </div>
                  <div className="mb-1 text-xl font-bold text-slate-900">
                    {s.value}
                  </div>
                  <div className="text-xs text-slate-400">{s.sub}</div>
                </div>
              ))}
            </div>
            <div className="rounded border border-slate-200 bg-white p-5">
              <h3 className="mb-4 text-sm font-semibold text-slate-900">
                Monthly budget: planned vs. actual (KES)
              </h3>
              {chartData.length === 0 ? (
                <p className="py-10 text-center text-sm text-slate-500">
                  No budget data recorded yet.
                </p>
              ) : (
                <ResponsiveContainer width="100%" height={240}>
                  <AreaChart
                    data={chartData}
                    margin={{ top: 4, right: 4, left: 0, bottom: 0 }}
                  >
                    <defs>
                      <linearGradient id="pd-p2" x1="0" y1="0" x2="0" y2="1">
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
                      <linearGradient id="pd-a2" x1="0" y1="0" x2="0" y2="1">
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
                    <Area
                      type="monotone"
                      dataKey="Planned"
                      stroke="#0f172a"
                      strokeWidth={2}
                      fill="url(#pd-p2)"
                      dot={false}
                    />
                    <Area
                      type="monotone"
                      dataKey="Actual"
                      stroke="#ea580c"
                      strokeWidth={2}
                      fill="url(#pd-a2)"
                      dot={false}
                    />
                  </AreaChart>
                </ResponsiveContainer>
              )}
            </div>
          </div>
        )}

        {/* ── DOCUMENTS ── */}
        {tab === "Documents" && (
          <div className="overflow-hidden rounded border border-slate-200 bg-white">
            <div className="overflow-x-auto">
              <table className="w-full">
                <thead>
                  <tr className="border-b border-slate-100 bg-slate-50">
                    {["Name", "Category", "Version", "Uploaded by", "Date"].map(
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
                  {documents.map((doc) => (
                    <tr
                      key={doc.id}
                      className="border-b border-slate-50 last:border-0"
                    >
                      <td className="px-4 py-3 text-sm font-medium text-slate-700">
                        {doc.name}
                      </td>
                      <td className="px-4 py-3">
                        <span className="rounded bg-slate-100 px-2 py-0.5 text-xs text-slate-600">
                          {doc.category}
                        </span>
                      </td>
                      <td className="px-4 py-3 text-xs font-medium text-slate-500">
                        {doc.version}
                      </td>
                      <td className="px-4 py-3 text-xs text-slate-500">
                        {doc.uploaded_by_name ?? "—"}
                      </td>
                      <td className="px-4 py-3 text-xs text-slate-500">
                        {fmtDate(doc.created_at)}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
            {documents.length === 0 && (
              <p className="py-12 text-center text-sm text-slate-500">
                No documents uploaded for this project.
              </p>
            )}
          </div>
        )}

        {/* ── TEAM ── */}
        {tab === "Team" && (
          <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
            {project.team.map((m) => (
              <div
                key={m.id}
                className="rounded border border-slate-200 bg-white p-5"
              >
                <div className="mb-3 flex items-center gap-3">
                  <span
                    className="flex h-11 w-11 shrink-0 items-center justify-center rounded-full text-sm font-bold text-white"
                    style={{ background: m.color }}
                    aria-hidden="true"
                  >
                    {m.initials}
                  </span>
                  <div className="min-w-0">
                    <div className="truncate text-sm font-semibold text-slate-900">
                      {m.name}
                    </div>
                    <div className="truncate text-xs text-slate-500">
                      {m.roleLabel || m.role}
                    </div>
                  </div>
                </div>
                <div className="space-y-1 border-t border-slate-100 pt-3 text-xs text-slate-500">
                  <div className="truncate">{m.email}</div>
                  {m.phone && <div>{m.phone}</div>}
                </div>
              </div>
            ))}
            {project.team.length === 0 && (
              <div className="col-span-full rounded border border-slate-200 bg-white py-12 text-center text-sm text-slate-500">
                <Users
                  size={20}
                  className="mx-auto mb-2 text-slate-300"
                  aria-hidden="true"
                />
                No members assigned to this project yet.
              </div>
            )}
          </div>
        )}

        {/* ── ACTIVITY ── */}
        {tab === "Activity" && (
          <div className="max-w-2xl rounded border border-slate-200 bg-white">
            {activity.length === 0 ? (
              <p className="py-12 text-center text-sm text-slate-500">
                No activity recorded for this project yet.
              </p>
            ) : (
              <ul className="divide-y divide-slate-50">
                {activity.map((item) => (
                  <li key={item.id} className="px-5 py-3.5">
                    <p className="text-sm text-slate-700">{item.text}</p>
                    <p className="mt-1 text-xs text-slate-400">
                      {item.actor_name} · {relTime(item.created_at)}
                    </p>
                  </li>
                ))}
              </ul>
            )}
          </div>
        )}
      </div>
    </div>
  )
}
