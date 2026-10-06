import { useCallback, useEffect, useState } from "react"
import { MapPin, Plus, Search, Star } from "lucide-react"
import type { NavigateFn } from "../App"
import { api, ApiRequestError } from "../lib/api"
import type { Supplier } from "../lib/api"
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

function Stars({ rating }: { rating: number }) {
  return (
    <span
      className="inline-flex items-center gap-0.5"
      title={`${rating.toFixed(1)} out of 5`}
      aria-label={`Rated ${rating.toFixed(1)} out of 5`}
    >
      {[1, 2, 3, 4, 5].map((i) => (
        <Star
          key={i}
          size={12}
          aria-hidden="true"
          className={
            i <= Math.round(rating)
              ? "fill-amber-500 text-amber-500"
              : "text-slate-300"
          }
        />
      ))}
      <span className="ml-1 text-xs text-slate-500">{rating.toFixed(1)}</span>
    </span>
  )
}

function AddSupplierModal({
  onClose,
  onDone,
}: {
  onClose: () => void
  onDone: () => void
}) {
  const [form, setForm] = useState({
    name: "",
    category: "",
    contact: "",
    phone: "",
    email: "",
    location: "",
    rating: "",
  })
  const [error, setError] = useState("")
  const [saving, setSaving] = useState(false)

  const set = (field: string, value: string) =>
    setForm((f) => ({ ...f, [field]: value }))

  const submit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError("")
    if (!form.name.trim()) {
      setError("Supplier name is required.")
      return
    }
    if (!form.category.trim()) {
      setError("Category is required.")
      return
    }
    setSaving(true)
    try {
      await api.post("/api/suppliers", {
        name: form.name,
        category: form.category,
        contact: form.contact,
        phone: form.phone,
        email: form.email,
        location: form.location,
        rating: form.rating ? Number(form.rating) : 0,
      })
      onDone()
    } catch (err) {
      setError(
        err instanceof ApiRequestError
          ? err.message
          : "Could not add the supplier.",
      )
      setSaving(false)
    }
  }

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-slate-900/40 p-4"
      role="dialog"
      aria-modal="true"
      aria-labelledby="as-title"
    >
      <div className="max-h-[90vh] w-full max-w-lg overflow-y-auto rounded border border-slate-200 bg-white shadow-xl">
        <div className="border-b border-slate-100 px-6 py-4">
          <h2 id="as-title" className="text-base font-semibold text-slate-900">
            Add supplier
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
            <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
              <div>
                <label htmlFor="as-name" className={labelClass}>
                  Supplier name{" "}
                  <span className="text-red-600" aria-hidden="true">
                    *
                  </span>
                </label>
                <input
                  id="as-name"
                  className={inputClass}
                  value={form.name}
                  onChange={(e) => set("name", e.target.value)}
                  autoFocus
                />
              </div>
              <div>
                <label htmlFor="as-cat" className={labelClass}>
                  Category{" "}
                  <span className="text-red-600" aria-hidden="true">
                    *
                  </span>
                </label>
                <input
                  id="as-cat"
                  className={inputClass}
                  value={form.category}
                  onChange={(e) => set("category", e.target.value)}
                  placeholder="e.g. Cement & Concrete"
                />
              </div>
              <div>
                <label htmlFor="as-contact" className={labelClass}>
                  Contact person
                </label>
                <input
                  id="as-contact"
                  className={inputClass}
                  value={form.contact}
                  onChange={(e) => set("contact", e.target.value)}
                />
              </div>
              <div>
                <label htmlFor="as-phone" className={labelClass}>
                  Phone
                </label>
                <input
                  id="as-phone"
                  type="tel"
                  className={inputClass}
                  value={form.phone}
                  onChange={(e) => set("phone", e.target.value)}
                />
              </div>
              <div>
                <label htmlFor="as-email" className={labelClass}>
                  Email
                </label>
                <input
                  id="as-email"
                  type="email"
                  className={inputClass}
                  value={form.email}
                  onChange={(e) => set("email", e.target.value)}
                />
              </div>
              <div>
                <label htmlFor="as-loc" className={labelClass}>
                  Location
                </label>
                <input
                  id="as-loc"
                  className={inputClass}
                  value={form.location}
                  onChange={(e) => set("location", e.target.value)}
                />
              </div>
              <div>
                <label htmlFor="as-rating" className={labelClass}>
                  Rating (0–5)
                </label>
                <input
                  id="as-rating"
                  type="number"
                  step="0.1"
                  min={0}
                  max={5}
                  className={inputClass}
                  value={form.rating}
                  onChange={(e) => set("rating", e.target.value)}
                />
              </div>
            </div>
          </div>
          <div className="flex justify-end gap-2 border-t border-slate-100 px-6 py-4">
            <Button type="button" variant="secondary" onClick={onClose}>
              Cancel
            </Button>
            <Button type="submit" loading={saving}>
              {saving ? "Saving…" : "Add supplier"}
            </Button>
          </div>
        </form>
      </div>
    </div>
  )
}

export default function SuppliersPage(_props: Props) {
  const { user } = useAuth()
  const [suppliers, setSuppliers] = useState<Supplier[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState("")
  const [search, setSearch] = useState("")
  const [cat, setCat] = useState("All")
  const [showAdd, setShowAdd] = useState(false)

  const load = useCallback(async () => {
    setLoading(true)
    setError("")
    try {
      const data = await api.get<{ suppliers: Supplier[] }>("/api/suppliers")
      setSuppliers(data.suppliers)
    } catch (err) {
      setError(err instanceof Error ? err.message : "Could not load suppliers.")
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    void load()
  }, [load])

  const categories = [
    "All",
    ...Array.from(new Set(suppliers.map((s) => s.category))),
  ]
  const filtered = suppliers.filter((s) => {
    const q = search.toLowerCase()
    return (
      (s.name.toLowerCase().includes(q) ||
        s.category.toLowerCase().includes(q)) &&
      (cat === "All" || s.category === cat)
    )
  })

  return (
    <div>
      <div className="border-b border-slate-200 bg-white px-4 py-5 sm:px-8">
        <div className="mb-4 flex flex-wrap items-center justify-between gap-3">
          <div>
            <h1 className="text-xl font-bold tracking-tight text-slate-900">
              Suppliers
            </h1>
            <p className="mt-0.5 text-sm text-slate-500">
              {suppliers.length} registered suppliers
            </p>
          </div>
          {user?.role !== "viewer" && (
            <Button onClick={() => setShowAdd(true)}>
              <Plus size={15} aria-hidden="true" />
              Add supplier
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
              placeholder="Search suppliers…"
              aria-label="Search suppliers"
              className="w-56 rounded border border-slate-300 bg-white py-2 pl-9 pr-4 text-sm outline-none focus:border-slate-800 focus:ring-2 focus:ring-slate-900/10"
            />
          </div>
          <div
            className="flex flex-wrap gap-1.5"
            role="group"
            aria-label="Filter by category"
          >
            {categories.map((c) => (
              <button
                key={c}
                onClick={() => setCat(c)}
                aria-pressed={cat === c}
                className={`rounded px-3 py-1.5 text-sm font-medium transition-colors ${
                  cat === c
                    ? "bg-slate-900 text-white"
                    : "border border-slate-300 bg-white text-slate-600 hover:bg-slate-50"
                }`}
              >
                {c}
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
            <LoadingBlock label="Loading suppliers…" />
          </div>
        ) : (
          <div className="overflow-hidden rounded border border-slate-200 bg-white">
            <div className="overflow-x-auto">
              <table className="w-full">
                <thead>
                  <tr className="border-b border-slate-100 bg-slate-50">
                    {[
                      "Supplier",
                      "Category",
                      "Contact",
                      "Phone",
                      "Location",
                      "Rating",
                      "Last order",
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
                  {filtered.map((s) => (
                    <tr
                      key={s.id}
                      className="border-b border-slate-50 last:border-0"
                    >
                      <td className="px-4 py-3 text-sm font-semibold text-slate-800">
                        {s.name}
                      </td>
                      <td className="px-4 py-3">
                        <span className="whitespace-nowrap rounded bg-slate-100 px-2 py-0.5 text-xs text-slate-600">
                          {s.category}
                        </span>
                      </td>
                      <td className="px-4 py-3 text-sm text-slate-600">
                        {s.contact || "—"}
                      </td>
                      <td className="px-4 py-3 text-xs text-slate-500">
                        {s.phone || "—"}
                      </td>
                      <td className="px-4 py-3">
                        <span className="flex items-center gap-1 text-xs text-slate-500">
                          <MapPin size={11} aria-hidden="true" />
                          {s.location || "—"}
                        </span>
                      </td>
                      <td className="px-4 py-3">
                        <Stars rating={s.rating} />
                      </td>
                      <td className="px-4 py-3 text-xs text-slate-500">
                        {s.last_order ?? "—"}
                      </td>
                      <td className="px-4 py-3">
                        <span
                          className={`rounded px-2 py-0.5 text-xs font-medium ${
                            s.status === "active"
                              ? "bg-green-50 text-green-700"
                              : "bg-slate-100 text-slate-500"
                          }`}
                        >
                          {s.status === "active" ? "Active" : "Inactive"}
                        </span>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
            {filtered.length === 0 && (
              <p className="py-12 text-center text-sm text-slate-500">
                No suppliers match your search.
              </p>
            )}
          </div>
        )}
      </div>

      {showAdd && (
        <AddSupplierModal
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
