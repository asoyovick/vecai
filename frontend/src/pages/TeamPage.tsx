import { useCallback, useEffect, useState } from "react"
import { Mail, Phone, Plus, Search } from "lucide-react"
import type { NavigateFn } from "../App"
import { api, ApiRequestError } from "../lib/api"
import type { Project, TeamMember } from "../lib/api"
import { useAuth } from "../lib/auth"
import {
  Button,
  ErrorBanner,
  LoadingBlock,
  SuccessBanner,
  inputClass,
  labelClass,
} from "../components/ui"

interface Props {
  onNavigate: NavigateFn
}

interface InviteResult {
  member: TeamMember
  temporaryPassword: string
}

function InviteModal({
  projects,
  onClose,
  onDone,
}: {
  projects: Project[]
  onClose: () => void
  onDone: (r: InviteResult) => void
}) {
  const [form, setForm] = useState({
    name: "",
    email: "",
    jobTitle: "",
    role: "member",
    phone: "",
    projectId: "",
  })
  const [error, setError] = useState("")
  const [saving, setSaving] = useState(false)

  const set = (field: string, value: string) =>
    setForm((f) => ({ ...f, [field]: value }))

  const submit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError("")
    if (!form.name.trim()) {
      setError("Full name is required.")
      return
    }
    if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(form.email)) {
      setError("Enter a valid email address.")
      return
    }
    if (!form.jobTitle.trim()) {
      setError("Job title is required.")
      return
    }
    setSaving(true)
    try {
      const data = await api.post<InviteResult>("/api/team/invite", {
        name: form.name,
        email: form.email,
        jobTitle: form.jobTitle,
        role: form.role,
        phone: form.phone,
        projectId: form.projectId ? Number(form.projectId) : null,
      })
      onDone(data)
    } catch (err) {
      setError(
        err instanceof ApiRequestError
          ? err.message
          : "Could not create the invitation.",
      )
      setSaving(false)
    }
  }

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-slate-900/40 p-4"
      role="dialog"
      aria-modal="true"
      aria-labelledby="inv-title"
    >
      <div className="w-full max-w-md rounded border border-slate-200 bg-white shadow-xl">
        <div className="border-b border-slate-100 px-6 py-4">
          <h2 id="inv-title" className="text-base font-semibold text-slate-900">
            Invite team member
          </h2>
          <p className="mt-0.5 text-xs text-slate-500">
            A temporary password is generated for their first sign-in.
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
              <label htmlFor="inv-name" className={labelClass}>
                Full name{" "}
                <span className="text-red-600" aria-hidden="true">
                  *
                </span>
              </label>
              <input
                id="inv-name"
                className={inputClass}
                value={form.name}
                onChange={(e) => set("name", e.target.value)}
                autoFocus
              />
            </div>
            <div>
              <label htmlFor="inv-email" className={labelClass}>
                Email{" "}
                <span className="text-red-600" aria-hidden="true">
                  *
                </span>
              </label>
              <input
                id="inv-email"
                type="email"
                className={inputClass}
                value={form.email}
                onChange={(e) => set("email", e.target.value)}
              />
            </div>
            <div className="grid grid-cols-2 gap-4">
              <div>
                <label htmlFor="inv-title" className={labelClass}>
                  Job title{" "}
                  <span className="text-red-600" aria-hidden="true">
                    *
                  </span>
                </label>
                <input
                  id="inv-title"
                  className={inputClass}
                  value={form.jobTitle}
                  onChange={(e) => set("jobTitle", e.target.value)}
                />
              </div>
              <div>
                <label htmlFor="inv-role" className={labelClass}>
                  Role
                </label>
                <select
                  id="inv-role"
                  className={inputClass}
                  value={form.role}
                  onChange={(e) => set("role", e.target.value)}
                >
                  <option value="member">Member (can edit)</option>
                  <option value="viewer">Viewer (read only)</option>
                </select>
              </div>
            </div>
            <div>
              <label htmlFor="inv-phone" className={labelClass}>
                Phone
              </label>
              <input
                id="inv-phone"
                type="tel"
                className={inputClass}
                value={form.phone}
                onChange={(e) => set("phone", e.target.value)}
              />
            </div>
            <div>
              <label htmlFor="inv-project" className={labelClass}>
                Assign to project
              </label>
              <select
                id="inv-project"
                className={inputClass}
                value={form.projectId}
                onChange={(e) => set("projectId", e.target.value)}
              >
                <option value="">No project assignment</option>
                {projects.map((p) => (
                  <option key={p.id} value={p.id}>
                    {p.name}
                  </option>
                ))}
              </select>
            </div>
          </div>
          <div className="flex justify-end gap-2 border-t border-slate-100 px-6 py-4">
            <Button type="button" variant="secondary" onClick={onClose}>
              Cancel
            </Button>
            <Button type="submit" loading={saving}>
              {saving ? "Creating…" : "Send invite"}
            </Button>
          </div>
        </form>
      </div>
    </div>
  )
}

interface Assignment {
  user_id: number
  project_id: number
  project_name: string
}

export default function TeamPage({ onNavigate }: Props) {
  const { user } = useAuth()
  const [members, setMembers] = useState<TeamMember[]>([])
  const [assignments, setAssignments] = useState<Assignment[]>([])
  const [projects, setProjects] = useState<Project[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState("")
  const [search, setSearch] = useState("")
  const [showInvite, setShowInvite] = useState(false)
  const [inviteResult, setInviteResult] = useState<InviteResult | null>(null)

  const load = useCallback(async () => {
    setLoading(true)
    setError("")
    try {
      const [t, p] = await Promise.all([
        api.get<{ members: TeamMember[]; assignments: Assignment[] }>(
          "/api/team",
        ),
        api.get<{ projects: Project[] }>("/api/projects"),
      ])
      setMembers(t.members)
      setAssignments(t.assignments)
      setProjects(p.projects)
    } catch (err) {
      setError(err instanceof Error ? err.message : "Could not load the team.")
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    void load()
  }, [load])

  const filtered = members.filter((m) => {
    const q = search.toLowerCase()
    return (
      m.name.toLowerCase().includes(q) ||
      m.job_title.toLowerCase().includes(q) ||
      m.email.toLowerCase().includes(q)
    )
  })

  const roleBadge = (role: string) =>
    role === "admin"
      ? { label: "Admin", cls: "bg-slate-900 text-white" }
      : role === "viewer"
        ? { label: "Viewer", cls: "bg-amber-50 text-amber-800" }
        : { label: "Member", cls: "bg-slate-100 text-slate-600" }

  return (
    <div>
      <div className="border-b border-slate-200 bg-white px-4 py-5 sm:px-8">
        <div className="mb-4 flex flex-wrap items-center justify-between gap-3">
          <div>
            <h1 className="text-xl font-bold tracking-tight text-slate-900">
              Team
            </h1>
            <p className="mt-0.5 text-sm text-slate-500">
              {members.length} people in the workspace
            </p>
          </div>
          {user?.role === "admin" && (
            <Button onClick={() => setShowInvite(true)}>
              <Plus size={15} aria-hidden="true" />
              Invite member
            </Button>
          )}
        </div>
        <div className="relative max-w-sm">
          <Search
            size={14}
            className="pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 text-slate-400"
            aria-hidden="true"
          />
          <input
            type="search"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            placeholder="Search team members…"
            aria-label="Search team members"
            className="w-full rounded border border-slate-300 bg-white py-2 pl-9 pr-4 text-sm outline-none focus:border-slate-800 focus:ring-2 focus:ring-slate-900/10"
          />
        </div>
      </div>

      <div className="px-4 py-6 sm:px-8">
        {error && (
          <div className="mb-4">
            <ErrorBanner message={error} onRetry={load} />
          </div>
        )}
        {inviteResult && (
          <div className="mb-4">
            <SuccessBanner
              message={`${inviteResult.member.name} added. Temporary password: ${inviteResult.temporaryPassword}. Share it securely; it is shown only once.`}
            />
          </div>
        )}

        {loading ? (
          <LoadingBlock label="Loading team…" />
        ) : (
          <div className="grid gap-3 sm:grid-cols-2 xl:grid-cols-3">
            {filtered.map((m) => {
              const rb = roleBadge(m.role)
              const memberProjects = assignments.filter(
                (a) => a.user_id === m.id,
              )
              return (
                <div
                  key={m.id}
                  className="rounded border border-slate-200 bg-white p-5"
                >
                  <div className="mb-4 flex items-center gap-3.5 border-b border-slate-100 pb-4">
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
                      <div className="mt-1 flex items-center gap-1.5">
                        <span className="rounded bg-slate-100 px-1.5 py-0.5 text-xs text-slate-600">
                          {m.job_title || "–"}
                        </span>
                        <span
                          className={`rounded px-1.5 py-0.5 text-xs font-medium ${rb.cls}`}
                        >
                          {rb.label}
                        </span>
                      </div>
                    </div>
                  </div>
                  <div className="mb-3 space-y-1.5">
                    <div className="flex items-center gap-2 text-xs text-slate-500">
                      <Mail
                        size={12}
                        className="shrink-0 text-slate-400"
                        aria-hidden="true"
                      />
                      <span className="truncate">{m.email}</span>
                    </div>
                    {m.phone && (
                      <div className="flex items-center gap-2 text-xs text-slate-500">
                        <Phone
                          size={12}
                          className="shrink-0 text-slate-400"
                          aria-hidden="true"
                        />
                        {m.phone}
                      </div>
                    )}
                  </div>
                  {memberProjects.length > 0 && (
                    <div>
                      <p className="mb-1.5 text-xs font-semibold uppercase tracking-wide text-slate-400">
                        Projects
                      </p>
                      <div className="flex flex-wrap gap-1.5">
                        {memberProjects.map((a) => (
                          <button
                            key={a.project_id}
                            onClick={() =>
                              onNavigate("project-detail", a.project_id)
                            }
                            className="rounded bg-slate-50 px-2 py-0.5 text-xs text-slate-600 transition-colors hover:bg-slate-100 hover:text-slate-900"
                          >
                            {a.project_name.length > 26
                              ? a.project_name.slice(0, 26) + "…"
                              : a.project_name}
                          </button>
                        ))}
                      </div>
                    </div>
                  )}
                </div>
              )
            })}
            {filtered.length === 0 && (
              <p className="col-span-full py-12 text-center text-sm text-slate-500">
                No team members match your search.
              </p>
            )}
          </div>
        )}
      </div>

      {showInvite && (
        <InviteModal
          projects={projects}
          onClose={() => setShowInvite(false)}
          onDone={(r) => {
            setShowInvite(false)
            setInviteResult(r)
            void load()
          }}
        />
      )}
    </div>
  )
}
