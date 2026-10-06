import { useEffect, useRef, useState } from "react"
import {
  LayoutDashboard,
  FolderKanban,
  CheckSquare,
  Package,
  FileText,
  Users,
  Truck,
  BarChart2,
  Sparkles,
  Settings,
  X,
  LogOut,
  ChevronDown,
} from "lucide-react"
import type { NavigateFn, Page } from "../App"
import { useAuth } from "../lib/auth"

interface NavItem {
  id: Page
  label: string
  icon: React.ComponentType<{ size?: number; strokeWidth?: number }>
}

const mainNav: NavItem[] = [
  { id: "dashboard", label: "Dashboard", icon: LayoutDashboard },
  { id: "projects", label: "Projects", icon: FolderKanban },
  { id: "tasks", label: "Tasks", icon: CheckSquare },
  { id: "materials", label: "Materials", icon: Package },
  { id: "documents", label: "Documents", icon: FileText },
  { id: "team", label: "Team", icon: Users },
  { id: "suppliers", label: "Suppliers", icon: Truck },
  { id: "reports", label: "Reports", icon: BarChart2 },
]

interface Props {
  currentPage: Page
  onNavigate: NavigateFn
  mobileOpen: boolean
  onClose: () => void
  onSignOut: () => Promise<void>
}

export default function Sidebar({
  currentPage,
  onNavigate,
  mobileOpen,
  onClose,
  onSignOut,
}: Props) {
  const { user } = useAuth()
  const [menuOpen, setMenuOpen] = useState(false)
  const menuRef = useRef<HTMLDivElement>(null)

  const isActive = (id: Page) =>
    currentPage === id ||
    (id === "projects" && currentPage === "project-detail")

  // Close the profile menu on outside click / Escape.
  useEffect(() => {
    if (!menuOpen) return
    const onClick = (e: MouseEvent) => {
      if (menuRef.current && !menuRef.current.contains(e.target as Node))
        setMenuOpen(false)
    }
    const onKey = (e: KeyboardEvent) => e.key === "Escape" && setMenuOpen(false)
    document.addEventListener("mousedown", onClick)
    document.addEventListener("keydown", onKey)
    return () => {
      document.removeEventListener("mousedown", onClick)
      document.removeEventListener("keydown", onKey)
    }
  }, [menuOpen])

  const nav = (
    <aside
      className={`fixed inset-y-0 left-0 z-40 flex w-60 flex-col bg-slate-900 transition-transform duration-200 lg:static lg:translate-x-0 ${
        mobileOpen ? "translate-x-0" : "-translate-x-full"
      }`}
      aria-label="Main navigation"
    >
      {/* Logo */}
      <div className="flex items-center justify-between border-b border-white/10 px-5 py-5">
        <div className="flex items-center gap-2.5">
          <div className="flex h-8 w-8 items-center justify-center rounded bg-orange-600">
            <span className="text-sm font-bold tracking-wide text-white">
              V
            </span>
          </div>
          <div>
            <div className="text-base font-bold leading-none tracking-wide text-white">
              VECAI
            </div>
            <div className="mt-0.5 text-xs leading-none text-slate-400">
              Construction Intelligence
            </div>
          </div>
        </div>
        <button
          className="rounded p-1 text-slate-400 hover:bg-white/10 hover:text-white lg:hidden"
          onClick={onClose}
          aria-label="Close navigation"
        >
          <X size={18} />
        </button>
      </div>

      {/* Main navigation */}
      <nav className="flex-1 overflow-y-auto px-2.5 py-4">
        {mainNav.map((item) => {
          const active = isActive(item.id)
          const Icon = item.icon
          return (
            <button
              key={item.id}
              onClick={() => onNavigate(item.id)}
              aria-current={active ? "page" : undefined}
              className={`relative mb-0.5 flex w-full items-center gap-3 rounded px-3 py-2 text-left text-sm font-medium transition-colors ${
                active
                  ? "bg-white/10 text-white"
                  : "text-slate-400 hover:bg-white/5 hover:text-slate-200"
              }`}
            >
              {active && (
                <span
                  className="absolute bottom-1 left-0 top-1 w-0.5 rounded-r bg-orange-600"
                  aria-hidden="true"
                />
              )}
              <Icon size={16} strokeWidth={active ? 2.2 : 1.8} />
              {item.label}
            </button>
          )
        })}

        <div className="my-3 border-t border-white/10" />

        <button
          onClick={() => onNavigate("ai-assistant")}
          aria-current={currentPage === "ai-assistant" ? "page" : undefined}
          className={`relative flex w-full items-center gap-3 rounded px-3 py-2 text-left text-sm font-medium transition-colors ${
            currentPage === "ai-assistant"
              ? "bg-white/10 text-white"
              : "text-slate-400 hover:bg-white/5 hover:text-slate-200"
          }`}
        >
          {currentPage === "ai-assistant" && (
            <span
              className="absolute bottom-1 left-0 top-1 w-0.5 rounded-r bg-orange-600"
              aria-hidden="true"
            />
          )}
          <Sparkles size={16} strokeWidth={1.8} />
          Assistant
          <span className="ml-auto rounded bg-white/10 px-1.5 py-0.5 text-[10px] font-medium text-slate-300">
            Beta
          </span>
        </button>
      </nav>

      {/* Bottom: settings + profile */}
      <div className="border-t border-white/10 px-2.5 py-3">
        <button
          onClick={() => onNavigate("settings")}
          aria-current={currentPage === "settings" ? "page" : undefined}
          className={`mb-1 flex w-full items-center gap-3 rounded px-3 py-2 text-left text-sm font-medium transition-colors ${
            currentPage === "settings"
              ? "bg-white/10 text-white"
              : "text-slate-400 hover:bg-white/5 hover:text-slate-200"
          }`}
        >
          <Settings size={15} strokeWidth={1.8} />
          Settings
        </button>

        <div className="relative" ref={menuRef}>
          <button
            onClick={() => setMenuOpen((v) => !v)}
            aria-expanded={menuOpen}
            aria-haspopup="menu"
            className="flex w-full items-center gap-2.5 rounded px-2 py-2 text-left transition-colors hover:bg-white/5"
          >
            <span
              className="flex h-8 w-8 shrink-0 items-center justify-center rounded-full text-xs font-bold text-white"
              style={{ background: user?.color ?? "#1e3a5f" }}
              aria-hidden="true"
            >
              {user?.initials ?? "—"}
            </span>
            <span className="min-w-0 flex-1">
              <span className="block truncate text-sm font-medium leading-tight text-white">
                {user?.name}
              </span>
              <span className="block truncate text-xs leading-tight text-slate-400">
                {user?.jobTitle || user?.role}
              </span>
            </span>
            <ChevronDown
              size={14}
              className={`shrink-0 text-slate-500 transition-transform ${
                menuOpen ? "rotate-180" : ""
              }`}
              aria-hidden="true"
            />
          </button>

          {menuOpen && (
            <div
              role="menu"
              className="absolute bottom-full left-0 z-50 mb-2 w-full overflow-hidden rounded border border-slate-200 bg-white py-1 shadow-lg"
            >
              <div className="border-b border-slate-100 px-3 py-2">
                <p className="truncate text-xs font-semibold text-slate-800">
                  {user?.name}
                </p>
                <p className="truncate text-xs text-slate-500">{user?.email}</p>
              </div>
              <button
                role="menuitem"
                onClick={async () => {
                  setMenuOpen(false)
                  await onSignOut()
                }}
                className="flex w-full items-center gap-2 px-3 py-2 text-left text-sm text-slate-700 hover:bg-slate-50"
              >
                <LogOut size={14} />
                Sign out
              </button>
            </div>
          )}
        </div>
      </div>
    </aside>
  )

  return (
    <>
      {/* Mobile overlay */}
      {mobileOpen && (
        <div
          className="fixed inset-0 z-30 bg-slate-900/50 lg:hidden"
          onClick={onClose}
          aria-hidden="true"
        />
      )}
      {nav}
    </>
  )
}
