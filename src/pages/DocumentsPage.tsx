import { useCallback, useEffect, useRef, useState } from "react"
import {
  Download,
  FileArchive,
  FileSpreadsheet,
  FileText,
  Image as ImageIcon,
  Plus,
  Search,
  Upload,
} from "lucide-react"
import type { NavigateFn } from "../App"
import { api, ApiRequestError, uploadDocument } from "../lib/api"
import type { DocumentItem, Project } from "../lib/api"
import { fmtDate, fmtSize } from "../lib/format"
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

const CATEGORIES = [
  "Architectural Drawing",
  "Structural Drawing",
  "BOQ",
  "Contract",
  "Permit",
  "Invoice",
  "Site Report",
  "Inspection Report",
  "Photo",
]

function DocIcon({ format }: { format: string }) {
  const props = { size: 17, strokeWidth: 1.5 } as const
  if (format === "XLSX" || format === "XLS" || format === "CSV")
    return (
      <FileSpreadsheet
        {...props}
        className="text-green-700"
        aria-hidden="true"
      />
    )
  if (format === "ZIP")
    return (
      <FileArchive {...props} className="text-violet-700" aria-hidden="true" />
    )
  if (["JPG", "JPEG", "PNG", "WEBP"].includes(format))
    return <ImageIcon {...props} className="text-sky-700" aria-hidden="true" />
  return <FileText {...props} className="text-slate-500" aria-hidden="true" />
}

function UploadModal({
  projects,
  onClose,
  onDone,
}: {
  projects: Project[]
  onClose: () => void
  onDone: () => void
}) {
  const [projectId, setProjectId] = useState(String(projects[0]?.id ?? ""))
  const [name, setName] = useState("")
  const [category, setCategory] = useState(CATEGORIES[0])
  const [version, setVersion] = useState("1.0")
  const [file, setFile] = useState<File | null>(null)
  const [error, setError] = useState("")
  const [saving, setSaving] = useState(false)
  const fileRef = useRef<HTMLInputElement>(null)

  const submit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError("")
    if (!file) {
      setError("Choose a file to upload.")
      return
    }
    if (!name.trim()) {
      setError("Document name is required.")
      return
    }
    setSaving(true)
    try {
      await uploadDocument(Number(projectId), file, { name, category, version })
      onDone()
    } catch (err) {
      setError(
        err instanceof ApiRequestError
          ? err.message
          : "Upload failed. Please try again.",
      )
      setSaving(false)
    }
  }

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-slate-900/40 p-4"
      role="dialog"
      aria-modal="true"
      aria-labelledby="up-title"
    >
      <div className="w-full max-w-md rounded border border-slate-200 bg-white shadow-xl">
        <div className="border-b border-slate-100 px-6 py-4">
          <h2 id="up-title" className="text-base font-semibold text-slate-900">
            Upload document
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
              <label htmlFor="up-project" className={labelClass}>
                Project{" "}
                <span className="text-red-600" aria-hidden="true">
                  *
                </span>
              </label>
              <select
                id="up-project"
                className={inputClass}
                value={projectId}
                onChange={(e) => setProjectId(e.target.value)}
              >
                {projects.map((p) => (
                  <option key={p.id} value={p.id}>
                    {p.name}
                  </option>
                ))}
              </select>
            </div>
            <div>
              <label htmlFor="up-file" className={labelClass}>
                File{" "}
                <span className="text-red-600" aria-hidden="true">
                  *
                </span>
              </label>
              <input
                id="up-file"
                ref={fileRef}
                type="file"
                className="block w-full text-sm text-slate-600 file:mr-3 file:rounded file:border-0 file:bg-slate-100 file:px-3 file:py-2 file:text-xs file:font-semibold file:text-slate-700 hover:file:bg-slate-200"
                onChange={(e) => {
                  const f = e.target.files?.[0] ?? null
                  setFile(f)
                  if (f && !name.trim()) setName(f.name.replace(/\.[^.]+$/, ""))
                }}
                accept=".pdf,.xls,.xlsx,.csv,.doc,.docx,.png,.jpg,.jpeg,.webp,.zip,.dwg,.ifc,.txt"
              />
              <p className="mt-1 text-xs text-slate-400">
                PDF, Office, images, ZIP, DWG · up to 25 MB
              </p>
            </div>
            <div>
              <label htmlFor="up-name" className={labelClass}>
                Document name{" "}
                <span className="text-red-600" aria-hidden="true">
                  *
                </span>
              </label>
              <input
                id="up-name"
                className={inputClass}
                value={name}
                onChange={(e) => setName(e.target.value)}
              />
            </div>
            <div className="grid grid-cols-2 gap-4">
              <div>
                <label htmlFor="up-cat" className={labelClass}>
                  Category
                </label>
                <select
                  id="up-cat"
                  className={inputClass}
                  value={category}
                  onChange={(e) => setCategory(e.target.value)}
                >
                  {CATEGORIES.map((c) => (
                    <option key={c}>{c}</option>
                  ))}
                </select>
              </div>
              <div>
                <label htmlFor="up-ver" className={labelClass}>
                  Version
                </label>
                <input
                  id="up-ver"
                  className={inputClass}
                  value={version}
                  onChange={(e) => setVersion(e.target.value)}
                />
              </div>
            </div>
          </div>
          <div className="flex justify-end gap-2 border-t border-slate-100 px-6 py-4">
            <Button type="button" variant="secondary" onClick={onClose}>
              Cancel
            </Button>
            <Button type="submit" loading={saving}>
              {saving ? "Uploading…" : "Upload"}
            </Button>
          </div>
        </form>
      </div>
    </div>
  )
}

export default function DocumentsPage(_props: Props) {
  const { user } = useAuth()
  const [documents, setDocuments] = useState<DocumentItem[]>([])
  const [projects, setProjects] = useState<Project[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState("")
  const [search, setSearch] = useState("")
  const [category, setCategory] = useState("All")
  const [projectFilter, setProjectFilter] = useState("All")
  const [showUpload, setShowUpload] = useState(false)
  const [downloadingId, setDownloadingId] = useState<number | null>(null)

  const load = useCallback(async () => {
    setLoading(true)
    setError("")
    try {
      const [d, p] = await Promise.all([
        api.get<{ documents: DocumentItem[] }>("/api/documents"),
        api.get<{ projects: Project[] }>("/api/projects"),
      ])
      setDocuments(d.documents)
      setProjects(p.projects)
    } catch (err) {
      setError(err instanceof Error ? err.message : "Could not load documents.")
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    void load()
  }, [load])

  const download = async (doc: DocumentItem) => {
    setDownloadingId(doc.id)
    setError("")
    try {
      const res = await fetch(`/api/documents/${doc.id}/download`, {
        credentials: "same-origin",
      })
      if (!res.ok) {
        const data = await res.json().catch(() => null)
        throw new Error(
          (data as { error?: string })?.error || "Download failed.",
        )
      }
      const blob = await res.blob()
      const url = URL.createObjectURL(blob)
      const a = window.document.createElement("a")
      a.href = url
      a.download = doc.original_name
      window.document.body.appendChild(a)
      a.click()
      a.remove()
      URL.revokeObjectURL(url)
    } catch (err) {
      setError(err instanceof Error ? err.message : "Download failed.")
    } finally {
      setDownloadingId(null)
    }
  }

  const filtered = documents.filter((d) => {
    const q = search.toLowerCase()
    return (
      d.name.toLowerCase().includes(q) &&
      (category === "All" || d.category === category) &&
      (projectFilter === "All" || d.project_id === Number(projectFilter))
    )
  })

  const countFor = (cat: string) =>
    cat === "All"
      ? documents.length
      : documents.filter((d) => d.category === cat).length
  const canEdit = user?.role !== "viewer"

  return (
    <div className="flex min-h-[calc(100vh-56px)] flex-col">
      <div className="border-b border-slate-200 bg-white px-4 py-5 sm:px-8">
        <div className="mb-4 flex flex-wrap items-center justify-between gap-3">
          <div>
            <h1 className="text-xl font-bold tracking-tight text-slate-900">
              Documents
            </h1>
            <p className="mt-0.5 text-sm text-slate-500">
              {documents.length} documents on file
            </p>
          </div>
          {canEdit && (
            <Button onClick={() => setShowUpload(true)}>
              <Plus size={15} aria-hidden="true" />
              Upload document
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
              placeholder="Search documents…"
              aria-label="Search documents"
              className="w-60 rounded border border-slate-300 bg-white py-2 pl-9 pr-4 text-sm outline-none focus:border-slate-800 focus:ring-2 focus:ring-slate-900/10"
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
        </div>
      </div>

      <div className="flex flex-1 flex-col md:flex-row">
        {/* Category sidebar */}
        <nav
          className="shrink-0 border-b border-slate-200 bg-white px-4 py-3 md:w-52 md:border-b-0 md:border-r md:py-4"
          aria-label="Filter by category"
        >
          <p className="mb-2 hidden text-xs font-semibold uppercase tracking-widest text-slate-400 md:block">
            Category
          </p>
          <div className="flex gap-1.5 overflow-x-auto md:flex-col md:overflow-visible">
            {["All", ...CATEGORIES].map((cat) => (
              <button
                key={cat}
                onClick={() => setCategory(cat)}
                aria-pressed={category === cat}
                className={`flex shrink-0 items-center justify-between gap-2 rounded px-3 py-1.5 text-sm transition-colors md:w-full md:rounded-none ${
                  category === cat
                    ? "bg-slate-100 font-semibold text-slate-900 md:border-l-2 md:border-orange-600"
                    : "text-slate-600 hover:bg-slate-50"
                }`}
              >
                <span className="truncate">{cat}</span>
                <span className="text-xs text-slate-400">{countFor(cat)}</span>
              </button>
            ))}
          </div>
        </nav>

        {/* Document list */}
        <div className="flex-1 p-4 sm:p-6">
          {error && (
            <div className="mb-4">
              <ErrorBanner message={error} onRetry={load} />
            </div>
          )}
          {loading ? (
            <div className="rounded border border-slate-200 bg-white">
              <LoadingBlock label="Loading documents…" />
            </div>
          ) : (
            <div className="overflow-hidden rounded border border-slate-200 bg-white">
              <div className="overflow-x-auto">
                <table className="w-full">
                  <thead>
                    <tr className="border-b border-slate-100 bg-slate-50">
                      {[
                        "Name",
                        "Category",
                        "Project",
                        "Version",
                        "Uploaded by",
                        "Date",
                        "Size",
                        "",
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
                    {filtered.map((doc) => (
                      <tr
                        key={doc.id}
                        className="border-b border-slate-50 last:border-0"
                      >
                        <td className="px-4 py-3">
                          <div className="flex items-center gap-3">
                            <DocIcon format={doc.format} />
                            <div className="min-w-0">
                              <div className="max-w-xs truncate text-sm font-medium text-slate-700">
                                {doc.name}
                              </div>
                              <div className="text-xs text-slate-400">
                                {doc.format || "—"} · {fmtSize(doc.size_bytes)}
                              </div>
                            </div>
                          </div>
                        </td>
                        <td className="px-4 py-3">
                          <span className="whitespace-nowrap rounded bg-slate-100 px-2 py-0.5 text-xs text-slate-600">
                            {doc.category}
                          </span>
                        </td>
                        <td className="px-4 py-3 text-xs text-slate-500">
                          {doc.project_name ?? "—"}
                        </td>
                        <td className="px-4 py-3 text-xs font-medium text-slate-600">
                          {doc.version}
                        </td>
                        <td className="px-4 py-3 text-xs text-slate-500">
                          {doc.uploaded_by_name ?? "—"}
                        </td>
                        <td className="px-4 py-3 text-xs text-slate-500">
                          {fmtDate(doc.created_at)}
                        </td>
                        <td className="px-4 py-3 text-xs text-slate-400">
                          {fmtSize(doc.size_bytes)}
                        </td>
                        <td className="px-4 py-3 text-right">
                          <button
                            onClick={() => download(doc)}
                            disabled={downloadingId === doc.id}
                            title="Download"
                            aria-label={`Download ${doc.name}`}
                            className="rounded p-1.5 text-slate-500 transition-colors hover:bg-slate-100 hover:text-slate-800 disabled:opacity-50"
                          >
                            <Download size={14} aria-hidden="true" />
                          </button>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
              {filtered.length === 0 && (
                <div className="py-14 text-center">
                  <Upload
                    size={28}
                    className="mx-auto mb-2 text-slate-300"
                    aria-hidden="true"
                  />
                  <p className="text-sm font-semibold text-slate-700">
                    No documents found
                  </p>
                  <p className="mt-1 text-sm text-slate-500">
                    Upload a document or adjust your filters.
                  </p>
                </div>
              )}
            </div>
          )}
        </div>
      </div>

      {showUpload && (
        <UploadModal
          projects={projects}
          onClose={() => setShowUpload(false)}
          onDone={() => {
            setShowUpload(false)
            void load()
          }}
        />
      )}
    </div>
  )
}
