import {
  HardHat,
  CheckSquare,
  Package,
  FileText,
  BarChart2,
  Users,
  Truck,
  Shield,
  ArrowRight,
  ChevronRight,
} from "lucide-react"

interface Props {
  onGetStarted: () => void
  onSignIn: () => void
}

const features = [
  {
    icon: <CheckSquare size={20} strokeWidth={1.8} />,
    title: "Task Management",
    desc: "Track every task across your projects — from foundation to finishing. List and Kanban views built for construction workflows.",
    cls: "bg-blue-50 text-blue-700",
  },
  {
    icon: <Package size={20} strokeWidth={1.8} />,
    title: "Materials Tracking",
    desc: "Monitor stock levels, log deliveries, and get alerts before materials run out on site. Linked directly to your suppliers.",
    cls: "bg-orange-50 text-orange-700",
  },
  {
    icon: <FileText size={20} strokeWidth={1.8} />,
    title: "Document Control",
    desc: "Architectural drawings, BOQs, contracts, permits, and site reports — version-controlled and accessible from anywhere.",
    cls: "bg-green-50 text-green-700",
  },
  {
    icon: <BarChart2 size={20} strokeWidth={1.8} />,
    title: "Budget & Cost Reporting",
    desc: "Track planned vs. actual expenditure monthly and catch overruns early with clear variance analysis in KES.",
    cls: "bg-violet-50 text-violet-700",
  },
  {
    icon: <HardHat size={20} strokeWidth={1.8} />,
    title: "Digital Site Reports",
    desc: "Replace paper site diaries with structured daily updates — weather, workers, work completed, and issues logged in minutes.",
    cls: "bg-sky-50 text-sky-700",
  },
  {
    icon: <BarChart2 size={20} strokeWidth={1.8} />,
    title: "Project Intelligence",
    desc: "Ask questions in plain language. Summarise site reports, identify at-risk tasks, and compile weekly status reports from live data.",
    cls: "bg-slate-100 text-slate-700",
  },
]

const roles = [
  {
    role: "Project Managers",
    desc: "Full project oversight — schedule, budget, team, and materials — in a single workspace.",
    icon: Users,
  },
  {
    role: "Site Engineers",
    desc: "Log daily progress, flag issues, track material deliveries, and keep the project team informed.",
    icon: HardHat,
  },
  {
    role: "Quantity Surveyors",
    desc: "Manage BOQs, track cost variances, and generate budget reports from recorded project data.",
    icon: BarChart2,
  },
  {
    role: "Architects",
    desc: "Upload, version, and distribute drawings so the whole team works from the latest revision.",
    icon: FileText,
  },
  {
    role: "Site Managers",
    desc: "Submit structured daily site reports from any device and keep the office up to date.",
    icon: Shield,
  },
  {
    role: "Project Owners",
    desc: "See real progress, budget position, and milestone updates without scheduling a call.",
    icon: Truck,
  },
]

export default function LandingPage({ onGetStarted, onSignIn }: Props) {
  return (
    <div className="min-h-screen bg-white text-slate-900">
      {/* Nav */}
      <nav className="sticky top-0 z-50 border-b border-slate-100 bg-white/95 backdrop-blur">
        <div className="mx-auto flex max-w-6xl items-center justify-between px-6 py-3.5">
          <div className="flex items-center gap-2.5">
            <div className="flex h-8 w-8 items-center justify-center rounded bg-orange-600">
              <HardHat size={17} color="white" strokeWidth={2} />
            </div>
            <div>
              <div className="text-base font-bold leading-none tracking-wide">
                VECAI
              </div>
              <div className="mt-0.5 text-[10px] leading-none tracking-widest text-slate-400">
                VICTORIA EAGLES CONSTRUCT AI
              </div>
            </div>
          </div>
          <div className="flex items-center gap-3">
            <button
              onClick={onSignIn}
              className="rounded border border-slate-300 px-4 py-2 text-sm font-medium text-slate-700 transition-colors hover:border-slate-500 hover:text-slate-900"
            >
              Sign in
            </button>
            <button
              onClick={onGetStarted}
              className="hidden items-center gap-1.5 rounded bg-slate-900 px-4 py-2 text-sm font-semibold text-white transition-colors hover:bg-slate-700 sm:flex"
            >
              View demo <ArrowRight size={14} aria-hidden="true" />
            </button>
          </div>
        </div>
      </nav>

      {/* Hero */}
      <section className="relative overflow-hidden bg-slate-900">
        <img
          src="https://images.unsplash.com/photo-1504307651254-35680f356dfd?w=1800&h=900&fit=crop&auto=format"
          alt=""
          aria-hidden="true"
          className="absolute inset-0 h-full w-full object-cover opacity-15"
        />
        <div
          className="absolute inset-x-0 top-0 h-0.5 bg-orange-600"
          aria-hidden="true"
        />
        <div className="relative mx-auto max-w-6xl px-6 py-24">
          <div className="max-w-2xl">
            <h1 className="mb-5 text-4xl font-bold leading-tight tracking-tight text-white sm:text-[2.75rem]">
              Manage every construction project — from ground-breaking to
              handover.
            </h1>
            <p className="mb-8 max-w-xl text-[17px] leading-relaxed text-slate-300">
              VECAI gives Kenyan construction teams — project managers,
              engineers, QSs, architects, and site managers — one platform to
              run projects without spreadsheets and paper reports.
            </p>
            <div className="flex flex-wrap items-center gap-3">
              <button
                onClick={onGetStarted}
                className="flex items-center gap-2 rounded bg-orange-600 px-6 py-3 text-[15px] font-semibold text-white transition-colors hover:bg-orange-700"
              >
                View demo <ArrowRight size={16} aria-hidden="true" />
              </button>
              <button
                onClick={onSignIn}
                className="rounded border border-white/20 px-6 py-3 text-[15px] font-medium text-white transition-colors hover:bg-white/10"
              >
                Sign in to workspace
              </button>
            </div>
            <p className="mt-6 text-xs text-slate-400">
              Live demo environment — real data, real workflows. Sign in with
              the credentials shown on the login page.
            </p>
          </div>
        </div>
      </section>

      {/* Features */}
      <section className="px-6 py-20" id="features">
        <div className="mx-auto max-w-6xl">
          <div className="mb-12 max-w-xl">
            <p className="mb-2 text-xs font-semibold uppercase tracking-widest text-orange-700">
              Platform capabilities
            </p>
            <h2 className="mb-3 text-3xl font-bold tracking-tight">
              Every tool your site team needs, in one platform.
            </h2>
            <p className="text-[15px] leading-relaxed text-slate-500">
              VECAI replaces the spreadsheets, WhatsApp groups, and paper
              reports that slow construction projects down.
            </p>
          </div>
          <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
            {features.map((f) => (
              <div
                key={f.title}
                className="rounded border border-slate-200 bg-white p-6 transition-colors hover:border-slate-400"
              >
                <div
                  className={`mb-4 flex h-10 w-10 items-center justify-center rounded ${f.cls}`}
                  aria-hidden="true"
                >
                  {f.icon}
                </div>
                <h3 className="mb-2 text-base font-semibold">{f.title}</h3>
                <p className="text-sm leading-relaxed text-slate-500">
                  {f.desc}
                </p>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* Site update highlight */}
      <section className="border-y border-slate-100 bg-slate-50 px-6 py-20">
        <div className="mx-auto grid max-w-6xl items-center gap-12 lg:grid-cols-2">
          <div>
            <p className="mb-2 text-xs font-semibold uppercase tracking-widest text-orange-700">
              Digital site reports
            </p>
            <h2 className="mb-4 text-3xl font-bold leading-tight tracking-tight">
              Replace paper site diaries — for good.
            </h2>
            <p className="mb-6 text-[15px] leading-relaxed text-slate-500">
              Site managers submit structured daily updates from their phone —
              recording work completed, materials used, workers on site, weather
              conditions, and issues — in under five minutes.
            </p>
            <ul className="mb-8 space-y-2.5">
              {[
                "Structured daily site reports submitted from the field",
                "Materials usage tracked against stock levels",
                "Issues and delays logged for the whole team to see",
                "Full audit trail — searchable and filterable",
              ].map((item) => (
                <li
                  key={item}
                  className="flex items-start gap-3 text-sm text-slate-700"
                >
                  <span
                    className="mt-1.5 h-1.5 w-1.5 shrink-0 rounded-full bg-orange-600"
                    aria-hidden="true"
                  />
                  {item}
                </li>
              ))}
            </ul>
            <button
              onClick={onGetStarted}
              className="inline-flex items-center gap-2 rounded bg-slate-900 px-5 py-2.5 text-sm font-semibold text-white transition-colors hover:bg-slate-700"
            >
              See site updates in action{" "}
              <ChevronRight size={15} aria-hidden="true" />
            </button>
          </div>

          {/* Example report — clearly presented as an example */}
          <div className="rounded border border-slate-200 bg-white shadow-sm">
            <div className="border-b border-slate-100 px-5 py-3.5">
              <div className="flex items-center justify-between">
                <span className="text-sm font-semibold text-slate-900">
                  Site update — 25 Sep 2026
                </span>
                <span className="rounded bg-green-50 px-2 py-0.5 text-xs font-medium text-green-700">
                  Submitted
                </span>
              </div>
              <span className="text-xs text-slate-400">
                Lakeview Residence · submitted by Peter Otieno
              </span>
            </div>
            <div className="space-y-4 p-5">
              <dl className="grid grid-cols-3 gap-3 text-xs">
                {[
                  ["Weather", "Partly Cloudy · 24°C"],
                  ["Workers on site", "28"],
                  ["Materials used", "4 items"],
                ].map(([k, v]) => (
                  <div
                    key={k}
                    className="rounded border border-slate-100 bg-slate-50 p-3"
                  >
                    <dt className="text-slate-400">{k}</dt>
                    <dd className="mt-0.5 font-semibold text-slate-700">{v}</dd>
                  </div>
                ))}
              </dl>
              <div>
                <h3 className="mb-1 text-xs font-semibold text-slate-700">
                  Work completed today
                </h3>
                <p className="text-xs leading-relaxed text-slate-500">
                  Completed casting of columns C3 and C4 on ground floor.
                  Formwork struck for beam B1 — concrete quality satisfactory.
                  Backfilling completed on west elevation.
                </p>
              </div>
              <div className="rounded border border-orange-200 bg-orange-50 p-3">
                <h3 className="text-xs font-semibold text-orange-800">
                  Issue reported
                </h3>
                <p className="mt-0.5 text-xs leading-relaxed text-orange-900">
                  Steel Y16 delivery from Devki delayed. Expected 29 Sep. Team
                  adjusted to C5–C6 formwork.
                </p>
              </div>
            </div>
          </div>
        </div>
      </section>

      {/* For teams */}
      <section className="px-6 py-20" id="for-teams">
        <div className="mx-auto max-w-6xl">
          <div className="mb-12 max-w-xl">
            <p className="mb-2 text-xs font-semibold uppercase tracking-widest text-orange-700">
              Built for every role
            </p>
            <h2 className="text-3xl font-bold tracking-tight">
              One platform. Every stakeholder.
            </h2>
          </div>
          <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
            {roles.map((r) => {
              const Icon = r.icon
              return (
                <div
                  key={r.role}
                  className="rounded border border-slate-100 bg-slate-50 p-5"
                >
                  <div className="mb-3 flex items-center gap-3">
                    <span
                      className="flex h-9 w-9 items-center justify-center rounded bg-slate-900"
                      aria-hidden="true"
                    >
                      <Icon
                        size={16}
                        className="text-orange-500"
                        strokeWidth={1.8}
                      />
                    </span>
                    <span className="text-sm font-semibold">{r.role}</span>
                  </div>
                  <p className="text-sm leading-relaxed text-slate-500">
                    {r.desc}
                  </p>
                </div>
              )
            })}
          </div>
        </div>
      </section>

      {/* CTA */}
      <section className="border-t border-slate-100 bg-slate-50 px-6 py-20">
        <div className="mx-auto max-w-xl text-center">
          <h2 className="mb-4 text-3xl font-bold tracking-tight">
            Ready to run your construction projects properly?
          </h2>
          <p className="mb-8 text-[15px] leading-relaxed text-slate-500">
            Open the demo workspace or sign in to an existing one. VECAI is
            built for Kenyan construction teams — from single-site residential
            to multi-project infrastructure programmes.
          </p>
          <div className="flex flex-wrap justify-center gap-3">
            <button
              onClick={onGetStarted}
              className="flex items-center gap-2 rounded bg-orange-600 px-6 py-3 text-[15px] font-semibold text-white transition-colors hover:bg-orange-700"
            >
              View demo platform <ArrowRight size={16} aria-hidden="true" />
            </button>
            <button
              onClick={onSignIn}
              className="rounded border border-slate-300 bg-white px-6 py-3 text-[15px] font-medium text-slate-700 transition-colors hover:border-slate-500"
            >
              Sign in
            </button>
          </div>
        </div>
      </section>

      {/* Footer */}
      <footer className="bg-slate-900 px-6 py-10">
        <div className="mx-auto flex max-w-6xl flex-col gap-6 sm:flex-row sm:items-center sm:justify-between">
          <div className="flex items-center gap-2.5">
            <div className="flex h-8 w-8 items-center justify-center rounded bg-orange-600">
              <HardHat size={15} color="white" strokeWidth={2} />
            </div>
            <div>
              <div className="text-sm font-bold text-white">VECAI</div>
              <div className="text-[10px] tracking-widest text-slate-400">
                VICTORIA EAGLES CONSTRUCT AI
              </div>
            </div>
          </div>
          <p className="text-xs text-slate-500">
            © 2026 Victoria Eagles Construct AI · Nairobi, Kenya
          </p>
        </div>
      </footer>
    </div>
  )
}
