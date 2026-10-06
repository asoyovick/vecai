import { useState } from "react"
import { HardHat, Lock, Mail, AlertCircle } from "lucide-react"
import { useAuth } from "../lib/auth"
import { ApiRequestError } from "../lib/api"

interface Props {
  onBack?: () => void
  onRegister?: () => void
}

export default function LoginPage({ onBack, onRegister }: Props) {
  const { login } = useAuth()
  const [email, setEmail] = useState("david.ochieng@vecai.co.ke")
  const [password, setPassword] = useState("")
  const [error, setError] = useState("")
  const [loading, setLoading] = useState(false)

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!email || !password) {
      setError("Please enter your email and password.")
      return
    }
    setError("")
    setLoading(true)
    try {
      await login(email, password)
      // Shell re-renders automatically once the session user is set.
    } catch (err) {
      setError(
        err instanceof ApiRequestError
          ? err.message
          : "Could not sign in. Please try again.",
      )
      setLoading(false)
    }
  }

  return (
    <div className="flex min-h-screen bg-white">
      {/* Left panel — brand */}
      <div className="relative hidden w-[420px] shrink-0 flex-col justify-between overflow-hidden bg-slate-900 p-10 lg:flex">
        <img
          src="https://images.unsplash.com/photo-1504307651254-35680f356dfd?w=960&h=1200&fit=crop&auto=format"
          alt=""
          aria-hidden="true"
          className="absolute inset-0 h-full w-full object-cover opacity-10"
        />
        <div className="relative">
          <div className="mb-10 flex items-center gap-3">
            <div className="flex h-10 w-10 items-center justify-center rounded bg-orange-600">
              <HardHat size={20} color="white" strokeWidth={2} />
            </div>
            <div>
              <div className="text-lg font-bold tracking-wide text-white">
                VECAI
              </div>
              <div className="text-xs tracking-widest text-slate-400">
                VICTORIA EAGLES CONSTRUCT AI
              </div>
            </div>
          </div>
          <h1 className="mb-4 text-3xl font-bold leading-tight tracking-tight text-white">
            Construction
            <br />
            project intelligence.
          </h1>
          <p className="max-w-[320px] text-[15px] leading-relaxed text-slate-400">
            Manage projects, teams, materials, budgets and documents — all in
            one place. Built for Kenyan construction professionals.
          </p>
        </div>
        <div className="relative">
          <div className="space-y-4">
            {[
              "Real-time project tracking and site updates",
              "Materials management with stock alerts",
              "Document control with version history",
              "Project intelligence across live data",
            ].map((f) => (
              <div key={f} className="flex items-start gap-3">
                <div
                  className="mt-1.5 h-3.5 w-1 shrink-0 rounded-sm bg-orange-600"
                  aria-hidden="true"
                />
                <span className="text-sm leading-relaxed text-slate-400">
                  {f}
                </span>
              </div>
            ))}
          </div>
        </div>
      </div>

      {/* Right panel — form */}
      <div className="flex flex-1 items-center justify-center p-6 sm:p-8">
        <div className="w-full max-w-[400px]">
          <div className="mb-8 flex items-center justify-between">
            <div className="flex items-center gap-2 lg:hidden">
              <div className="flex h-8 w-8 items-center justify-center rounded bg-orange-600">
                <HardHat size={16} color="white" strokeWidth={2} />
              </div>
              <span className="text-lg font-bold text-slate-900">VECAI</span>
            </div>
            {onBack && (
              <button
                onClick={onBack}
                className="ml-auto rounded text-sm font-medium text-slate-500 hover:text-slate-900 focus-visible:outline-2 focus-visible:outline-slate-800"
              >
                ← Back to website
              </button>
            )}
          </div>

          <h2 className="mb-1 text-2xl font-bold tracking-tight text-slate-900">
            Sign in to VECAI
          </h2>
          <p className="mb-7 text-sm text-slate-500">
            Enter your credentials to access your workspace.
          </p>

          {error && (
            <div
              className="mb-5 flex items-center gap-2 rounded border border-red-200 bg-red-50 px-3.5 py-3"
              role="alert"
            >
              <AlertCircle size={15} color="#b91c1c" />
              <span className="text-sm text-red-800">{error}</span>
            </div>
          )}

          <form onSubmit={handleSubmit} noValidate>
            <div className="mb-4">
              <label
                htmlFor="login-email"
                className="mb-1.5 block text-sm font-medium text-slate-700"
              >
                Email address
              </label>
              <div className="relative">
                <Mail
                  size={15}
                  className="pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 text-slate-400"
                  aria-hidden="true"
                />
                <input
                  id="login-email"
                  type="email"
                  autoComplete="email"
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  placeholder="you@organisation.co.ke"
                  className="w-full rounded border border-slate-300 bg-white py-2.5 pl-9 pr-4 text-sm text-slate-900 outline-none transition-colors placeholder:text-slate-400 focus:border-slate-800 focus:ring-2 focus:ring-slate-900/10"
                />
              </div>
            </div>

            <div className="mb-5">
              <label
                htmlFor="login-password"
                className="mb-1.5 block text-sm font-medium text-slate-700"
              >
                Password
              </label>
              <div className="relative">
                <Lock
                  size={15}
                  className="pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 text-slate-400"
                  aria-hidden="true"
                />
                <input
                  id="login-password"
                  type="password"
                  autoComplete="current-password"
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  placeholder="••••••••"
                  className="w-full rounded border border-slate-300 bg-white py-2.5 pl-9 pr-4 text-sm text-slate-900 outline-none transition-colors placeholder:text-slate-400 focus:border-slate-800 focus:ring-2 focus:ring-slate-900/10"
                />
              </div>
            </div>

            <button
              type="submit"
              disabled={loading}
              className="flex w-full items-center justify-center gap-2 rounded bg-slate-900 py-2.5 text-sm font-semibold text-white transition-colors hover:bg-slate-700 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-slate-800 disabled:cursor-wait disabled:opacity-60"
            >
              {loading ? "Signing in…" : "Sign in"}
            </button>
          </form>

          <div className="mt-6 rounded border border-slate-200 bg-slate-50 p-3.5">
            <p className="text-xs font-semibold text-slate-600">
              Demo environment
            </p>
            <p className="mt-0.5 text-xs leading-relaxed text-slate-500">
              Sign in with{" "}
              <code className="rounded bg-white px-1 py-0.5 font-mono text-[11px]">
                david.ochieng@vecai.co.ke
              </code>{" "}
              · password{" "}
              <code className="rounded bg-white px-1 py-0.5 font-mono text-[11px]">
                demo1234
              </code>
            </p>
          </div>

          {onRegister && (
            <p className="mt-6 text-center text-sm text-slate-500">
              New to VECAI?{" "}
              <button
                onClick={onRegister}
                className="font-semibold text-slate-900 underline-offset-2 hover:underline"
              >
                Create an account
              </button>
            </p>
          )}

          <p className="mt-8 text-center text-xs text-slate-400">
            VECAI · Victoria Eagles Construct AI · Nairobi, Kenya
          </p>
        </div>
      </div>
    </div>
  )
}
