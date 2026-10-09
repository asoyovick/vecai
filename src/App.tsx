import { useCallback, useEffect, useState } from "react"
import Sidebar from "./components/Sidebar"
import Topbar from "./components/Topbar"
import LandingPage from "./pages/LandingPage"
import LoginPage from "./pages/LoginPage"
import RegisterPage from "./pages/RegisterPage"
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

export type Page = "landing" | "login" | "register" | "dashboard" | "projects" | "project-detail" | "tasks" | "materials" | "documents" | "site-update" | "ai-assistant" | "reports" | "team" | "suppliers" | "settings"

export type NavigateFn = (page: Page, projectId?: number) => void

// ── URL routing ───────────────────────────────────────────────────────────────
// `/` is the public landing page. `/login` is public too. Authenticated
// workspace pages live under `/app`.

function routeFromPath(pathname: string): { page: Page; projectId?: number } {
  if (pathname === "/" || pathname === "") return { page: "landing" }
  if (pathname === "/login") return { page: "login" }
  if (pathname === "/register") return { page: "register" }
  const projectMatch = pathname.match(/^\/app\/projects\/(\d+)/)
  if (projectMatch)
    return { page: "project-detail", projectId: Number(projectMatch[1]) }
  if (pathname === "/app/projects" || pathname === "/app/projects/")
    return { page: "projects" }
  if (pathname === "/app/tasks") return { page: "tasks" }
  if (pathname === "/app/materials") return { page: "materials" }
  if (pathname === "/app/documents") return { page: "documents" }
  if (pathname === "/app/site-updates") return { page: "site-update" }
  if (pathname === "/app/assistant") return { page: "ai-assistant" }
  if (pathname === "/app/reports") return { page: "reports" }
  if (pathname === "/app/team") return { page: "team" }
  if (pathname === "/app/suppliers") return { page: "suppliers" }
  if (pathname === "/app/settings") return { page: "settings" }
  if (
    pathname === "/app" ||
    pathname === "/app/" ||
    pathname.startsWith("/app/")
  ) {
    return { page: "dashboard" }
  }
  return { page: "landing" }
}

function pathForRoute(page: Page, projectId?: number): string {
  switch (page) {
    case "landing":
      return "/"
    case "login":
      return "/login"
    case "register":
      return "/register"
    case "dashboard":
      return "/app"
    case "projects":
      return "/app/projects"
    case "project-detail":
      return projectId !== undefined
        ? `/app/projects/${projectId}`
        : "/app/projects"
    case "tasks":
      return "/app/tasks"
    case "materials":
      return "/app/materials"
    case "documents":
      return "/app/documents"
    case "site-update":
      return "/app/site-updates"
    case "ai-assistant":
      return "/app/assistant"
    case "reports":
      return "/app/reports"
    case "team":
      return "/app/team"
    case "suppliers":
      return "/app/suppliers"
    case "settings":
      return "/app/settings"
  }
}

function isAppPath(pathname: string): boolean {
  return pathname === "/app" || pathname.startsWith("/app/")
}

function Shell() {
  const { user, loading, logout } = useAuth()
  const [route, setRoute] = useState(() =>
    routeFromPath(window.location.pathname),
  )
  const [selectedProjectId, setSelectedProjectId] = useState<number>(
    () => routeFromPath(window.location.pathname).projectId ?? 1,
  )
  const [mobileNavOpen, setMobileNavOpen] = useState(false)
  const page = route.page

  // Close the mobile nav whenever the route changes.
  useEffect(() => setMobileNavOpen(false), [page])

  // Browser back/forward.
  useEffect(() => {
    const onPop = () => setRoute(routeFromPath(window.location.pathname))
    window.addEventListener("popstate", onPop)
    return () => window.removeEventListener("popstate", onPop)
  }, [])

  const navigate: NavigateFn = useCallback((p, projectId) => {
    if (projectId !== undefined) setSelectedProjectId(projectId)
    setRoute({ page: p, projectId })
    window.history.pushState(null, "", pathForRoute(p, projectId))
    window.scrollTo(0, 0)
  }, [])

  // After a successful login or registration, take the user into the app.
  useEffect(() => {
    if (user && (page === "login" || page === "register")) navigate("dashboard")
  }, [user, page, navigate])

  if (loading) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-slate-100">
        <LoadingBlock label="Loading workspace…" />
      </div>
    )
  }

  // Public landing page.
  if (page === "landing") {
    return (
      <LandingPage
        onGetStarted={() => navigate("register")}
        onSignIn={() => navigate("login")}
      />
    )
  }

  // Login and registration are public; /app requires a session.
  if (page === "register") {
    return (
      <RegisterPage
        onBack={() => navigate("landing")}
        onSignIn={() => navigate("login")}
      />
    )
  }
  if (
    page === "login" ||
    !user ||
    (isAppPath(window.location.pathname) && !user)
  ) {
    return (
      <LoginPage
        onBack={() => navigate("landing")}
        onRegister={() => navigate("register")}
      />
    )
  }

  return (
    <div className="flex h-screen overflow-hidden bg-slate-100">
      <Sidebar
        currentPage={page}
        onNavigate={navigate}
        mobileOpen={mobileNavOpen}
        onClose={() => setMobileNavOpen(false)}
        onSignOut={async () => {
          await logout()
          navigate("landing")
        }}
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
