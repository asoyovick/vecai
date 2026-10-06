import { useCallback, useEffect, useState } from "react"
import { AlertTriangle, CheckCircle, Cloud, Plus, X } from "lucide-react"
import type { NavigateFn } from "../App"
import { api, ApiRequestError } from "../lib/api"
import type { Project, SiteUpdate } from "../lib/api"
import { fmtDate, todayISO } from "../lib/format"
import { useAuth } from "../lib/auth"
import {
  Button,
  ErrorBanner,
  LoadingBlock,
  inputClass,
  labelClass,
} from "../components/ui"

interface Props {
  onNavigate: NavigateFn
}

interface MaterialEntry {
  name: string
  quantity: string
  unit: string
}

const WEATHER = [
  "Clear / Sunny",
  "Partly Cloudy",
  "Overcast",
  "Light Rain",
  "Heavy Rain",
  "Windy",
  "Foggy",
]
const UNITS = [
  "bags",
  "m³",
  "tonnes",
  "pcs",
  "lengths",
  "sheets",
  "rolls",
  "litres",
]

export default function SiteUpdatePage({ onNavigate }: Props) {
  const { user } = useAuth()
  const [projects, setProjects] = useState<Project[]>([])
  const [previous, setPrevious] = useState<SiteUpdate[]>([])
  const [loading, setLoading] = useState(true)
  const [pageError, setPageError] = useState("")

  const [form, setForm] = useState({
    date: todayISO(),
    projectId: "",
    weather: "Partly Cloudy",
    temperature: "",
    workersOnSite: "",
    workCompleted: "",
    issues: "",
    notes: "",
  })
  const [materials, setMaterials] = useState<MaterialEntry[]>([
    { name: "", quantity: "", unit: "bags" },
  ])
  const [fieldError, setFieldError] = useState("")
  const [saving, setSaving] = useState(false)
  const [submitted, setSubmitted] = useState<SiteUpdate | null>(null)

  const load = useCallback(async () => {
    setLoading(true)
    setPageError("")
    try {
      const [p, su] = await Promise.all([
        api.get<{ projects: Project[] }>("/api/projects"),
        api.get<{ siteUpdates: SiteUpdate[] }>("/api/site-updates"),
      ])
      setProjects(p.projects)
      setPrevious(su.siteUpdates)
      setForm((f) => ({
        ...f,
        projectId: f.projectId || String(p.projects[0]?.id ?? ""),
      }))
    } catch (err) {
      setPageError(
        err instanceof Error ? err.message : "Could not load the form.",
      )
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    void load()
  }, [load])

  const update = (field: string, value: string) =>
    setForm((f) => ({ ...f, [field]: value }))

  const addMaterial = () =>
    setMaterials((m) => [...m, { name: "", quantity: "", unit: "bags" }])
  const removeMaterial = (i: number) =>
    setMaterials((m) => m.filter((_, idx) => idx !== i))
  const updateMaterial = (
    i: number,
    field: keyof MaterialEntry,
    value: string,
  ) =>
    setMaterials((m) =>
      m.map((mat, idx) => (idx === i ? { ...mat, [field]: value } : mat)),
    )

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setFieldError("")
    if (!form.workCompleted.trim()) {
      setFieldError(
        "Describe the work completed today — this is the core of the daily record.",
      )
      return
    }
    for (const m of materials) {
      if (
        m.name.trim() &&
        (!m.quantity ||
          Number(m.quantity) <= 0 ||
          Number.isNaN(Number(m.quantity)))
      ) {
        setFieldError(`Enter a valid quantity for “${m.name}”.`)
        return
      }
    }
    setSaving(true)
    try {
      const data = await api.post<{ siteUpdate: SiteUpdate }>(
        "/api/site-updates",
        {
          projectId: Number(form.projectId),
          date: form.date,
          weather: form.weather,
          temperature: form.temperature,
          workersOnSite: form.workersOnSite ? Number(form.workersOnSite) : 0,
          workCompleted: form.workCompleted,
          issues: form.issues,
          notes: form.notes,
          materialsUsed: materials
            .filter((m) => m.name.trim())
            .map((m) => ({
              name: m.name,
              quantity: Number(m.quantity),
              unit: m.unit,
            })),
        },
      )
      setSubmitted(data.siteUpdate)
      window.scrollTo(0, 0)
    } catch (err) {
      setFieldError(
        err instanceof ApiRequestError
          ? err.message
          : "Could not submit the site update.",
      )
    } finally {
      setSaving(false)
    }
  }

  if (loading) return <LoadingBlock label="Loading…" />
  if (pageError)
    return (
      <div className="p-6">
        <ErrorBanner message={pageError} onRetry={load} />
      </div>
    )

  if (submitted) {
    const projectName =
      projects.find((p) => p.id === submitted.project_id)?.name ?? "the project"
    return (
      <div className="flex min-h-[70vh] items-center justify-center px-4 py-10">
        <div className="w-full max-w-md rounded border border-slate-200 bg-white p-8 text-center">
          <CheckCircle
            size={44}
            className="mx-auto mb-4 text-green-700"
            aria-hidden="true"
          />
          <h2 className="mb-2 text-xl font-bold text-slate-900">
            Site update submitted
          </h2>
          <p className="mb-6 text-sm leading-relaxed text-slate-500">
            Your update for{" "}
            <strong className="font-semibold text-slate-700">
              {projectName}
            </strong>{" "}
            on {fmtDate(submitted.date)} has been recorded and shared with the
            project team.
          </p>
          <div className="flex justify-center gap-3">
            <Button
              variant="secondary"
              onClick={() => {
                setSubmitted(null)
                void load()
              }}
            >
              Add another
            </Button>
            <Button
              onClick={() => onNavigate("project-detail", submitted.project_id)}
            >
              View project
            </Button>
          </div>
        </div>
      </div>
    )
  }

  const inputStyle = inputClass

  return (
    <div>
      <div className="border-b border-slate-200 bg-white px-4 py-5 sm:px-8">
        <h1 className="text-xl font-bold tracking-tight text-slate-900">
          Add site update
        </h1>
        <p className="mt-0.5 text-sm text-slate-500">
          Record daily site activity, materials used, and any issues
          encountered.
        </p>
      </div>

      <div className="px-4 py-6 sm:px-8">
        <form onSubmit={handleSubmit} noValidate>
          <div className="grid gap-6 xl:grid-cols-3 xl:items-start">
            {/* Main column */}
            <div className="space-y-4 xl:col-span-2">
              <section
                className="rounded border border-slate-200 bg-white p-5"
                aria-labelledby="su-basic"
              >
                <h2
                  id="su-basic"
                  className="mb-4 border-b border-slate-100 pb-3 text-sm font-semibold text-slate-900"
                >
                  Basic information
                </h2>
                <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
                  <div>
                    <label htmlFor="su-date" className={labelClass}>
                      Date{" "}
                      <span className="text-red-600" aria-hidden="true">
                        *
                      </span>
                    </label>
                    <input
                      id="su-date"
                      type="date"
                      required
                      className={inputStyle}
                      value={form.date}
                      onChange={(e) => update("date", e.target.value)}
                    />
                  </div>
                  <div>
                    <label htmlFor="su-project" className={labelClass}>
                      Project{" "}
                      <span className="text-red-600" aria-hidden="true">
                        *
                      </span>
                    </label>
                    <select
                      id="su-project"
                      required
                      className={inputStyle}
                      value={form.projectId}
                      onChange={(e) => update("projectId", e.target.value)}
                    >
                      {projects.map((p) => (
                        <option key={p.id} value={p.id}>
                          {p.name}
                        </option>
                      ))}
                    </select>
                  </div>
                </div>
              </section>

              <section
                className="rounded border border-slate-200 bg-white p-5"
                aria-labelledby="su-conditions"
              >
                <h2
                  id="su-conditions"
                  className="mb-4 border-b border-slate-100 pb-3 text-sm font-semibold text-slate-900"
                >
                  Site conditions
                </h2>
                <div className="grid grid-cols-1 gap-4 sm:grid-cols-3">
                  <div>
                    <label htmlFor="su-weather" className={labelClass}>
                      Weather
                    </label>
                    <select
                      id="su-weather"
                      className={inputStyle}
                      value={form.weather}
                      onChange={(e) => update("weather", e.target.value)}
                    >
                      {WEATHER.map((w) => (
                        <option key={w}>{w}</option>
                      ))}
                    </select>
                  </div>
                  <div>
                    <label htmlFor="su-temp" className={labelClass}>
                      Temperature (°C)
                    </label>
                    <input
                      id="su-temp"
                      type="number"
                      className={inputStyle}
                      value={form.temperature}
                      onChange={(e) => update("temperature", e.target.value)}
                      placeholder="e.g. 24"
                    />
                  </div>
                  <div>
                    <label htmlFor="su-workers" className={labelClass}>
                      Workers on site
                    </label>
                    <input
                      id="su-workers"
                      type="number"
                      min={0}
                      className={inputStyle}
                      value={form.workersOnSite}
                      onChange={(e) => update("workersOnSite", e.target.value)}
                      placeholder="e.g. 28"
                    />
                  </div>
                </div>
              </section>

              <section
                className="rounded border border-slate-200 bg-white p-5"
                aria-labelledby="su-work"
              >
                <h2
                  id="su-work"
                  className="mb-4 border-b border-slate-100 pb-3 text-sm font-semibold text-slate-900"
                >
                  Work completed today{" "}
                  <span className="text-red-600" aria-hidden="true">
                    *
                  </span>
                </h2>
                <textarea
                  value={form.workCompleted}
                  onChange={(e) => update("workCompleted", e.target.value)}
                  rows={5}
                  required
                  aria-describedby="su-work-hint"
                  placeholder="Describe all work completed today, including specific locations and quantities where applicable."
                  className={`${inputStyle} resize-y leading-relaxed`}
                />
                <p id="su-work-hint" className="mt-1.5 text-xs text-slate-400">
                  Example: “Completed casting of columns C3 and C4 on ground
                  floor. Formwork struck for beam B1.”
                </p>
              </section>

              <section
                className="rounded border border-slate-200 bg-white p-5"
                aria-labelledby="su-materials"
              >
                <div className="mb-4 flex items-center justify-between border-b border-slate-100 pb-3">
                  <h2
                    id="su-materials"
                    className="text-sm font-semibold text-slate-900"
                  >
                    Materials used
                  </h2>
                  <button
                    type="button"
                    onClick={addMaterial}
                    className="inline-flex items-center gap-1 text-xs font-semibold text-slate-700 hover:text-slate-900"
                  >
                    <Plus size={13} aria-hidden="true" /> Add row
                  </button>
                </div>
                <div className="space-y-2">
                  {materials.map((mat, i) => (
                    <div
                      key={i}
                      className="grid grid-cols-[1fr_90px_1fr_auto] items-center gap-2"
                    >
                      <label className="sr-only" htmlFor={`mat-name-${i}`}>
                        Material name
                      </label>
                      <input
                        id={`mat-name-${i}`}
                        type="text"
                        value={mat.name}
                        onChange={(e) =>
                          updateMaterial(i, "name", e.target.value)
                        }
                        placeholder="e.g. Portland Cement"
                        className={inputStyle}
                      />
                      <label className="sr-only" htmlFor={`mat-qty-${i}`}>
                        Quantity
                      </label>
                      <input
                        id={`mat-qty-${i}`}
                        type="number"
                        step="any"
                        min={0}
                        value={mat.quantity}
                        onChange={(e) =>
                          updateMaterial(i, "quantity", e.target.value)
                        }
                        placeholder="Qty"
                        className={inputStyle}
                      />
                      <label className="sr-only" htmlFor={`mat-unit-${i}`}>
                        Unit
                      </label>
                      <select
                        id={`mat-unit-${i}`}
                        value={mat.unit}
                        onChange={(e) =>
                          updateMaterial(i, "unit", e.target.value)
                        }
                        className={inputStyle}
                      >
                        {UNITS.map((u) => (
                          <option key={u}>{u}</option>
                        ))}
                      </select>
                      {materials.length > 1 && (
                        <button
                          type="button"
                          onClick={() => removeMaterial(i)}
                          className="rounded p-1.5 text-slate-400 hover:bg-slate-50 hover:text-slate-600"
                          aria-label={`Remove material row ${i + 1}`}
                        >
                          <X size={14} aria-hidden="true" />
                        </button>
                      )}
                    </div>
                  ))}
                </div>
              </section>

              <section
                className="rounded border border-slate-200 bg-white p-5"
                aria-labelledby="su-issues"
              >
                <div className="mb-4 flex items-center gap-2 border-b border-slate-100 pb-3">
                  <AlertTriangle
                    size={14}
                    className="text-orange-600"
                    aria-hidden="true"
                  />
                  <h2
                    id="su-issues"
                    className="text-sm font-semibold text-slate-900"
                  >
                    Issues / delays
                  </h2>
                  <span className="text-xs text-slate-400">Optional</span>
                </div>
                <textarea
                  value={form.issues}
                  onChange={(e) => update("issues", e.target.value)}
                  rows={3}
                  placeholder="Report any issues, delays, safety concerns, or items requiring management attention."
                  className={`${inputStyle} resize-y leading-relaxed`}
                />
              </section>

              <section
                className="rounded border border-slate-200 bg-white p-5"
                aria-labelledby="su-notes"
              >
                <h2
                  id="su-notes"
                  className="mb-4 border-b border-slate-100 pb-3 text-sm font-semibold text-slate-900"
                >
                  Additional notes{" "}
                  <span className="ml-1 text-xs font-normal text-slate-400">
                    Optional
                  </span>
                </h2>
                <textarea
                  value={form.notes}
                  onChange={(e) => update("notes", e.target.value)}
                  rows={2}
                  placeholder="Any other observations, visitor records, or notes for the project file."
                  className={`${inputStyle} resize-y leading-relaxed`}
                />
              </section>
            </div>

            {/* Right column */}
            <div className="space-y-4">
              <div className="rounded border border-slate-200 bg-white p-5 xl:sticky xl:top-4">
                <h2 className="mb-3 text-sm font-semibold text-slate-900">
                  Submit update
                </h2>
                <div className="mb-4 space-y-2 border-b border-slate-100 pb-4 text-sm">
                  <div className="flex justify-between">
                    <span className="text-slate-500">Submitted by</span>
                    <span className="font-medium text-slate-700">
                      {user?.name}
                    </span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-slate-500">Date</span>
                    <span className="font-medium text-slate-700">
                      {form.date}
                    </span>
                  </div>
                  <div className="flex justify-between gap-3">
                    <span className="text-slate-500">Project</span>
                    <span className="text-right font-medium text-slate-700">
                      {projects.find((p) => p.id === Number(form.projectId))
                        ?.name ?? "—"}
                    </span>
                  </div>
                </div>
                {fieldError && (
                  <p
                    className="mb-3 rounded border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-800"
                    role="alert"
                  >
                    {fieldError}
                  </p>
                )}
                <Button type="submit" loading={saving} className="w-full">
                  {saving ? "Submitting…" : "Submit site update"}
                </Button>
                <p className="mt-3 text-center text-xs text-slate-400">
                  Updates are recorded against the project and shown to the
                  whole team.
                </p>
              </div>

              <div className="rounded border border-slate-200 bg-white p-5">
                <h2 className="mb-3 text-sm font-semibold text-slate-900">
                  Previous updates
                </h2>
                {previous.length === 0 ? (
                  <p className="py-3 text-center text-xs text-slate-400">
                    No updates submitted yet.
                  </p>
                ) : (
                  <ul className="divide-y divide-slate-50">
                    {previous.slice(0, 6).map((u) => (
                      <li key={u.id} className="py-2.5">
                        <div className="flex items-center justify-between">
                          <span className="text-xs font-semibold text-slate-700">
                            {fmtDate(u.date)}
                          </span>
                          <span className="text-xs text-slate-400">
                            {u.workers_on_site} workers
                          </span>
                        </div>
                        <div className="mt-1 flex items-center gap-1.5 text-xs text-slate-400">
                          <Cloud size={11} aria-hidden="true" />
                          {u.weather || "—"} · {u.project_name}
                        </div>
                      </li>
                    ))}
                  </ul>
                )}
              </div>
            </div>
          </div>
        </form>
      </div>
    </div>
  )
}
