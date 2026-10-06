import { useEffect, useState } from "react"
import Sidebar from "./components/Sidebar"
import Topbar from "./components/Topbar"
import LandingPage from "./pages/LandingPage"
import LoginPage from "./pages/LoginPage"
import DashboardPage from "./pages/DashboardPage"
import ProjectsPage from "./pages/ProjectsPage"
import ProjectDetailPage from "./pages/ProjectDetailPage"
import TasksPage from "./pages/TasksPage"
import MaterialsPage from "./pages/MaterialsPage"
import DocumentsPage from "./pages/DocumentsPage"
import SiteUpdatePage from "./pages/SiteUpdatePage"
import AIAssistantPage from "./pages/AIAssistantPage"
import ReportsPage from "./pages/ReportsPage"
import TeamPage from "./pages/TeamPage"
import SuppliersPage from "./pages/SuppliersPage"
import SettingsPage from "./pages/SettingsPage"
import { AuthProvider, useAuth } from "./lib/auth"
import { LoadingBlock } from "./components/ui"

export type Page = "landing" | "login" | "dashboard" | "projects" | "project-detail" | "tasks" | "materials" | "documents" | "site-update" | "ai-assistant" | "reports" | "team" | "suppliers" | "settings"

export type NavigateFn = (page: Page, projectId?: number) => void

function Shell() {
  const { user, loading, logout } = useAuth()
  const [page, setPage] = useState<Page>("dashboard")
  const [selectedProjectId, setSelectedProjectId] = useState<number>(1)
  const [mobileNavOpen, setMobileNavOpen] = useState(false)

  // Close the mobile nav whenever the route changes.
  useEffect(() => setMobileNavOpen(false), [page])

  if (loading) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-slate-100">
        <LoadingBlock label="Loading workspace…" />
      </div>
    )
  }

  const navigate: NavigateFn = (p, projectId) => {
    if (projectId !== undefined) setSelectedProjectId(projectId)
    setPage(p)
    window.scrollTo(0, 0)
  }

  if (page === "landing") {
    return (
      <LandingPage
        onGetStarted={() => setPage("login")}
        onSignIn={() => setPage("login")}
      />
    )
  }

  if (page === "login" || !user) {
    return <LoginPage onBack={() => setPage("landing")} />
  }

  return (
    <div className="flex h-screen overflow-hidden bg-slate-100">
      <Sidebar
        currentPage={page}
        onNavigate={navigate}
        mobileOpen={mobileNavOpen}
        onClose={() => setMobileNavOpen(false)}
        onSignOut={logout}
      />
      <main className="min-w-0 flex-1 overflow-y-auto">
        <Topbar page={page} onOpenNav={() => setMobileNavOpen(true)} />
        {page === "dashboard" && <DashboardPage onNavigate={navigate} />}
        {page === "projects" && <ProjectsPage onNavigate={navigate} />}
        {page === "project-detail" && (
          <ProjectDetailPage
            projectId={selectedProjectId}
            onNavigate={navigate}
          />
        )}
        {page === "tasks" && <TasksPage onNavigate={navigate} />}
        {page === "materials" && <MaterialsPage onNavigate={navigate} />}
        {page === "documents" && <DocumentsPage onNavigate={navigate} />}
        {page === "site-update" && <SiteUpdatePage onNavigate={navigate} />}
        {page === "ai-assistant" && <AIAssistantPage onNavigate={navigate} />}
        {page === "reports" && <ReportsPage onNavigate={navigate} />}
        {page === "team" && <TeamPage onNavigate={navigate} />}
        {page === "suppliers" && <SuppliersPage onNavigate={navigate} />}
        {page === "settings" && <SettingsPage onNavigate={navigate} />}
      </main>
    </div>
  )
}

export default function App() {
  return (
    <AuthProvider>
      <Shell />
    </AuthProvider>
  )
}
