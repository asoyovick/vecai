import { useCallback, useEffect, useState } from "react"
import { MapPin, Plus, Search } from "lucide-react"
import type { NavigateFn } from "../App"
import { api, ApiRequestError } from "../lib/api"
import type { Project } from "../lib/api"
import { fmtDate, kes } from "../lib/format"
import { useAuth } from "../lib/auth"
import {
  ErrorBanner,
  LoadingBlock,
  Progress,
  StatusBadge,
  inputClass,
  labelClass,
  Button,
} from "../components/ui"

interface Props {
  onNavigate: NavigateFn
}

const STATUSES = [
  "Tendering",
  "Design",
  "Foundation",
  "Construction",
  "Roofing",
  "Finishing",
  "Handover",
  "Completed",
]

function NewProjectModal({
  onClose,
  onCreated,
}: {
  onClose: () => void
  onCreated: (p: Project) => void
}) {
  const [form, setForm] = useState({
    name: "",
    location: "",
    client: "",
    contractor: "",
    status: "Tendering",
    startDate: "",
    expectedCompletion: "",
    budgetTotal: "",
    description: "",
  })
  const [error, setError] = useState("")
  const [saving, setSaving] = useState(false)

  const set = (field: string, value: string) =>
    setForm((f) => ({ ...f, [field]: value }))

  const submit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError("")
    if (!form.name.trim()) {
      setError("Project name is required.")
      return
    }
    setSaving(true)
    try {
      const data = await api.post<{ project: Project }>("/api/projects", {
        name: form.name,
        location: form.location,
        client: form.client,
        contractor: form.contractor,
        status: form.status,
        startDate: form.startDate || null,
        expectedCompletion: form.expectedCompletion || null,
        budgetTotal: form.budgetTotal ? Number(form.budgetTotal) : 0,
        description: form.description,
      })
      onCreated(data.project)
    } catch (err) {
      setError(
        err instanceof ApiRequestError
          ? err.message
          : "Could not create the project.",
      )
      setSaving(false)
    }
  }

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-slate-900/40 p-4"
      role="dialog"
      aria-modal="true"
      aria-labelledby="np-title"
    >
      <div className="max-h-[90vh] w-full max-w-lg overflow-y-auto rounded border border-slate-200 bg-white shadow-xl">
        <div className="border-b border-slate-100 px-6 py-4">
          <h2 id="np-title" className="text-base font-semibold text-slate-900">
            New project
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
              <label htmlFor="np-name" className={labelClass}>
                Project name{" "}
                <span className="text-red-600" aria-hidden="true">
                  *
                </span>
              </label>
              <input
                id="np-name"
                className={inputClass}
                value={form.name}
                onChange={(e) => set("name", e.target.value)}
                required
                autoFocus
              />
            </div>
            <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
              <div>
                <label htmlFor="np-loc" className={labelClass}>
                  Location
                </label>
                <input
                  id="np-loc"
                  className={inputClass}
                  value={form.location}
                  onChange={(e) => set("location", e.target.value)}
                />
              </div>
              <div>
                <label htmlFor="np-status" className={labelClass}>
                  Status
                </label>
                <select
                  id="np-status"
                  className={inputClass}
                  value={form.status}
                  onChange={(e) => set("status", e.target.value)}
                >
                  {STATUSES.map((s) => (
                    <option key={s}>{s}</option>
                  ))}
                </select>
              </div>
              <div>
                <label htmlFor="np-client" className={labelClass}>
                  Client / owner
                </label>
                <input
                  id="np-client"
                  className={inputClass}
                  value={form.client}
                  onChange={(e) => set("client", e.target.value)}
                />
              </div>
              <div>
                <label htmlFor="np-con" className={labelClass}>
                  Main contractor
                </label>
                <input
                  id="np-con"
                  className={inputClass}
                  value={form.contractor}
                  onChange={(e) => set("contractor", e.target.value)}
                />
              </div>
              <div>
                <label htmlFor="np-start" className={labelClass}>
                  Start date
                </label>
                <input
                  id="np-start"
                  type="date"
                  className={inputClass}
                  value={form.startDate}
                  onChange={(e) => set("startDate", e.target.value)}
                />
              </div>
              <div>
                <label htmlFor="np-end" className={labelClass}>
                  Expected completion
                </label>
                <input
                  id="np-end"
                  type="date"
                  className={inputClass}
                  value={form.expectedCompletion}
                  onChange={(e) => set("expectedCompletion", e.target.value)}
                />
              </div>
            </div>
            <div>
              <label htmlFor="np-budget" className={labelClass}>
                Total budget (KES)
              </label>
              <input
                id="np-budget"
                type="number"
                min={0}
                className={inputClass}
                value={form.budgetTotal}
                onChange={(e) => set("budgetTotal", e.target.value)}
                placeholder="e.g. 12500000"
              />
            </div>
            <div>
              <label htmlFor="np-desc" className={labelClass}>
                Description
              </label>
              <textarea
                id="np-desc"
                rows={3}
                className={inputClass}
                value={form.description}
                onChange={(e) => set("description", e.target.value)}
              />
            </div>
          </div>
          <div className="flex justify-end gap-2 border-t border-slate-100 px-6 py-4">
            <Button type="button" variant="secondary" onClick={onClose}>
              Cancel
            </Button>
            <Button type="submit" loading={saving}>
              {saving ? "Creating…" : "Create project"}
            </Button>
          </div>
        </form>
      </div>
    </div>
  )
}

export default function ProjectsPage({ onNavigate }: Props) {
  const { user } = useAuth()
  const [projects, setProjects] = useState<Project[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState("")
  const [search, setSearch] = useState("")
  const [filter, setFilter] = useState("All")
  const [showNew, setShowNew] = useState(false)

  const load = useCallback(async () => {
    setLoading(true)
    setError("")
    try {
      const data = await api.get<{ projects: Project[] }>("/api/projects")
      setProjects(data.projects)
    } catch (err) {
      setError(err instanceof Error ? err.message : "Could not load projects.")
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    void load()
  }, [load])

  const statuses = [
    "All",
    ...Array.from(new Set(projects.map((p) => p.status))),
  ]
  const filtered = projects.filter((p) => {
    const q = search.toLowerCase()
    const matchSearch =
      p.name.toLowerCase().includes(q) || p.location.toLowerCase().includes(q)
    return matchSearch && (filter === "All" || p.status === filter)
  })

  return (
    <div>
      <div className="border-b border-slate-200 bg-white px-4 py-5 sm:px-8">
        <div className="flex flex-wrap items-center justify-between gap-3">
          <div>
            <h1 className="text-xl font-bold tracking-tight text-slate-900">
              Projects
            </h1>
            <p className="mt-0.5 text-sm text-slate-500">
              {projects.length} {projects.length === 1 ? "project" : "projects"}{" "}
              · {projects.filter((p) => p.status !== "Completed").length} active
            </p>
          </div>
          {user?.role === "admin" && (
            <button
              onClick={() => setShowNew(true)}
              className="inline-flex items-center gap-2 rounded bg-slate-900 px-4 py-2 text-sm font-semibold text-white transition-colors hover:bg-slate-700 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-slate-800"
            >
              <Plus size={15} strokeWidth={2.5} aria-hidden="true" />
              New project
            </button>
          )}
        </div>

        <div className="mt-4 flex flex-wrap items-center gap-3">
          <div className="relative">
            <Search
              size={15}
              className="pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 text-slate-400"
              aria-hidden="true"
            />
            <input
              type="search"
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              placeholder="Search projects…"
              aria-label="Search projects"
              className="w-56 rounded border border-slate-300 bg-white py-2 pl-9 pr-4 text-sm text-slate-900 outline-none transition-colors placeholder:text-slate-400 focus:border-slate-800 focus:ring-2 focus:ring-slate-900/10 sm:w-64"
            />
          </div>
          <div
            className="flex flex-wrap gap-1.5"
            role="group"
            aria-label="Filter by status"
          >
            {statuses.map((s) => (
              <button
                key={s}
                onClick={() => setFilter(s)}
                aria-pressed={filter === s}
                className={`rounded px-3 py-1.5 text-sm font-medium transition-colors ${
                  filter === s
                    ? "bg-slate-900 text-white"
                    : "border border-slate-300 bg-white text-slate-600 hover:bg-slate-50"
                }`}
              >
                {s}
              </button>
            ))}
          </div>
        </div>
      </div>

      <div className="px-4 py-6 sm:px-8">
        {loading ? (
          <div className="rounded border border-slate-200 bg-white">
            <LoadingBlock label="Loading projects…" />
          </div>
        ) : error ? (
          <ErrorBanner message={error} onRetry={load} />
        ) : (
          <div className="overflow-hidden rounded border border-slate-200 bg-white">
            <div className="overflow-x-auto">
              <table className="w-full">
                <thead>
                  <tr className="border-b border-slate-100 bg-slate-50">
                    {[
                      "Project",
                      "Status",
                      "Progress",
                      "Budget",
                      "Completion",
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
                  {filtered.map((project) => (
                    <tr
                      key={project.id}
                      onClick={() => onNavigate("project-detail", project.id)}
                      className="cursor-pointer border-b border-slate-50 transition-colors last:border-0 hover:bg-slate-50 focus-visible:outline-2 focus-visible:-outline-offset-2 focus-visible:outline-slate-800"
                      tabIndex={0}
                      onKeyDown={(e) =>
                        e.key === "Enter" &&
                        onNavigate("project-detail", project.id)
                      }
                    >
                      <td className="px-4 py-3.5">
                        <div className="text-sm font-medium text-slate-900">
                          {project.name}
                        </div>
                        <div className="mt-0.5 flex items-center gap-1 text-xs text-slate-400">
                          <MapPin size={11} aria-hidden="true" />
                          {project.location}
                        </div>
                      </td>
                      <td className="px-4 py-3.5">
                        <StatusBadge status={project.status} />
                      </td>
                      <td className="px-4 py-3.5" style={{ minWidth: 120 }}>
                        <div className="flex items-center gap-2">
                          <Progress
                            value={project.progress}
                            className="flex-1"
                          />
                          <span className="w-8 text-right text-xs font-semibold text-slate-700">
                            {project.progress}%
                          </span>
                        </div>
                      </td>
                      <td className="px-4 py-3.5">
                        <div className="text-sm font-medium text-slate-900">
                          {kes(project.budgetSpent)}
                        </div>
                        <div className="mt-0.5 text-xs text-slate-400">
                          of {kes(project.budgetTotal)}
                        </div>
                      </td>
                      <td className="px-4 py-3.5 text-sm text-slate-600">
                        {fmtDate(project.expectedCompletion)}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
            {filtered.length === 0 && (
              <div className="py-14 text-center">
                <p className="text-sm font-semibold text-slate-700">
                  No projects match your search.
                </p>
                <p className="mt-1 text-sm text-slate-500">
                  Try adjusting the search or filter.
                </p>
              </div>
            )}
          </div>
        )}
      </div>

      {showNew && (
        <NewProjectModal
          onClose={() => setShowNew(false)}
          onCreated={(p) => {
            setShowNew(false)
            onNavigate("project-detail", p.id)
          }}
        />
      )}
    </div>
  )
}
