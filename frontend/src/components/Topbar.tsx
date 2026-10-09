import { Menu } from "lucide-react"
import type { Page } from "../App"

const TITLES: Partial<Record<Page, string>> = {
  dashboard: "Dashboard",
  projects: "Projects",
  "project-detail": "Project",
  tasks: "Tasks",
  materials: "Materials",
  documents: "Documents",
  "site-update": "Site Update",
  "ai-assistant": "Assistant",
  reports: "Reports",
  team: "Team",
  suppliers: "Suppliers",
  settings: "Settings",
}

interface Props {
  page: Page
  onOpenNav: () => void
}

/** Mobile-only header with the nav toggle; hidden on desktop where the sidebar is persistent. */
export default function Topbar({ page, onOpenNav }: Props) {
  return (
    <header className="sticky top-0 z-20 flex items-center gap-3 border-b border-slate-200 bg-white px-4 py-3 lg:hidden">
      <button
        onClick={onOpenNav}
        className="rounded p-1.5 text-slate-600 hover:bg-slate-100 focus-visible:outline-2 focus-visible:outline-slate-800"
        aria-label="Open navigation"
      >
        <Menu size={20} />
      </button>
      <span className="text-base font-bold text-slate-900">
        {TITLES[page] ?? "VECAI"}
      </span>
    </header>
  )
}
