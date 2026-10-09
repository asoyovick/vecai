import { useEffect, useState } from "react"
import type { NavigateFn } from "../App"
import { api, ApiRequestError } from "../lib/api"
import type { User } from "../lib/api"
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

const TABS = [
  { id: "profile", label: "Profile" },
  { id: "organisation", label: "Organisation" },
  { id: "notifications", label: "Notifications" },
  { id: "security", label: "Security" },
] as const
type Tab = typeof TABS[number]["id"]

const NOTIFICATION_SETTINGS = [
  {
    id: "material-alerts",
    label: "Material stock alerts",
    desc: "Notify when materials reach critical or low stock levels",
  },
  {
    id: "task-reminders",
    label: "Task due date reminders",
    desc: "Receive reminders 48 hours before task due dates",
  },
  {
    id: "site-updates",
    label: "Site update submitted",
    desc: "Notify when team members submit site updates",
  },
  {
    id: "document-uploads",
    label: "Document uploads",
    desc: "Notify when new documents are uploaded to your projects",
  },
  {
    id: "budget-thresholds",
    label: "Budget threshold alerts",
    desc: "Notify when a project budget reaches 80% spent",
  },
  {
    id: "weekly-summary",
    label: "Weekly project summary",
    desc: "Receive a weekly summary every Monday morning",
  },
]

export default function SettingsPage(_props: Props) {
  const { user, setUser } = useAuth()
  const [tab, setTab] = useState<Tab>("profile")

  return (
    <div>
      <div className="border-b border-slate-200 bg-white px-4 py-5 sm:px-8">
        <h1 className="text-xl font-bold tracking-tight text-slate-900">
          Settings
        </h1>
        <p className="mt-0.5 text-sm text-slate-500">
          Manage your account and workspace preferences
        </p>
      </div>

      <div className="flex flex-col md:flex-row">
        <nav
          className="shrink-0 border-b border-slate-200 bg-white px-4 py-3 md:w-52 md:border-b-0 md:border-r md:py-4"
          aria-label="Settings sections"
        >
          <div className="flex gap-1.5 overflow-x-auto md:flex-col md:overflow-visible">
            {TABS.map((t) => (
              <button
                key={t.id}
                onClick={() => setTab(t.id)}
                aria-pressed={tab === t.id}
                className={`shrink-0 rounded px-3 py-2 text-left text-sm transition-colors md:w-full md:rounded-none ${
                  tab === t.id
                    ? "bg-slate-100 font-semibold text-slate-900 md:border-l-2 md:border-orange-600"
                    : "text-slate-600 hover:bg-slate-50"
                }`}
              >
                {t.label}
              </button>
            ))}
          </div>
        </nav>

        <div className="flex-1 p-4 sm:p-8">
          {tab === "profile" && <ProfileTab user={user} setUser={setUser} />}
          {tab === "organisation" && <OrganisationTab />}
          {tab === "notifications" && <NotificationsTab />}
          {tab === "security" && <SecurityTab />}
        </div>
      </div>
    </div>
  )
}

// ── Profile ───────────────────────────────────────────────────────────────────

function ProfileTab({
  user,
  setUser,
}: {
  user: User | null
  setUser: (u: User) => void
}) {
  const [form, setForm] = useState({
    name: "",
    email: "",
    phone: "",
    jobTitle: "",
  })
  const [error, setError] = useState("")
  const [saved, setSaved] = useState("")
  const [saving, setSaving] = useState(false)

  useEffect(() => {
    if (user)
      setForm({
        name: user.name,
        email: user.email,
        phone: user.phone,
        jobTitle: user.jobTitle,
      })
  }, [user])

  const submit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError("")
    setSaved("")
    if (!form.name.trim()) {
      setError("Name is required.")
      return
    }
    if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(form.email)) {
      setError("Enter a valid email address.")
      return
    }
    setSaving(true)
    try {
      const data = await api.patch<{ user: User }>("/api/me", form)
      setUser(data.user)
      setSaved("Your profile has been updated.")
    } catch (err) {
      setError(
        err instanceof ApiRequestError
          ? err.message
          : "Could not save your profile.",
      )
    } finally {
      setSaving(false)
    }
  }

  if (!user) return null

  return (
    <div className="max-w-lg">
      {error && (
        <div className="mb-4">
          <ErrorBanner message={error} />
        </div>
      )}
      {saved && (
        <div className="mb-4">
          <SuccessBanner message={saved} />
        </div>
      )}
      <form
        onSubmit={submit}
        noValidate
        className="rounded border border-slate-200 bg-white p-6"
      >
        <h2 className="mb-5 border-b border-slate-100 pb-3 text-sm font-semibold text-slate-900">
          Personal information
        </h2>
        <div className="mb-5 flex items-center gap-4">
          <span
            className="flex h-14 w-14 items-center justify-center rounded-full text-lg font-bold text-white"
            style={{ background: user.color }}
            aria-hidden="true"
          >
            {user.initials}
          </span>
          <p className="text-xs text-slate-400">
            Your initials avatar is generated from your name.
          </p>
        </div>
        <div className="mb-4 grid grid-cols-1 gap-4 sm:grid-cols-2">
          <div>
            <label htmlFor="pf-name" className={labelClass}>
              Full name{" "}
              <span className="text-red-600" aria-hidden="true">
                *
              </span>
            </label>
            <input
              id="pf-name"
              className={inputClass}
              value={form.name}
              onChange={(e) => setForm((f) => ({ ...f, name: e.target.value }))}
            />
          </div>
          <div>
            <label htmlFor="pf-title" className={labelClass}>
              Role / title
            </label>
            <input
              id="pf-title"
              className={inputClass}
              value={form.jobTitle}
              onChange={(e) =>
                setForm((f) => ({ ...f, jobTitle: e.target.value }))
              }
            />
          </div>
        </div>
        <div className="mb-4">
          <label htmlFor="pf-email" className={labelClass}>
            Email address{" "}
            <span className="text-red-600" aria-hidden="true">
              *
            </span>
          </label>
          <input
            id="pf-email"
            type="email"
            className={inputClass}
            value={form.email}
            onChange={(e) => setForm((f) => ({ ...f, email: e.target.value }))}
          />
        </div>
        <div className="mb-5">
          <label htmlFor="pf-phone" className={labelClass}>
            Phone number
          </label>
          <input
            id="pf-phone"
            type="tel"
            className={inputClass}
            value={form.phone}
            onChange={(e) => setForm((f) => ({ ...f, phone: e.target.value }))}
          />
        </div>
        <Button type="submit" loading={saving}>
          Save changes
        </Button>
      </form>
    </div>
  )
}

// ── Organisation ──────────────────────────────────────────────────────────────

interface Org {
  name: string
  businessType: string
  country: string
  currency: string
}

function OrganisationTab() {
  const { user } = useAuth()
  const [org, setOrg] = useState<Org | null>(null)
  const [error, setError] = useState("")
  const [saved, setSaved] = useState("")
  const [saving, setSaving] = useState(false)

  useEffect(() => {
    api
      .get<{ organisation: Org }>("/api/organisation")
      .then((d) => setOrg(d.organisation))
      .catch((e) => setError(e.message))
  }, [])

  const submit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!org) return
    setError("")
    setSaved("")
    setSaving(true)
    try {
      const data = await api.patch<{ organisation: Org }>(
        "/api/organisation",
        org,
      )
      setOrg(data.organisation)
      setSaved("Organisation details saved.")
    } catch (err) {
      setError(
        err instanceof ApiRequestError
          ? err.message
          : "Could not save organisation details.",
      )
    } finally {
      setSaving(false)
    }
  }

  if (error && !org) return <ErrorBanner message={error} />

  return (
    <div className="max-w-lg">
      {error && (
        <div className="mb-4">
          <ErrorBanner message={error} />
        </div>
      )}
      {saved && (
        <div className="mb-4">
          <SuccessBanner message={saved} />
        </div>
      )}
      {!org ? (
        <LoadingBlock />
      ) : (
        <form
          onSubmit={submit}
          noValidate
          className="rounded border border-slate-200 bg-white p-6"
        >
          <h2 className="mb-5 border-b border-slate-100 pb-3 text-sm font-semibold text-slate-900">
            Organisation details
          </h2>
          <div className="mb-4">
            <label htmlFor="org-name" className={labelClass}>
              Organisation name
            </label>
            <input
              id="org-name"
              className={inputClass}
              value={org.name}
              onChange={(e) => setOrg({ ...org, name: e.target.value })}
              disabled={user?.role !== "admin"}
            />
          </div>
          <div className="mb-4">
            <label htmlFor="org-type" className={labelClass}>
              Business type
            </label>
            <select
              id="org-type"
              className={inputClass}
              value={org.businessType}
              onChange={(e) => setOrg({ ...org, businessType: e.target.value })}
              disabled={user?.role !== "admin"}
            >
              {[
                "Construction Company",
                "Project Developer",
                "Engineering Firm",
                "Quantity Surveying Practice",
              ].map((t) => (
                <option key={t}>{t}</option>
              ))}
            </select>
          </div>
          <div className="mb-5 grid grid-cols-1 gap-4 sm:grid-cols-2">
            <div>
              <label htmlFor="org-country" className={labelClass}>
                Country
              </label>
              <select
                id="org-country"
                className={inputClass}
                value={org.country}
                onChange={(e) => setOrg({ ...org, country: e.target.value })}
                disabled={user?.role !== "admin"}
              >
                {["Kenya", "Uganda", "Tanzania"].map((c) => (
                  <option key={c}>{c}</option>
                ))}
              </select>
            </div>
            <div>
              <label htmlFor="org-currency" className={labelClass}>
                Default currency
              </label>
              <select
                id="org-currency"
                className={inputClass}
                value={org.currency}
                onChange={(e) => setOrg({ ...org, currency: e.target.value })}
                disabled={user?.role !== "admin"}
              >
                {[
                  "KES (Kenyan Shilling)",
                  "UGX (Ugandan Shilling)",
                  "USD (US Dollar)",
                ].map((c) => (
                  <option key={c}>{c}</option>
                ))}
              </select>
            </div>
          </div>
          {user?.role === "admin" ? (
            <Button type="submit" loading={saving}>
              Save changes
            </Button>
          ) : (
            <p className="text-xs text-slate-400">
              Only workspace admins can change organisation details.
            </p>
          )}
        </form>
      )}
    </div>
  )
}

// ── Notifications ─────────────────────────────────────────────────────────────

function NotificationsTab() {
  const [prefs, setPrefs] = useState<Record<string, boolean> | null>(null)
  const [error, setError] = useState("")
  const [saved, setSaved] = useState(false)

  useEffect(() => {
    api
      .get<{ preferences: Record<string, boolean> }>("/api/me/preferences")
      .then((d) => setPrefs(d.preferences))
      .catch((e) => setError(e.message))
  }, [])

  const toggle = async (id: string) => {
    if (!prefs) return
    const next = { ...prefs, [id]: !prefs[id] }
    setPrefs(next)
    setSaved(false)
    try {
      const data = await api.patch<{ preferences: Record<string, boolean> }>(
        "/api/me/preferences",
        { [id]: next[id] },
      )
      setPrefs(data.preferences)
      setSaved(true)
    } catch (err) {
      setError(
        err instanceof Error ? err.message : "Could not save the preference.",
      )
      setPrefs(prefs) // revert optimistic toggle
    }
  }

  if (error && !prefs) return <ErrorBanner message={error} />
  if (!prefs) return <LoadingBlock />

  return (
    <div className="max-w-lg">
      {error && (
        <div className="mb-4">
          <ErrorBanner message={error} />
        </div>
      )}
      {saved && (
        <div className="mb-4">
          <SuccessBanner message="Preferences saved." />
        </div>
      )}
      <div className="rounded border border-slate-200 bg-white p-6">
        <h2 className="mb-5 border-b border-slate-100 pb-3 text-sm font-semibold text-slate-900">
          Notification preferences
        </h2>
        <div className="space-y-1">
          {NOTIFICATION_SETTINGS.map((item) => (
            <div
              key={item.id}
              className="flex items-start justify-between gap-4 py-3"
            >
              <div className="min-w-0">
                <label
                  htmlFor={`pref-${item.id}`}
                  className="text-sm font-medium text-slate-700"
                >
                  {item.label}
                </label>
                <p className="mt-0.5 text-xs text-slate-400">{item.desc}</p>
              </div>
              <button
                id={`pref-${item.id}`}
                role="switch"
                aria-checked={prefs[item.id]}
                aria-label={item.label}
                onClick={() => toggle(item.id)}
                className={`relative h-5 w-9 shrink-0 rounded-full transition-colors focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-slate-800 ${
                  prefs[item.id] ? "bg-slate-900" : "bg-slate-200"
                }`}
              >
                <span
                  className="absolute top-0.5 h-4 w-4 rounded-full bg-white shadow transition-[left]"
                  style={{ left: prefs[item.id] ? 18 : 2 }}
                  aria-hidden="true"
                />
              </button>
            </div>
          ))}
        </div>
      </div>
    </div>
  )
}

// ── Security ──────────────────────────────────────────────────────────────────

function SecurityTab() {
  const [form, setForm] = useState({
    currentPassword: "",
    newPassword: "",
    confirmPassword: "",
  })
  const [error, setError] = useState("")
  const [saved, setSaved] = useState("")
  const [saving, setSaving] = useState(false)

  const submit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError("")
    setSaved("")
    if (!form.currentPassword || !form.newPassword) {
      setError("Fill in both password fields.")
      return
    }
    if (form.newPassword !== form.confirmPassword) {
      setError("New passwords do not match.")
      return
    }
    setSaving(true)
    try {
      await api.post("/api/me/password", {
        currentPassword: form.currentPassword,
        newPassword: form.newPassword,
      })
      // The API invalidates all sessions on password change; force a fresh sign-in.
      window.location.href = "/login"
    } catch (err) {
      setError(
        err instanceof ApiRequestError
          ? err.message
          : "Could not change the password.",
      )
      setSaving(false)
    }
  }

  return (
    <div className="max-w-lg">
      {error && (
        <div className="mb-4">
          <ErrorBanner message={error} />
        </div>
      )}
      {saved && (
        <div className="mb-4">
          <SuccessBanner message={saved} />
        </div>
      )}
      <form
        onSubmit={submit}
        noValidate
        className="rounded border border-slate-200 bg-white p-6"
      >
        <h2 className="mb-1 border-b border-slate-100 pb-3 text-sm font-semibold text-slate-900">
          Change password
        </h2>
        <p className="mb-4 mt-2 text-xs text-slate-400">
          Changing your password signs out all other sessions.
        </p>
        <div className="space-y-4">
          <div>
            <label htmlFor="sec-cur" className={labelClass}>
              Current password
            </label>
            <input
              id="sec-cur"
              type="password"
              autoComplete="current-password"
              className={inputClass}
              value={form.currentPassword}
              onChange={(e) =>
                setForm((f) => ({ ...f, currentPassword: e.target.value }))
              }
            />
          </div>
          <div>
            <label htmlFor="sec-new" className={labelClass}>
              New password
            </label>
            <input
              id="sec-new"
              type="password"
              autoComplete="new-password"
              className={inputClass}
              value={form.newPassword}
              onChange={(e) =>
                setForm((f) => ({ ...f, newPassword: e.target.value }))
              }
            />
            <p className="mt-1 text-xs text-slate-400">
              At least 8 characters, including a letter and a number.
            </p>
          </div>
          <div>
            <label htmlFor="sec-conf" className={labelClass}>
              Confirm new password
            </label>
            <input
              id="sec-conf"
              type="password"
              autoComplete="new-password"
              className={inputClass}
              value={form.confirmPassword}
              onChange={(e) =>
                setForm((f) => ({ ...f, confirmPassword: e.target.value }))
              }
            />
          </div>
          <Button type="submit" loading={saving}>
            Update password
          </Button>
        </div>
      </form>
    </div>
  )
}
