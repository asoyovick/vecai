"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { ApiError } from "@/lib/api";
import {
  clearSession,
  getToken,
  restoreSession,
  roleLabel,
  type User,
} from "@/lib/auth";
import {
  createProject,
  listProjects,
  PROJECT_STATUSES,
  type Project,
} from "@/lib/projects";

const inputClasses =
  "w-full rounded-xl border border-slate-200 bg-white px-4 py-2.5 text-sm text-slate-900 outline-none placeholder:text-slate-400 focus:border-[#F28500] focus:ring-2 focus:ring-[#F28500]/20";

export default function DashboardPage() {
  const router = useRouter();
  const [user, setUser] = useState<User | null>(null);
  const [projects, setProjects] = useState<Project[] | null>(null);
  const [loadError, setLoadError] = useState<string | null>(null);
  const [showForm, setShowForm] = useState(false);

  // Client-side auth guard: validate the stored token, then load projects.
  useEffect(() => {
    let cancelled = false;

    (async () => {
      const user = await restoreSession();
      if (cancelled) return;
      if (!user) {
        router.replace("/login");
        return;
      }
      setUser(user);
      try {
        const token = getToken()!;
        const data = await listProjects(token);
        if (!cancelled) setProjects(data);
      } catch (err) {
        if (cancelled) return;
        setLoadError(
          err instanceof ApiError && err.status === 401
            ? "Your session expired. Please sign in again."
            : "Could not load your projects. Please try again.",
        );
        if (err instanceof ApiError && err.status === 401) {
          clearSession();
          router.replace("/login");
        }
      }
    })();

    return () => {
      cancelled = true;
    };
  }, [router]);

  function handleCreated(project: Project) {
    setProjects((prev) => [project, ...(prev ?? [])]);
    setShowForm(false);
  }

  function handleSignOut() {
    clearSession();
    router.push("/");
  }

  if (!user) {
    return (
      <div className="min-h-screen bg-slate-50 flex items-center justify-center">
        <p className="text-sm text-slate-500">Loading…</p>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-slate-50">
      {/* Top bar */}
      <header className="bg-[#000080] text-white">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-4 flex items-center justify-between">
          <Link href="/" className="text-xl font-black tracking-tight">
            VECAI
          </Link>
          <div className="flex items-center gap-4">
            <span className="hidden sm:block text-sm text-white/80">
              {user.fullname} · {roleLabel(user.role)}
            </span>
            <button
              type="button"
              onClick={handleSignOut}
              className="rounded-lg border border-white/20 px-3 py-1.5 text-sm font-medium text-white hover:bg-white/10"
            >
              Sign out
            </button>
          </div>
        </div>
      </header>

      <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-10">
        <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
          <div>
            <h1 className="text-2xl font-bold text-[#000080]">Your projects</h1>
            <p className="mt-1 text-sm text-slate-500">
              {projects === null
                ? "Loading projects…"
                : `${projects.length} project${projects.length === 1 ? "" : "s"}`}
            </p>
          </div>
          <button
            type="button"
            onClick={() => setShowForm((v) => !v)}
            className="rounded-xl bg-[#F28500] px-5 py-2.5 text-sm font-bold text-white shadow-md hover:bg-[#F28500]/90"
          >
            {showForm ? "Cancel" : "+ New project"}
          </button>
        </div>

        {showForm && (
          <CreateProjectForm
            token={getToken()!}
            onCreated={handleCreated}
            onCancel={() => setShowForm(false)}
          />
        )}

        {loadError && (
          <p
            role="alert"
            className="mt-6 rounded-xl bg-red-50 border border-red-100 px-4 py-3 text-sm text-red-700"
          >
            {loadError}
          </p>
        )}

        {projects && projects.length === 0 && !showForm && (
          <div className="mt-10 rounded-2xl border border-dashed border-slate-300 bg-white p-12 text-center">
            <p className="text-sm text-slate-500">
              No projects yet. Create your first one to get started.
            </p>
          </div>
        )}

        {projects && projects.length > 0 && (
          <ul className="mt-8 grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
            {projects.map((project) => (
              <li key={project.id}>
                <Link
                  href={`/dashboard/projects/${project.id}`}
                  className="block rounded-2xl bg-white border border-slate-200 p-6 shadow-sm hover:border-[#F28500]/50 hover:shadow-md transition-colors"
                >
                  <div className="flex items-center justify-between gap-2">
                    <h2 className="font-bold text-[#000080] truncate">
                      {project.name}
                    </h2>
                    <StatusBadge status={project.status} />
                  </div>
                  <p className="mt-1 text-sm text-slate-500 line-clamp-2">
                    {project.description || "No description"}
                  </p>
                  <dl className="mt-4 grid grid-cols-2 gap-2 text-xs text-slate-500">
                    <div>
                      <dt className="font-medium text-slate-400">Location</dt>
                      <dd>{project.location || "—"}</dd>
                    </div>
                    <div>
                      <dt className="font-medium text-slate-400">Client</dt>
                      <dd>{project.client_name || "—"}</dd>
                    </div>
                    <div>
                      <dt className="font-medium text-slate-400">Start</dt>
                      <dd>{project.start_date || "—"}</dd>
                    </div>
                    <div>
                      <dt className="font-medium text-slate-400">Target end</dt>
                      <dd>{project.expected_end_date || "—"}</dd>
                    </div>
                  </dl>
                </Link>
              </li>
            ))}
          </ul>
        )}
      </main>
    </div>
  );
}

function StatusBadge({ status }: { status: string }) {
  const styles: Record<string, string> = {
    Planning: "bg-amber-50 text-amber-700 border-amber-200",
    Active: "bg-emerald-50 text-emerald-700 border-emerald-200",
    Paused: "bg-slate-100 text-slate-600 border-slate-200",
    Completed: "bg-sky-50 text-sky-700 border-sky-200",
    Archived: "bg-slate-100 text-slate-400 border-slate-200",
  };
  return (
    <span
      className={`shrink-0 rounded-full border px-2.5 py-0.5 text-xs font-medium ${
        styles[status] ?? "bg-slate-100 text-slate-600 border-slate-200"
      }`}
    >
      {status}
    </span>
  );
}

function CreateProjectForm({
  token,
  onCreated,
  onCancel,
}: {
  token: string;
  onCreated: (project: Project) => void;
  onCancel: () => void;
}) {
  const [form, setForm] = useState({
    name: "",
    description: "",
    location: "",
    client_name: "",
    start_date: "",
    expected_end_date: "",
  });
  const [status, setStatus] = useState<string>("Planning");
  const [error, setError] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);

  function update(field: keyof typeof form) {
    return (
      e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>,
    ) => setForm((prev) => ({ ...prev, [field]: e.target.value }));
  }

  async function handleSubmit(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault();
    if (submitting) return;
    setSubmitting(true);
    setError(null);

    try {
      const project = await createProject(token, {
        name: form.name.trim(),
        description: form.description.trim() || undefined,
        location: form.location.trim(),
        client_name: form.client_name.trim() || undefined,
        status,
        start_date: form.start_date || undefined,
        expected_end_date: form.expected_end_date || undefined,
      });
      onCreated(project);
    } catch (err) {
      setError(
        err instanceof ApiError
          ? err.message
          : "Could not create the project. Please try again.",
      );
      setSubmitting(false);
    }
  }

  return (
    <form
      onSubmit={handleSubmit}
      noValidate
      className="mt-6 rounded-2xl bg-white border border-slate-200 p-6 shadow-sm space-y-4"
    >
      <h2 className="font-bold text-[#000080]">Create a project</h2>
      {error && (
        <p
          role="alert"
          className="rounded-xl bg-red-50 border border-red-100 px-4 py-3 text-sm text-red-700"
        >
          {error}
        </p>
      )}
      <div className="grid gap-4 sm:grid-cols-2">
        <div className="sm:col-span-2">
          <label htmlFor="proj-name" className="mb-1.5 block text-sm font-medium text-slate-700">
            Project name
          </label>
          <input
            id="proj-name"
            type="text"
            required
            value={form.name}
            onChange={update("name")}
            placeholder="e.g. Lekki Duplex"
            className={inputClasses}
          />
        </div>
        <div className="sm:col-span-2">
          <label htmlFor="proj-desc" className="mb-1.5 block text-sm font-medium text-slate-700">
            Description <span className="font-normal text-slate-400">(optional)</span>
          </label>
          <textarea
            id="proj-desc"
            rows={2}
            value={form.description}
            onChange={update("description")}
            placeholder="What are you building?"
            className={inputClasses}
          />
        </div>
        <div>
          <label htmlFor="proj-location" className="mb-1.5 block text-sm font-medium text-slate-700">
            Location
          </label>
          <input
            id="proj-location"
            type="text"
            required
            value={form.location}
            onChange={update("location")}
            placeholder="e.g. Lagos"
            className={inputClasses}
          />
        </div>
        <div>
          <label htmlFor="proj-client" className="mb-1.5 block text-sm font-medium text-slate-700">
            Client <span className="font-normal text-slate-400">(optional)</span>
          </label>
          <input
            id="proj-client"
            type="text"
            value={form.client_name}
            onChange={update("client_name")}
            placeholder="e.g. ACME Corp"
            className={inputClasses}
          />
        </div>
        <div>
          <label htmlFor="proj-status" className="mb-1.5 block text-sm font-medium text-slate-700">
            Status
          </label>
          <select
            id="proj-status"
            value={status}
            onChange={(e) => setStatus(e.target.value)}
            className={inputClasses}
          >
            {PROJECT_STATUSES.map((s) => (
              <option key={s} value={s}>
                {s}
              </option>
            ))}
          </select>
        </div>
        <div className="grid grid-cols-2 gap-4">
          <div>
            <label htmlFor="proj-start" className="mb-1.5 block text-sm font-medium text-slate-700">
              Start date
            </label>
            <input
              id="proj-start"
              type="date"
              value={form.start_date}
              onChange={update("start_date")}
              className={inputClasses}
            />
          </div>
          <div>
            <label htmlFor="proj-end" className="mb-1.5 block text-sm font-medium text-slate-700">
              Target end
            </label>
            <input
              id="proj-end"
              type="date"
              value={form.expected_end_date}
              onChange={update("expected_end_date")}
              className={inputClasses}
            />
          </div>
        </div>
      </div>
      <div className="flex justify-end gap-3 pt-2">
        <button
          type="button"
          onClick={onCancel}
          className="rounded-xl border border-slate-200 px-4 py-2.5 text-sm font-medium text-slate-600 hover:bg-slate-50"
        >
          Cancel
        </button>
        <button
          type="submit"
          disabled={submitting}
          className="rounded-xl bg-[#F28500] px-5 py-2.5 text-sm font-bold text-white shadow-md hover:bg-[#F28500]/90 disabled:cursor-not-allowed disabled:opacity-60"
        >
          {submitting ? "Creating…" : "Create project"}
        </button>
      </div>
    </form>
  );
}
