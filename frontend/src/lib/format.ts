// Formatting + small domain helpers shared across pages.

import type { Task } from "./api"

export function kes(n: number): string {
  return "KES " + Math.round(n).toLocaleString("en-KE")
}

export function kesShort(n: number): string {
  if (Math.abs(n) >= 1_000_000)
    return `${(n / 1_000_000).toFixed(1).replace(/\.0$/, "")}M`
  if (Math.abs(n) >= 1_000) return `${(n / 1_000).toFixed(0)}K`
  return String(Math.round(n))
}

export function fmtDate(d: string | null | undefined): string {
  if (!d) return "–"
  const date = new Date(d.length === 10 ? d + "T00:00:00" : d)
  if (Number.isNaN(date.getTime())) return "–"
  return date.toLocaleDateString("en-KE", {
    day: "numeric",
    month: "short",
    year: "numeric",
  })
}

export function fmtDateTime(d: string): string {
  const date = new Date(d.includes("T") ? d : d.replace(" ", "T") + "Z")
  if (Number.isNaN(date.getTime())) return d
  return date.toLocaleString("en-KE", {
    day: "numeric",
    month: "short",
    hour: "2-digit",
    minute: "2-digit",
  })
}

export function relTime(d: string): string {
  const date = new Date(d.includes("T") ? d : d.replace(" ", "T") + "Z")
  if (Number.isNaN(date.getTime())) return d
  const diff = Date.now() - date.getTime()
  const mins = Math.round(diff / 60000)
  if (mins < 1) return "Just now"
  if (mins < 60) return `${mins}m ago`
  const hours = Math.round(mins / 60)
  if (hours < 24) return `${hours}h ago`
  const days = Math.round(hours / 24)
  if (days < 7) return `${days}d ago`
  return fmtDate(d.slice(0, 10))
}

export function isOverdue(task: Pick<Task, "due_date" | "status">): boolean {
  if (!task.due_date || task.status === "completed") return false
  const today = new Date()
  today.setHours(0, 0, 0, 0)
  return new Date(task.due_date + "T00:00:00") < today
}

export function fmtSize(bytes: number): string {
  if (bytes >= 1024 * 1024) return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
  if (bytes >= 1024) return `${Math.round(bytes / 1024)} KB`
  return `${bytes} B`
}

export function todayISO(): string {
  const d = new Date()
  const pad = (n: number) => String(n).padStart(2, "0")
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}`
}
