// Shared UI primitives: one visual language for buttons, inputs, badges,
// spinners, banners, and state blocks (loading / empty / error).

import type {
  ReactNode,
  ButtonHTMLAttributes,
  InputHTMLAttributes,
  SelectHTMLAttributes,
  TextareaHTMLAttributes,
} from "react"

export const inputClass =
  "w-full px-3 py-2 text-sm text-slate-900 bg-white border border-slate-300 rounded outline-none transition-colors " +
  "placeholder:text-slate-400 focus:border-slate-800 focus:ring-2 focus:ring-slate-900/10 disabled:bg-slate-100 disabled:text-slate-400"

export const labelClass = "block text-sm font-medium text-slate-700 mb-1.5"

export const selectClass =
  inputClass +
  ' pr-8 appearance-none bg-[url("data:image/svg+xml;charset=utf-8,%3Csvg%20xmlns%3D%22http%3A%2F%2Fwww.w3.org%2F2000%2Fsvg%22%20width%3D%2212%22%20height%3D%2212%22%20viewBox%3D%220%200%2012%2012%22%3E%3Cpath%20d%3D%22M2%204l4%204%204-4%22%20fill%3D%22none%22%20stroke%3D%22%2364748b%22%20stroke-width%3D%221.5%22%2F%3E%3C%2Fsvg%3E")] bg-[length:12px] bg-[right_0.6rem_center] bg-no-repeat'

// ── Buttons ───────────────────────────────────────────────────────────────────

type Variant = "primary" | "secondary" | "ghost"

interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: Variant
  loading?: boolean
}

const buttonBase =
  "inline-flex items-center justify-center gap-1.5 px-4 py-2 text-sm font-semibold rounded transition-colors " +
  "focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-slate-800 " +
  "disabled:opacity-50 disabled:cursor-not-allowed"

const buttonVariants: Record<Variant, string> = {
  primary: "bg-slate-900 text-white hover:bg-slate-700 active:bg-slate-900",
  secondary:
    "border border-slate-300 text-slate-700 bg-white hover:bg-slate-50 hover:border-slate-400",
  ghost: "text-slate-600 hover:bg-slate-100 hover:text-slate-900",
}

export function Button({
  variant = "primary",
  loading,
  className = "",
  children,
  disabled,
  ...rest
}: ButtonProps) {
  return (
    <button
      className={`${buttonBase} ${buttonVariants[variant]} ${className}`}
      disabled={disabled || loading}
      {...rest}
    >
      {loading && <Spinner size={13} />}
      {children}
    </button>
  )
}

export function Spinner({
  size = 16,
  className = "",
}: {
  size?: number
  className?: string
}) {
  return (
    <svg
      width={size}
      height={size}
      viewBox="0 0 24 24"
      fill="none"
      className={`animate-spin ${className}`}
      aria-hidden="true"
    >
      <circle
        cx="12"
        cy="12"
        r="9"
        stroke="currentColor"
        strokeWidth="2.5"
        opacity="0.25"
      />
      <path
        d="M21 12a9 9 0 0 0-9-9"
        stroke="currentColor"
        strokeWidth="2.5"
        strokeLinecap="round"
      />
    </svg>
  )
}

// ── Badges ────────────────────────────────────────────────────────────────────

export function StatusBadge({ status }: { status: string }) {
  const map: Record<string, string> = {
    Construction: "bg-blue-50 text-blue-800",
    Foundation: "bg-orange-50 text-orange-800",
    Roofing: "bg-green-50 text-green-800",
    Tendering: "bg-slate-100 text-slate-600",
    Design: "bg-violet-50 text-violet-800",
    Finishing: "bg-emerald-50 text-emerald-800",
    Completed: "bg-green-50 text-green-800",
    Handover: "bg-sky-50 text-sky-800",
  }
  return (
    <span
      className={`inline-flex items-center px-2 py-0.5 rounded text-xs font-medium ${map[status] || "bg-slate-100 text-slate-700"}`}
    >
      {status}
    </span>
  )
}

export function TaskBadge({ status }: { status: string }) {
  const map: Record<string, string> = {
    "not-started": "bg-slate-100 text-slate-600",
    "in-progress": "bg-blue-50 text-blue-800",
    blocked: "bg-red-50 text-red-700",
    completed: "bg-green-50 text-green-800",
  }
  const label: Record<string, string> = {
    "not-started": "Not started",
    "in-progress": "In progress",
    blocked: "Blocked",
    completed: "Completed",
  }
  return (
    <span
      className={`inline-flex items-center px-2 py-0.5 rounded text-xs font-medium ${map[status] || "bg-slate-100"}`}
    >
      {label[status] || status}
    </span>
  )
}

export function MaterialBadge({ status }: { status: string }) {
  const map: Record<string, { label: string; cls: string }> = {
    critical: { label: "Critical", cls: "bg-red-50 text-red-700" },
    low: { label: "Low stock", cls: "bg-orange-50 text-orange-800" },
    ordered: { label: "On order", cls: "bg-sky-50 text-sky-800" },
    ok: { label: "In stock", cls: "bg-green-50 text-green-800" },
  }
  const s = map[status] || { label: status, cls: "bg-slate-100 text-slate-700" }
  return (
    <span
      className={`inline-flex items-center px-2 py-0.5 rounded text-xs font-medium whitespace-nowrap ${s.cls}`}
    >
      {s.label}
    </span>
  )
}

// ── Progress ──────────────────────────────────────────────────────────────────

export function Progress({
  value,
  className = "",
}: {
  value: number
  className?: string
}) {
  const v = Math.max(0, Math.min(100, value))
  return (
    <div
      className={`h-1.5 w-full bg-slate-200 rounded-full overflow-hidden ${className}`}
      role="progressbar"
      aria-valuenow={v}
      aria-valuemin={0}
      aria-valuemax={100}
    >
      <div
        className="h-full bg-slate-900 rounded-full transition-[width] duration-300"
        style={{ width: `${v}%` }}
      />
    </div>
  )
}

// ── Avatars ───────────────────────────────────────────────────────────────────

export function Avatar({
  member,
  size = 28,
}: {
  member: { name: string; initials: string; color: string }
  size?: number
}) {
  return (
    <span
      title={member.name}
      className="inline-flex items-center justify-center rounded-full text-white font-semibold shrink-0"
      style={{
        width: size,
        height: size,
        background: member.color,
        fontSize: Math.round(size * 0.36),
      }}
    >
      {member.initials}
    </span>
  )
}

// ── Banners & state blocks ────────────────────────────────────────────────────

export function ErrorBanner({
  message,
  onRetry,
}: {
  message: string
  onRetry?: () => void
}) {
  return (
    <div
      className="flex items-start gap-2.5 rounded border border-red-200 bg-red-50 px-4 py-3"
      role="alert"
    >
      <span
        className="mt-0.5 inline-block h-2 w-2 shrink-0 rounded-full bg-red-600"
        aria-hidden="true"
      />
      <div className="flex-1">
        <p className="text-sm font-medium text-red-800">{message}</p>
        {onRetry && (
          <button
            onClick={onRetry}
            className="mt-1 text-xs font-semibold text-red-700 underline underline-offset-2 hover:text-red-900"
          >
            Try again
          </button>
        )}
      </div>
    </div>
  )
}

export function SuccessBanner({ message }: { message: string }) {
  return (
    <div
      className="flex items-center gap-2.5 rounded border border-green-200 bg-green-50 px-4 py-3"
      role="status"
    >
      <span
        className="inline-block h-2 w-2 shrink-0 rounded-full bg-green-600"
        aria-hidden="true"
      />
      <p className="text-sm font-medium text-green-800">{message}</p>
    </div>
  )
}

export function LoadingBlock({ label = "Loading…" }: { label?: string }) {
  return (
    <div
      className="flex items-center justify-center gap-2.5 py-16 text-sm text-slate-500"
      role="status"
      aria-live="polite"
    >
      <Spinner size={15} />
      {label}
    </div>
  )
}

export function EmptyState({
  title,
  hint,
  action,
}: {
  title: string
  hint?: string
  action?: ReactNode
}) {
  return (
    <div className="py-14 text-center">
      <p className="text-sm font-semibold text-slate-700">{title}</p>
      {hint && <p className="mt-1 text-sm text-slate-500">{hint}</p>}
      {action && <div className="mt-4 flex justify-center">{action}</div>}
    </div>
  )
}
