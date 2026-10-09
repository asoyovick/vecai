import { useCallback, useEffect, useState } from "react"
import { AlertTriangle, Plus, Search } from "lucide-react"
import type { NavigateFn } from "../App"
import { api, ApiRequestError } from "../lib/api"
import type { Material, Project } from "../lib/api"
import { fmtDate, kes } from "../lib/format"
import { useAuth } from "../lib/auth"
import {
  Button,
  ErrorBanner,
  LoadingBlock,
  MaterialBadge,
  Progress,
  inputClass,
  labelClass,
} from "../components/ui"

interface Props {
  onNavigate: NavigateFn
}

const FILTERS = [
  { id: "All", label: "All" },
  { id: "critical", label: "Critical" },
  { id: "low", label: "Low stock" },
  { id: "ordered", label: "On order" },
  { id: "ok", label: "In stock" },
]

function DeliveryModal({
  material,
  onClose,
  onDone,
}: {
  material: Material
  onClose: () => void
  onDone: () => void
}) {
  const [quantity, setQuantity] = useState("")
  const [note, setNote] = useState("")
  const [error, setError] = useState("")
  const [saving, setSaving] = useState(false)

  const submit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError("")
    const qty = Number(quantity)
    if (!quantity || Number.isNaN(qty) || qty <= 0) {
      setError("Enter a quantity greater than zero.")
      return
    }
    setSaving(true)
    try {
      await api.post(`/api/materials/${material.id}/deliveries`, {
        quantity: qty,
        note,
      })
      onDone()
    } catch (err) {
      setError(
        err instanceof ApiRequestError
          ? err.message
          : "Could not log the delivery.",
      )
      setSaving(false)
    }
  }

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-slate-900/40 p-4"
      role="dialog"
      aria-modal="true"
      aria-labelledby="dl-title"
    >
      <div className="w-full max-w-md rounded border border-slate-200 bg-white shadow-xl">
        <div className="border-b border-slate-100 px-6 py-4">
          <h2 id="dl-title" className="text-base font-semibold text-slate-900">
            Log delivery
          </h2>
          <p className="mt-0.5 text-xs text-slate-500">
            {material.name} · currently {material.available_qty} {material.unit}{" "}
            on hand
          </p>
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
              <label htmlFor="dl-qty" className={labelClass}>
                Quantity received ({material.unit}){" "}
                <span className="text-red-600" aria-hidden="true">
                  *
                </span>
              </label>
              <input
                id="dl-qty"
                type="number"
                step="any"
                min={0}
                required
                autoFocus
                className={inputClass}
                value={quantity}
                onChange={(e) => setQuantity(e.target.value)}
              />
            </div>
            <div>
              <label htmlFor="dl-note" className={labelClass}>
                Note (delivery note no., supplier, etc.)
              </label>
              <input
                id="dl-note"
                className={inputClass}
                value={note}
                onChange={(e) => setNote(e.target.value)}
                maxLength={500}
              />
            </div>
          </div>
          <div className="flex justify-end gap-2 border-t border-slate-100 px-6 py-4">
            <Button type="button" variant="secondary" onClick={onClose}>
              Cancel
            </Button>
            <Button type="submit" loading={saving}>
              {saving ? "Saving…" : "Log delivery"}
            </Button>
          </div>
        </form>
      </div>
    </div>
  )
}

function AddMaterialModal({
  projects,
  onClose,
  onDone,
}: {
  projects: Project[]
  onClose: () => void
  onDone: () => void
}) {
  const [form, setForm] = useState({
    projectId: String(projects[0]?.id ?? ""),
    name: "",
    unit: "bags",
    requiredQty: "",
    availableQty: "",
    supplierId: "",
    unitCost: "",
  })
  const [suppliers, setSuppliers] = useState<Array<{
    id: number
    name: string
  }>>([])
  const [error, setError] = useState("")
  const [saving, setSaving] = useState(false)

  useEffect(() => {
    api
      .get<{ suppliers: Array<{ id: number; name: string }> }>("/api/suppliers")
      .then((d) => setSuppliers(d.suppliers))
      .catch(() => setSuppliers([]))
  }, [])

  const set = (field: string, value: string) =>
    setForm((f) => ({ ...f, [field]: value }))

  const submit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError("")
    if (!form.name.trim()) {
      setError("Material name is required.")
      return
    }
    setSaving(true)
    try {
      await api.post("/api/materials", {
        projectId: Number(form.projectId),
        name: form.name,
        unit: form.unit,
        requiredQty: form.requiredQty ? Number(form.requiredQty) : 0,
        availableQty: form.availableQty ? Number(form.availableQty) : 0,
        supplierId: form.supplierId ? Number(form.supplierId) : null,
        unitCost: form.unitCost ? Number(form.unitCost) : 0,
      })
      onDone()
    } catch (err) {
      setError(
        err instanceof ApiRequestError
          ? err.message
          : "Could not add the material.",
      )
      setSaving(false)
    }
  }

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-slate-900/40 p-4"
      role="dialog"
      aria-modal="true"
      aria-labelledby="am-title"
    >
      <div className="max-h-[90vh] w-full max-w-lg overflow-y-auto rounded border border-slate-200 bg-white shadow-xl">
        <div className="border-b border-slate-100 px-6 py-4">
          <h2 id="am-title" className="text-base font-semibold text-slate-900">
            Track a material
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
              <label htmlFor="am-project" className={labelClass}>
                Project{" "}
                <span className="text-red-600" aria-hidden="true">
                  *
                </span>
              </label>
              <select
                id="am-project"
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
            <div className="grid grid-cols-2 gap-4">
              <div>
                <label htmlFor="am-name" className={labelClass}>
                  Material{" "}
                  <span className="text-red-600" aria-hidden="true">
                    *
                  </span>
                </label>
                <input
                  id="am-name"
                  className={inputClass}
                  value={form.name}
                  onChange={(e) => set("name", e.target.value)}
                  placeholder="e.g. Portland Cement (50kg bags)"
                />
              </div>
              <div>
                <label htmlFor="am-unit" className={labelClass}>
                  Unit
                </label>
                <input
                  id="am-unit"
                  className={inputClass}
                  value={form.unit}
                  onChange={(e) => set("unit", e.target.value)}
                  placeholder="bags, m³, tonnes…"
                />
              </div>
              <div>
                <label htmlFor="am-req" className={labelClass}>
                  Required quantity
                </label>
                <input
                  id="am-req"
                  type="number"
                  step="any"
                  min={0}
                  className={inputClass}
                  value={form.requiredQty}
                  onChange={(e) => set("requiredQty", e.target.value)}
                />
              </div>
              <div>
                <label htmlFor="am-avail" className={labelClass}>
                  Available quantity
                </label>
                <input
                  id="am-avail"
                  type="number"
                  step="any"
                  min={0}
                  className={inputClass}
                  value={form.availableQty}
                  onChange={(e) => set("availableQty", e.target.value)}
                />
              </div>
              <div>
                <label htmlFor="am-sup" className={labelClass}>
                  Supplier
                </label>
                <select
                  id="am-sup"
                  className={inputClass}
                  value={form.supplierId}
                  onChange={(e) => set("supplierId", e.target.value)}
                >
                  <option value="">None</option>
                  {suppliers.map((s) => (
                    <option key={s.id} value={s.id}>
                      {s.name}
                    </option>
                  ))}
                </select>
              </div>
              <div>
                <label htmlFor="am-cost" className={labelClass}>
                  Unit cost (KES)
                </label>
                <input
                  id="am-cost"
                  type="number"
                  min={0}
                  className={inputClass}
                  value={form.unitCost}
                  onChange={(e) => set("unitCost", e.target.value)}
                />
              </div>
            </div>
          </div>
          <div className="flex justify-end gap-2 border-t border-slate-100 px-6 py-4">
            <Button type="button" variant="secondary" onClick={onClose}>
              Cancel
            </Button>
            <Button type="submit" loading={saving}>
              {saving ? "Saving…" : "Add material"}
            </Button>
          </div>
        </form>
      </div>
    </div>
  )
}

export default function MaterialsPage(_props: Props) {
  const { user } = useAuth()
  const [materials, setMaterials] = useState<Material[]>([])
  const [projects, setProjects] = useState<Project[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState("")
  const [search, setSearch] = useState("")
  const [filter, setFilter] = useState("All")
  const [projectFilter, setProjectFilter] = useState<string>("All")
  const [deliveryTarget, setDeliveryTarget] = useState<Material | null>(null)
  const [showAdd, setShowAdd] = useState(false)

  const load = useCallback(async () => {
    setLoading(true)
    setError("")
    try {
      const [m, p] = await Promise.all([
        api.get<{ materials: Material[] }>("/api/materials"),
        api.get<{ projects: Project[] }>("/api/projects"),
      ])
      setMaterials(m.materials)
      setProjects(p.projects)
    } catch (err) {
      setError(err instanceof Error ? err.message : "Could not load materials.")
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    void load()
  }, [load])

  const filtered = materials.filter((m) => {
    const q = search.toLowerCase()
    return (
      m.name.toLowerCase().includes(q) &&
      (filter === "All" || m.status === filter) &&
      (projectFilter === "All" || m.project_id === Number(projectFilter))
    )
  })

  const criticalCount = materials.filter((m) => m.status === "critical").length
  const lowCount = materials.filter((m) => m.status === "low").length
  const canEdit = user?.role !== "viewer"

  return (
    <div>
      <div className="border-b border-slate-200 bg-white px-4 py-5 sm:px-8">
        <div className="mb-4 flex flex-wrap items-center justify-between gap-3">
          <div>
            <h1 className="text-xl font-bold tracking-tight text-slate-900">
              Materials
            </h1>
            <p className="mt-0.5 text-sm text-slate-500">
              {materials.length} materials tracked across projects
            </p>
          </div>
          {canEdit && (
            <Button onClick={() => setShowAdd(true)}>
              <Plus size={15} aria-hidden="true" />
              Track material
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
              placeholder="Search materials…"
              aria-label="Search materials"
              className="w-56 rounded border border-slate-300 bg-white py-2 pl-9 pr-4 text-sm outline-none focus:border-slate-800 focus:ring-2 focus:ring-slate-900/10"
            />
          </div>
          <select
            value={projectFilter}
            onChange={(e) => setProjectFilter(e.target.value)}
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
            className="flex flex-wrap gap-1.5"
            role="group"
            aria-label="Filter by status"
          >
            {FILTERS.map((f) => (
              <button
                key={f.id}
                onClick={() => setFilter(f.id)}
                aria-pressed={filter === f.id}
                className={`rounded px-3 py-1.5 text-sm font-medium transition-colors ${
                  filter === f.id
                    ? "bg-slate-900 text-white"
                    : "border border-slate-300 bg-white text-slate-600 hover:bg-slate-50"
                }`}
              >
                {f.label}
                {f.id === "critical" && criticalCount > 0 && (
                  <span className="ml-1.5 rounded bg-red-100 px-1 text-xs font-bold text-red-700">
                    {criticalCount}
                  </span>
                )}
                {f.id === "low" && lowCount > 0 && (
                  <span className="ml-1.5 rounded bg-orange-100 px-1 text-xs font-bold text-orange-700">
                    {lowCount}
                  </span>
                )}
              </button>
            ))}
          </div>
        </div>
      </div>

      <div className="px-4 py-6 sm:px-8">
        {error && (
          <div className="mb-4">
            <ErrorBanner message={error} onRetry={load} />
          </div>
        )}

        {loading ? (
          <div className="rounded border border-slate-200 bg-white">
            <LoadingBlock label="Loading materials…" />
          </div>
        ) : (
          <>
            {(criticalCount > 0 || lowCount > 0) && filter === "All" && (
              <div
                className="mb-4 flex items-start gap-3 rounded border border-red-200 bg-red-50 p-4"
                role="alert"
              >
                <AlertTriangle
                  size={16}
                  className="mt-0.5 shrink-0 text-red-600"
                  aria-hidden="true"
                />
                <div>
                  <p className="text-sm font-semibold text-red-800">
                    {criticalCount} critical{" "}
                    {criticalCount === 1 ? "item" : "items"} and {lowCount}{" "}
                    low-stock {lowCount === 1 ? "item" : "items"} need
                    attention.
                  </p>
                  {criticalCount > 0 && (
                    <p className="mt-0.5 text-xs text-red-700">
                      Critical:{" "}
                      {materials
                        .filter((m) => m.status === "critical")
                        .map((m) => m.name)
                        .join(" · ")}
                    </p>
                  )}
                </div>
              </div>
            )}

            <div className="overflow-hidden rounded border border-slate-200 bg-white">
              <div className="overflow-x-auto">
                <table className="w-full">
                  <thead>
                    <tr className="border-b border-slate-100 bg-slate-50">
                      {[
                        "Material",
                        "Project",
                        "Required",
                        "Available",
                        "Used",
                        "Supplier",
                        "Unit cost",
                        "Last delivery",
                        "Status",
                        "",
                      ].map((h) => (
                        <th
                          key={h}
                          className="whitespace-nowrap px-4 py-3 text-left text-xs font-semibold uppercase tracking-wide text-slate-500"
                        >
                          {h}
                        </th>
                      ))}
                    </tr>
                  </thead>
                  <tbody>
                    {filtered.map((m) => {
                      const pctUsed = m.required_qty
                        ? Math.round((m.used_qty / m.required_qty) * 100)
                        : 0
                      return (
                        <tr
                          key={m.id}
                          className="border-b border-slate-50 last:border-0"
                        >
                          <td className="px-4 py-3">
                            <div className="flex items-center gap-2">
                              {(m.status === "critical" ||
                                m.status === "low") && (
                                <AlertTriangle
                                  size={13}
                                  className={
                                    m.status === "critical"
                                      ? "shrink-0 text-red-600"
                                      : "shrink-0 text-orange-500"
                                  }
                                  aria-hidden="true"
                                />
                              )}
                              <span className="text-sm font-medium text-slate-700">
                                {m.name}
                              </span>
                            </div>
                          </td>
                          <td className="px-4 py-3 text-xs text-slate-500">
                            {m.project_name}
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
                          <td className="px-4 py-3">
                            <div className="text-sm text-slate-600">
                              {m.used_qty.toLocaleString()}
                            </div>
                            <Progress value={pctUsed} className="mt-1" />
                          </td>
                          <td className="max-w-[160px] truncate px-4 py-3 text-xs text-slate-500">
                            {m.supplier_name ?? "–"}
                          </td>
                          <td className="px-4 py-3 text-xs text-slate-600">
                            {m.unit_cost ? kes(m.unit_cost) : "–"}
                          </td>
                          <td className="px-4 py-3 text-xs text-slate-500">
                            {fmtDate(m.last_delivery)}
                          </td>
                          <td className="px-4 py-3">
                            <MaterialBadge status={m.status} />
                          </td>
                          <td className="px-4 py-3 text-right">
                            {canEdit && (
                              <button
                                onClick={() => setDeliveryTarget(m)}
                                className="rounded border border-slate-300 px-2.5 py-1 text-xs font-semibold text-slate-700 transition-colors hover:bg-slate-50 focus-visible:outline-2 focus-visible:outline-slate-800"
                              >
                                Log delivery
                              </button>
                            )}
                          </td>
                        </tr>
                      )
                    })}
                  </tbody>
                </table>
              </div>
              {filtered.length === 0 && (
                <p className="py-12 text-center text-sm text-slate-500">
                  No materials match your filters.
                </p>
              )}
            </div>
          </>
        )}
      </div>

      {deliveryTarget && (
        <DeliveryModal
          material={deliveryTarget}
          onClose={() => setDeliveryTarget(null)}
          onDone={() => {
            setDeliveryTarget(null)
            void load()
          }}
        />
      )}
      {showAdd && (
        <AddMaterialModal
          projects={projects}
          onClose={() => setShowAdd(false)}
          onDone={() => {
            setShowAdd(false)
            void load()
          }}
        />
      )}
    </div>
  )
}
