import { useCallback, useEffect, useState } from "react"
import { Clock, Plus, Search } from "lucide-react"
import type { NavigateFn } from "../App"
import { api, ApiRequestError } from "../lib/api"
import type { Project, Task } from "../lib/api"
import { fmtDate, isOverdue } from "../lib/format"
import { useAuth } from "../lib/auth"
import {
  Button,
  ErrorBanner,
  LoadingBlock,
  TaskBadge,
  inputClass,
  labelClass,
} from "../components/ui"

interface Props {
  onNavigate: NavigateFn
}

const STATUSES = ["not-started", "in-progress", "blocked", "completed"] as const
type Status = typeof STATUSES[number]

const STATUS_LABEL: Record<Status, string> = {
  "not-started": "Not started",
  "in-progress": "In progress",
  blocked: "Blocked",
  completed: "Completed",
}

const priorityColor: Record<string, string> = {
  high: "text-red-700",
  medium: "text-orange-700",
  low: "text-slate-500",
}

function NewTaskModal({
  projects,
  onClose,
  onCreated,
}: {
  projects: Project[]
  onClose: () => void
  onCreated: () => void
}) {
  const [form, setForm] = useState({
    projectId: String(projects[0]?.id ?? ""),
    name: "",
    description: "",
    assigneeId: "",
    startDate: "",
    dueDate: "",
    priority: "medium",
  })
  const [team, setTeam] = useState<Array<{ id: number; name: string }>>([])
  const [error, setError] = useState("")
  const [saving, setSaving] = useState(false)

  useEffect(() => {
    const pid = Number(form.projectId)
    if (!pid) {
      setTeam([])
      return
    }
    api
      .get<{ project: Project }>(`/api/projects/${pid}`)
      .then((d) => setTeam(d.project.team))
      .catch(() => setTeam([]))
  }, [form.projectId])

  const set = (field: string, value: string) =>
    setForm((f) => ({ ...f, [field]: value }))

  const submit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError("")
    if (!form.name.trim()) {
      setError("Task name is required.")
      return
    }
    if (!form.dueDate) {
      setError("Due date is required.")
      return
    }
    setSaving(true)
    try {
      await api.post("/api/tasks", {
        projectId: Number(form.projectId),
        name: form.name,
        description: form.description,
        assigneeId: form.assigneeId ? Number(form.assigneeId) : null,
        startDate: form.startDate || null,
        dueDate: form.dueDate,
        priority: form.priority,
      })
      onCreated()
    } catch (err) {
      setError(
        err instanceof ApiRequestError
          ? err.message
          : "Could not create the task.",
      )
      setSaving(false)
    }
  }

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-slate-900/40 p-4"
      role="dialog"
      aria-modal="true"
      aria-labelledby="nt-title"
    >
      <div className="max-h-[90vh] w-full max-w-lg overflow-y-auto rounded border border-slate-200 bg-white shadow-xl">
        <div className="border-b border-slate-100 px-6 py-4">
          <h2 id="nt-title" className="text-base font-semibold text-slate-900">
            New task
          </h2>
        </div>
        <form onSubmit={submit} noValidate>
          <div className="space-y-4 px-6 py-5">
            {error && (
              <p
                className="rounded border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-800"
                role="alert"
              >
                {error}
              </p>
            )}
            <div>
              <label htmlFor="nt-project" className={labelClass}>
                Project{" "}
                <span className="text-red-600" aria-hidden="true">
                  *
                </span>
              </label>
              <select
                id="nt-project"
                className={inputClass}
                value={form.projectId}
                onChange={(e) => set("projectId", e.target.value)}
              >
                {projects.map((p) => (
                  <option key={p.id} value={p.id}>
                    {p.name}
                  </option>
                ))}
              </select>
            </div>
            <div>
              <label htmlFor="nt-name" className={labelClass}>
                Task name{" "}
                <span className="text-red-600" aria-hidden="true">
                  *
                </span>
              </label>
              <input
                id="nt-name"
                className={inputClass}
                value={form.name}
                onChange={(e) => set("name", e.target.value)}
                autoFocus
              />
            </div>
            <div>
              <label htmlFor="nt-desc" className={labelClass}>
                Description
              </label>
              <textarea
                id="nt-desc"
                rows={3}
                className={inputClass}
                value={form.description}
                onChange={(e) => set("description", e.target.value)}
              />
            </div>
            <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
              <div>
                <label htmlFor="nt-assignee" className={labelClass}>
                  Assignee
                </label>
                <select
                  id="nt-assignee"
                  className={inputClass}
                  value={form.assigneeId}
                  onChange={(e) => set("assigneeId", e.target.value)}
                >
                  <option value="">Unassigned</option>
                  {team.map((m) => (
                    <option key={m.id} value={m.id}>
                      {m.name}
                    </option>
                  ))}
                </select>
              </div>
              <div>
                <label htmlFor="nt-priority" className={labelClass}>
                  Priority
                </label>
                <select
                  id="nt-priority"
                  className={inputClass}
                  value={form.priority}
                  onChange={(e) => set("priority", e.target.value)}
                >
                  <option value="high">High</option>
                  <option value="medium">Medium</option>
                  <option value="low">Low</option>
                </select>
              </div>
              <div>
                <label htmlFor="nt-start" className={labelClass}>
                  Start date
                </label>
                <input
                  id="nt-start"
                  type="date"
                  className={inputClass}
                  value={form.startDate}
                  onChange={(e) => set("startDate", e.target.value)}
                />
              </div>
              <div>
                <label htmlFor="nt-due" className={labelClass}>
                  Due date{" "}
                  <span className="text-red-600" aria-hidden="true">
                    *
                  </span>
                </label>
                <input
                  id="nt-due"
                  type="date"
                  required
                  className={inputClass}
                  value={form.dueDate}
                  onChange={(e) => set("dueDate", e.target.value)}
                />
              </div>
            </div>
          </div>
          <div className="flex justify-end gap-2 border-t border-slate-100 px-6 py-4">
            <Button type="button" variant="secondary" onClick={onClose}>
              Cancel
            </Button>
            <Button type="submit" loading={saving}>
              {saving ? "Creating…" : "Create task"}
            </Button>
          </div>
        </form>
      </div>
    </div>
  )
}

export default function TasksPage(_props: Props) {
  const { user } = useAuth()
  const [tasks, setTasks] = useState<Task[]>([])
  const [projects, setProjects] = useState<Project[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState("")
  const [search, setSearch] = useState("")
  const [filterStatus, setFilterStatus] = useState<string>("All")
  const [filterProject, setFilterProject] = useState<string>("All")
  const [view, setView] = useState<"list" | "kanban">("list")
  const [showNew, setShowNew] = useState(false)
  const [updatingId, setUpdatingId] = useState<number | null>(null)

  const load = useCallback(async () => {
    setLoading(true)
    setError("")
    try {
      const [t, p] = await Promise.all([
        api.get<{ tasks: Task[] }>("/api/tasks"),
        api.get<{ projects: Project[] }>("/api/projects"),
      ])
      setTasks(t.tasks)
      setProjects(p.projects)
    } catch (err) {
      setError(err instanceof Error ? err.message : "Could not load tasks.")
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    void load()
  }, [load])

  const setTaskStatus = async (task: Task, status: Task["status"]) => {
    if (status === task.status) return
    setUpdatingId(task.id)
    try {
      await api.patch(`/api/tasks/${task.id}`, { status })
      setTasks((prev) =>
        prev.map((t) => (t.id === task.id ? { ...t, status } : t)),
      )
    } catch (err) {
      setError(
        err instanceof ApiRequestError
          ? err.message
          : "Could not update the task.",
      )
      await load()
    } finally {
      setUpdatingId(null)
    }
  }

  const filtered = tasks.filter((t) => {
    const q = search.toLowerCase()
    return (
      t.name.toLowerCase().includes(q) &&
      (filterStatus === "All" || t.status === filterStatus) &&
      (filterProject === "All" || t.project_id === Number(filterProject))
    )
  })

  const canEdit = user?.role !== "viewer"

  return (
    <div>
      <div className="border-b border-slate-200 bg-white px-4 py-5 sm:px-8">
        <div className="mb-4 flex flex-wrap items-center justify-between gap-3">
          <div>
            <h1 className="text-xl font-bold tracking-tight text-slate-900">
              Tasks
            </h1>
            <p className="mt-0.5 text-sm text-slate-500">
              {tasks.filter((t) => t.status !== "completed").length} open ·{" "}
              {tasks.filter((t) => t.status === "blocked").length} blocked ·{" "}
              {tasks.filter((t) => t.status === "completed").length} completed
            </p>
          </div>
          {canEdit && (
            <Button onClick={() => setShowNew(true)}>
              <Plus size={15} strokeWidth={2.5} aria-hidden="true" />
              New task
            </Button>
          )}
        </div>

        <div className="flex flex-wrap items-center gap-3">
          <div className="relative">
            <Search
              size={14}
              className="pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 text-slate-400"
              aria-hidden="true"
            />
            <input
              type="search"
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              placeholder="Search tasks…"
              aria-label="Search tasks"
              className="w-52 rounded border border-slate-300 bg-white py-2 pl-9 pr-4 text-sm outline-none focus:border-slate-800 focus:ring-2 focus:ring-slate-900/10"
            />
          </div>
          <select
            value={filterStatus}
            onChange={(e) => setFilterStatus(e.target.value)}
            aria-label="Filter by status"
            className="rounded border border-slate-300 bg-white px-3 py-2 text-sm outline-none focus:border-slate-800"
          >
            <option value="All">All statuses</option>
            {STATUSES.map((s) => (
              <option key={s} value={s}>
                {STATUS_LABEL[s]}
              </option>
            ))}
          </select>
          <select
            value={filterProject}
            onChange={(e) => setFilterProject(e.target.value)}
            aria-label="Filter by project"
            className="rounded border border-slate-300 bg-white px-3 py-2 text-sm outline-none focus:border-slate-800"
          >
            <option value="All">All projects</option>
            {projects.map((p) => (
              <option key={p.id} value={p.id}>
                {p.name}
              </option>
            ))}
          </select>

          <div
            className="ml-auto flex rounded bg-slate-100 p-0.5"
            role="group"
            aria-label="View mode"
          >
            {(["list", "kanban"] as const).map((v) => (
              <button
                key={v}
                onClick={() => setView(v)}
                aria-pressed={view === v}
                className={`rounded px-3 py-1.5 text-xs font-semibold capitalize transition-colors ${
                  view === v
                    ? "bg-white text-slate-900 shadow-sm"
                    : "text-slate-500 hover:text-slate-700"
                }`}
              >
                {v}
              </button>
            ))}
          </div>
        </div>
      </div>

      <div className="px-4 py-6 sm:px-8">
        {error && (
          <div className="mb-4">
            <ErrorBanner message={error} />
          </div>
        )}

        {loading ? (
          <div className="rounded border border-slate-200 bg-white">
            <LoadingBlock label="Loading tasks…" />
          </div>
        ) : view === "list" ? (
          <div className="overflow-hidden rounded border border-slate-200 bg-white">
            <div className="overflow-x-auto">
              <table className="w-full">
                <thead>
                  <tr className="border-b border-slate-100 bg-slate-50">
                    {[
                      "Task",
                      "Project",
                      "Assignee",
                      "Due",
                      "Status",
                      "Priority",
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
                  {filtered.map((task) => {
                    const overdue = isOverdue(task)
                    return (
                      <tr
                        key={task.id}
                        className="border-b border-slate-50 last:border-0"
                      >
                        <td className="px-4 py-3" style={{ maxWidth: 360 }}>
                          <div
                            className={`text-sm font-medium ${
                              task.status === "completed"
                                ? "text-slate-400 line-through"
                                : "text-slate-700"
                            }`}
                          >
                            {task.name}
                          </div>
                        </td>
                        <td className="px-4 py-3 text-xs text-slate-500">
                          {task.project_name ?? "—"}
                        </td>
                        <td className="px-4 py-3">
                          {task.assignee_name ? (
                            <div className="flex items-center gap-2">
                              <span
                                className="flex h-5.5 w-5.5 items-center justify-center rounded-full text-[8px] font-bold text-white"
                                style={{
                                  width: 22,
                                  height: 22,
                                  background: task.assignee_color ?? "#334155",
                                }}
                                aria-hidden="true"
                              >
                                {task.assignee_initials}
                              </span>
                              <span className="text-xs text-slate-600">
                                {task.assignee_name}
                              </span>
                            </div>
                          ) : (
                            <span className="text-xs text-slate-400">
                              Unassigned
                            </span>
                          )}
                        </td>
                        <td className="px-4 py-3">
                          <span
                            className={`text-xs font-medium ${
                              overdue ? "text-red-700" : "text-slate-500"
                            }`}
                          >
                            {overdue && "⚠ "}
                            {fmtDate(task.due_date)}
                          </span>
                        </td>
                        <td className="px-4 py-3">
                          {canEdit ? (
                            <select
                              value={task.status}
                              disabled={updatingId === task.id}
                              onChange={(e) =>
                                setTaskStatus(
                                  task,
                                  e.target.value as Task["status"],
                                )
                              }
                              aria-label={`Change status of ${task.name}`}
                              className="rounded border border-slate-300 bg-white px-2 py-1 text-xs text-slate-700 outline-none focus:border-slate-800 disabled:opacity-50"
                            >
                              {STATUSES.map((s) => (
                                <option key={s} value={s}>
                                  {STATUS_LABEL[s]}
                                </option>
                              ))}
                            </select>
                          ) : (
                            <TaskBadge status={task.status} />
                          )}
                        </td>
                        <td className="px-4 py-3">
                          <span
                            className={`text-xs font-semibold capitalize ${priorityColor[task.priority]}`}
                          >
                            {task.priority}
                          </span>
                        </td>
                      </tr>
                    )
                  })}
                </tbody>
              </table>
            </div>
            {filtered.length === 0 && (
              <p className="py-12 text-center text-sm text-slate-500">
                No tasks match your filters.
              </p>
            )}
          </div>
        ) : (
          <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
            {STATUSES.map((status) => {
              const colTasks = filtered.filter((t) => t.status === status)
              return (
                <section key={status} aria-label={STATUS_LABEL[status]}>
                  <div className="mb-3 flex items-center gap-2">
                    <h2 className="text-sm font-semibold text-slate-700">
                      {STATUS_LABEL[status]}
                    </h2>
                    <span className="ml-auto rounded bg-slate-100 px-1.5 py-0.5 text-xs font-medium text-slate-500">
                      {colTasks.length}
                    </span>
                  </div>
                  <div className="space-y-2">
                    {colTasks.map((task) => {
                      const overdue = isOverdue(task)
                      return (
                        <div
                          key={task.id}
                          className={`rounded border bg-white p-3 ${
                            overdue ? "border-red-200" : "border-slate-200"
                          }`}
                        >
                          <div className="mb-1.5 text-sm font-medium leading-snug text-slate-700">
                            {task.name}
                          </div>
                          <div className="mb-2 text-xs text-slate-400">
                            {task.project_name}
                          </div>
                          <div className="flex items-center justify-between">
                            <span className="text-xs text-slate-500">
                              {task.assignee_name ?? "Unassigned"}
                            </span>
                            <span
                              className={`flex items-center gap-1 text-xs ${
                                overdue
                                  ? "font-semibold text-red-700"
                                  : "text-slate-400"
                              }`}
                            >
                              <Clock size={11} aria-hidden="true" />
                              {fmtDate(task.due_date)}
                            </span>
                          </div>
                        </div>
                      )
                    })}
                    {colTasks.length === 0 && (
                      <div className="rounded border border-dashed border-slate-200 py-7 text-center text-xs text-slate-400">
                        No tasks
                      </div>
                    )}
                  </div>
                </section>
              )
            })}
          </div>
        )}
      </div>

      {showNew && (
        <NewTaskModal
          projects={projects}
          onClose={() => setShowNew(false)}
          onCreated={() => {
            setShowNew(false)
            void load()
          }}
        />
      )}
    </div>
  )
}
