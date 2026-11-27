"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { useParams, useRouter } from "next/navigation";
import { ApiError } from "@/lib/api";
import { clearSession, getToken, restoreSession, roleLabel, type User } from "@/lib/auth";
import {
  deleteProject,
  getProject,
  PROJECT_STATUSES,
  updateProject,
  type Project,
  type UpdateProjectInput,
} from "@/lib/projects";

const inputClasses =
  "w-full rounded-xl border border-slate-200 bg-white px-4 py-2.5 text-sm text-slate-900 outline-none placeholder:text-slate-400 focus:border-[#F28500] focus:ring-2 focus:ring-[#F28500]/20";

type Tab = "overview" | "team" | "materials" | "boq" | "reports" | "documents";

export default function ProjectDetailPage() {
  const params = useParams();
  const router = useRouter();
  const projectId = params.id as string;

  const [user, setUser] = useState<User | null>(null);
  const [project, setProject] = useState<Project | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [activeTab, setActiveTab] = useState<Tab>("overview");
  const [editing, setEditing] = useState(false);
  const [saving, setSaving] = useState(false);
  const [editForm, setEditForm] = useState<UpdateProjectInput>({});

  useEffect(() => {
    let cancelled = false;
    (async () => {
      const u = await restoreSession();
      if (cancelled) return;
      if (!u) {
        router.replace("/login");
        return;
      }
      setUser(u);
      try {
        const token = getToken()!;
        const p = await getProject(token, projectId);
        if (!cancelled) {
          setProject(p);
          setEditForm({
            name: p.name,
            description: p.description,
            location: p.location,
            client_name: p.client_name,
            status: p.status,
            start_date: p.start_date,
            expected_end_date: p.expected_end_date,
          });
        }
      } catch (err) {
        if (cancelled) return;
        if (err instanceof ApiError && err.status === 401) {
          clearSession();
          router.replace("/login");
          return;
        }
        setError(err instanceof ApiError ? err.message : "Failed to load project.");
      } finally {
        if (!cancelled) setLoading(false);
      }
    })();
    return () => {
      cancelled = true;
    };
  }, [projectId, router]);

  async function handleSave() {
    if (!project) return;
    setSaving(true);
    setError(null);
    try {
      const token = getToken()!;
      const updated = await updateProject(token, project.id, editForm);
      setProject(updated);
      setEditing(false);
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Failed to update project.");
    } finally {
      setSaving(false);
    }
  }

  async function handleArchive() {
    if (!project) return;
    if (!confirm("Archive this project? You can change status back later by editing.")) return;
    setSaving(true);
    setError(null);
    try {
      const token = getToken()!;
      const updated = await updateProject(token, project.id, { status: "Archived" });
      setProject(updated);
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Failed to archive project.");
    } finally {
      setSaving(false);
    }
  }

  async function handleDelete() {
    if (!project) return;
    if (!confirm("Permanently delete this project? This cannot be undone.")) return;
    setSaving(true);
    setError(null);
    try {
      const token = getToken()!;
      await deleteProject(token, project.id);
      router.push("/dashboard");
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Failed to delete project.");
      setSaving(false);
    }
  }

  if (loading) {
    return (
      <div className="min-h-screen bg-slate-50 flex items-center justify-center">
        <p className="text-sm text-slate-500">Loading…</p>
      </div>
    );
  }

  if (!project) {
    return (
      <div className="min-h-screen bg-slate-50 flex items-center justify-center">
        <div className="text-center">
          <p className="text-sm text-slate-500">{error || "Project not found."}</p>
          <Link href="/dashboard" className="mt-4 inline-block text-[#F28500] font-medium">
            ← Back to dashboard
          </Link>
        </div>
      </div>
    );
  }

  const tabs: { id: Tab; label: string }[] = [
    { id: "overview", label: "Overview" },
    { id: "team", label: "Team" },
    { id: "materials", label: "Materials" },
    { id: "boq", label: "BOQ" },
    { id: "reports", label: "Reports" },
    { id: "documents", label: "Documents" },
  ];

  return (
    <div className="min-h-screen bg-slate-50">
      <header className="bg-[#000080] text-white">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-4 flex items-center justify-between">
          <Link href="/dashboard" className="text-sm text-white/80 hover:text-white">
            ← Dashboard
          </Link>
          <span className="text-sm text-white/80">{user ? `${user.fullname} · ${roleLabel(user.role)}` : ""}</span>
        </div>
      </header>

      <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {error && (
          <p role="alert" className="mb-4 rounded-xl bg-red-50 border border-red-100 px-4 py-3 text-sm text-red-700">
            {error}
          </p>
        )}

        <div className="flex flex-col sm:flex-row sm:items-start sm:justify-between gap-4 mb-6">
          <div>
            <h1 className="text-2xl font-bold text-[#000080]">{project.name}</h1>
            <p className="mt-1 text-sm text-slate-500">{project.location || "No location"}</p>
          </div>
          <div className="flex flex-wrap gap-2">
            <button
              type="button"
              onClick={() => setEditing((v) => !v)}
              className="rounded-lg border border-slate-300 px-4 py-2 text-sm font-medium text-slate-700 hover:bg-slate-50"
            >
              {editing ? "Cancel edit" : "Edit"}
            </button>
            <button
              type="button"
              onClick={handleArchive}
              disabled={saving || project.status === "Archived"}
              className="rounded-lg border border-amber-300 px-4 py-2 text-sm font-medium text-amber-700 hover:bg-amber-50 disabled:opacity-50"
            >
              Archive
            </button>
            <button
              type="button"
              onClick={handleDelete}
              disabled={saving}
              className="rounded-lg border border-red-300 px-4 py-2 text-sm font-medium text-red-700 hover:bg-red-50 disabled:opacity-50"
            >
              Delete
            </button>
          </div>
        </div>

        {editing && (
          <div className="mb-6 rounded-2xl border border-slate-200 bg-white p-6 shadow-sm">
            <h2 className="text-lg font-bold text-[#000080] mb-4">Edit project</h2>
            <div className="grid gap-4 sm:grid-cols-2">
              <label className="block">
                <span className="mb-1 block text-sm font-medium text-slate-700">Name</span>
                <input className={inputClasses} value={editForm.name ?? ""} onChange={(e) => setEditForm((f) => ({ ...f, name: e.target.value }))} />
              </label>
              <label className="block">
                <span className="mb-1 block text-sm font-medium text-slate-700">Location</span>
                <input className={inputClasses} value={editForm.location ?? ""} onChange={(e) => setEditForm((f) => ({ ...f, location: e.target.value }))} />
              </label>
              <label className="block sm:col-span-2">
                <span className="mb-1 block text-sm font-medium text-slate-700">Description</span>
                <textarea className={inputClasses} rows={3} value={editForm.description ?? ""} onChange={(e) => setEditForm((f) => ({ ...f, description: e.target.value }))} />
              </label>
              <label className="block">
                <span className="mb-1 block text-sm font-medium text-slate-700">Client</span>
                <input className={inputClasses} value={editForm.client_name ?? ""} onChange={(e) => setEditForm((f) => ({ ...f, client_name: e.target.value }))} />
              </label>
              <label className="block">
                <span className="mb-1 block text-sm font-medium text-slate-700">Status</span>
                <select className={inputClasses} value={editForm.status ?? project.status} onChange={(e) => setEditForm((f) => ({ ...f, status: e.target.value }))}>
                  {PROJECT_STATUSES.map((s) => (
                    <option key={s} value={s}>{s}</option>
                  ))}
                </select>
              </label>
              <label className="block">
                <span className="mb-1 block text-sm font-medium text-slate-700">Start date</span>
                <input type="date" className={inputClasses} value={editForm.start_date ?? ""} onChange={(e) => setEditForm((f) => ({ ...f, start_date: e.target.value }))} />
              </label>
              <label className="block">
                <span className="mb-1 block text-sm font-medium text-slate-700">Expected end date</span>
                <input type="date" className={inputClasses} value={editForm.expected_end_date ?? ""} onChange={(e) => setEditForm((f) => ({ ...f, expected_end_date: e.target.value }))} />
              </label>
            </div>
            <div className="mt-4 flex gap-2">
              <button type="button" onClick={handleSave} disabled={saving} className="rounded-xl bg-[#F28500] px-4 py-2 text-sm font-bold text-white disabled:opacity-50">
                {saving ? "Saving…" : "Save changes"}
              </button>
              <button type="button" onClick={() => setEditing(false)} className="rounded-xl border border-slate-300 px-4 py-2 text-sm font-medium text-slate-700">
                Cancel
              </button>
            </div>
          </div>
        )}

        <div className="border-b border-slate-200 mb-6">
          <nav className="flex gap-4 overflow-x-auto">
            {tabs.map((t) => (
              <button
                key={t.id}
                type="button"
                onClick={() => setActiveTab(t.id)}
                className={`whitespace-nowrap pb-3 text-sm font-medium border-b-2 transition-colors ${
                  activeTab === t.id
                    ? "border-[#F28500] text-[#F28500]"
                    : "border-transparent text-slate-500 hover:text-slate-700"
                }`}
              >
                {t.label}
              </button>
            ))}
          </nav>
        </div>

        <div className="rounded-2xl border border-slate-200 bg-white p-6 shadow-sm">
          {activeTab === "overview" && (
            <div className="space-y-4">
              <div>
                <h3 className="text-sm font-medium text-slate-400">Description</h3>
                <p className="mt-1 text-slate-700">{project.description || "—"}</p>
              </div>
              <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
                <div><h3 className="text-sm font-medium text-slate-400">Status</h3><p className="mt-1 text-slate-700">{project.status}</p></div>
                <div><h3 className="text-sm font-medium text-slate-400">Client</h3><p className="mt-1 text-slate-700">{project.client_name || "—"}</p></div>
                <div><h3 className="text-sm font-medium text-slate-400">Start date</h3><p className="mt-1 text-slate-700">{project.start_date || "—"}</p></div>
                <div><h3 className="text-sm font-medium text-slate-400">Expected end</h3><p className="mt-1 text-slate-700">{project.expected_end_date || "—"}</p></div>
              </div>
            </div>
          )}
          {activeTab === "team" && (
            <div>
              <p className="text-sm text-slate-500">Team management is available in the Team issue (see issue #10). This tab will list and manage members.</p>
            </div>
          )}
          {activeTab === "materials" && (
            <div>
              <p className="text-sm text-slate-500">Materials tracking UI (issue #11) will be shown here.</p>
            </div>
          )}
          {activeTab === "boq" && (
            <div>
              <p className="text-sm text-slate-500">BOQ management UI (issue #13) will be shown here.</p>
            </div>
          )}
          {activeTab === "reports" && (
            <div>
              <p className="text-sm text-slate-500">Daily reports UI (issue #12 frontend, #18 in list) will be shown here.</p>
            </div>
          )}
          {activeTab === "documents" && (
            <div>
              <p className="text-sm text-slate-500">Documents upload/download UI (issue #14) will be shown here.</p>
            </div>
          )}
        </div>
      </main>
    </div>
  );
}
