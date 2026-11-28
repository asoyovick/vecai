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
import {
  inviteProjectMember,
  listProjectMembers,
  removeProjectMember,
  type ProjectMember,
} from "@/lib/members";
import {
  createMaterial,
  deleteMaterial,
  getMaterialsSummary,
  listMaterials,
  updateMaterial,
  type Material,
  type MaterialSummary,
} from "@/lib/materials";
import {
  createReport,
  deleteReport,
  getReportSummary,
  listReports,
  type DailyReport,
  type ReportSummary,
} from "@/lib/reports";

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
  const [members, setMembers] = useState<ProjectMember[] | null>(null);
  const [membersLoading, setMembersLoading] = useState(false);
  const [memberError, setMemberError] = useState<string | null>(null);
  const [inviteUserId, setInviteUserId] = useState("");
  const [inviteRole, setInviteRole] = useState("engineer");
  const [materials, setMaterials] = useState<Material[] | null>(null);
  const [materialsSummary, setMaterialsSummary] = useState<MaterialSummary | null>(null);
  const [materialsLoading, setMaterialsLoading] = useState(false);
  const [materialError, setMaterialError] = useState<string | null>(null);
  const [reports, setReports] = useState<DailyReport[] | null>(null);
  const [reportSummary, setReportSummary] = useState<ReportSummary | null>(null);
  const [reportsLoading, setReportsLoading] = useState(false);
  const [reportError, setReportError] = useState<string | null>(null);
  const [reportForm, setReportForm] = useState({
    date: new Date().toISOString().slice(0, 10),
    completed_work: "",
    workers: "0",
    remarks: "",
  });
  const [savingReport, setSavingReport] = useState(false);

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

  useEffect(() => {
    if (activeTab === "team" && project) {
      loadMembers();
    }
    if (activeTab === "materials" && project) {
      loadMaterials();
    }
    if (activeTab === "reports" && project) {
      loadReports();
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [activeTab, project?.id]);

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

  async function loadMembers() {
    if (!project) return;
    setMembersLoading(true);
    setMemberError(null);
    try {
      const token = getToken()!;
      const data = await listProjectMembers(token, project.id);
      setMembers(data);
    } catch (err) {
      setMemberError(err instanceof ApiError ? err.message : "Failed to load members.");
    } finally {
      setMembersLoading(false);
    }
  }

  async function loadMaterials() {
    if (!project) return;
    setMaterialsLoading(true);
    setMaterialError(null);
    try {
      const token = getToken()!;
      const [items, summary] = await Promise.all([
        listMaterials(token, project.id),
        getMaterialsSummary(token, project.id),
      ]);
      setMaterials(items);
      setMaterialsSummary(summary);
    } catch (err) {
      setMaterialError(err instanceof ApiError ? err.message : "Failed to load materials.");
    } finally {
      setMaterialsLoading(false);
    }
  }

  async function loadReports() {
    if (!project) return;
    setReportsLoading(true);
    setReportError(null);
    try {
      const token = getToken()!;
      const [items, summary] = await Promise.all([
        listReports(token, project.id),
        getReportSummary(token, project.id),
      ]);
      setReports(items);
      setReportSummary(summary);
    } catch (err) {
      setReportError(err instanceof ApiError ? err.message : "Failed to load reports.");
    } finally {
      setReportsLoading(false);
    }
  }

  async function handleCreateReport(e: React.FormEvent) {
    e.preventDefault();
    if (!project) return;
    setSavingReport(true);
    setReportError(null);
    try {
      const token = getToken()!;
      await createReport(token, project.id, {
        date: reportForm.date,
        completed_work: reportForm.completed_work,
        workers: Number(reportForm.workers) || 0,
        remarks: reportForm.remarks || undefined,
      });
      setReportForm({
        date: new Date().toISOString().slice(0, 10),
        completed_work: "",
        workers: "0",
        remarks: "",
      });
      await loadReports();
    } catch (err) {
      setReportError(err instanceof ApiError ? err.message : "Failed to create report.");
    } finally {
      setSavingReport(false);
    }
  }

  async function handleDeleteReport(reportId: string) {
    if (!project) return;
    if (!confirm("Delete this daily report?")) return;
    setReportError(null);
    try {
      const token = getToken()!;
      await deleteReport(token, project.id, reportId);
      await loadReports();
    } catch (err) {
      setReportError(err instanceof ApiError ? err.message : "Failed to delete report.");
    }
  }

  async function handleInviteMember(e: React.FormEvent) {
    e.preventDefault();
    if (!project || !inviteUserId.trim()) return;
    setMemberError(null);
    try {
      const token = getToken()!;
      await inviteProjectMember(token, project.id, inviteUserId.trim(), inviteRole);
      setInviteUserId("");
      await loadMembers();
    } catch (err) {
      setMemberError(err instanceof ApiError ? err.message : "Failed to invite member.");
    }
  }

  async function handleRemoveMember(memberId: string) {
    if (!project) return;
    if (!confirm("Remove this member from the project?")) return;
    setMemberError(null);
    try {
      const token = getToken()!;
      await removeProjectMember(token, project.id, memberId);
      await loadMembers();
    } catch (err) {
      setMemberError(err instanceof ApiError ? err.message : "Failed to remove member.");
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
              <h3 className="text-lg font-bold text-[#000080] mb-4">Team members</h3>
              {memberError && (
                <p role="alert" className="mb-4 rounded-xl bg-red-50 border border-red-100 px-4 py-3 text-sm text-red-700">
                  {memberError}
                </p>
              )}
              <form onSubmit={handleInviteMember} className="mb-6 flex flex-wrap gap-2 items-end">
                <div className="flex-1 min-w-[200px]">
                  <label className="block text-sm font-medium text-slate-700 mb-1">User ID</label>
                  <input
                    className={inputClasses}
                    value={inviteUserId}
                    onChange={(e) => setInviteUserId(e.target.value)}
                    placeholder="Enter user UUID"
                    required
                  />
                </div>
                <div className="w-48">
                  <label className="block text-sm font-medium text-slate-700 mb-1">Role</label>
                  <select className={inputClasses} value={inviteRole} onChange={(e) => setInviteRole(e.target.value)}>
                    <option value="engineer">Engineer</option>
                    <option value="architect">Architect</option>
                    <option value="contractor">Contractor</option>
                    <option value="quantity_surveyor">Quantity Surveyor</option>
                    <option value="supplier">Supplier</option>
                  </select>
                </div>
                <button type="submit" className="rounded-xl bg-[#F28500] px-4 py-2.5 text-sm font-bold text-white hover:bg-[#F28500]/90">
                  Invite
                </button>
              </form>
              {membersLoading && <p className="text-sm text-slate-500">Loading members…</p>}
              {!membersLoading && members && members.length === 0 && (
                <p className="text-sm text-slate-500">No team members yet.</p>
              )}
              {!membersLoading && members && members.length > 0 && (
                <ul className="divide-y divide-slate-200">
                  {members.map((m) => (
                    <li key={m.id} className="py-3 flex items-center justify-between gap-4">
                      <div>
                        <p className="font-medium text-slate-800">{m.name || m.user_id}</p>
                        <p className="text-sm text-slate-500">{m.email || ""} {m.role ? `· ${m.role}` : ""}</p>
                      </div>
                      <button
                        type="button"
                        onClick={() => handleRemoveMember(m.id)}
                        className="text-sm text-red-600 hover:text-red-800 font-medium"
                      >
                        Remove
                      </button>
                    </li>
                  ))}
                </ul>
              )}
            </div>
          )}
          {activeTab === "materials" && (
            <div>
              <h3 className="text-lg font-bold text-[#000080] mb-4">Materials</h3>
              {materialError && (
                <p role="alert" className="mb-4 rounded-xl bg-red-50 border border-red-100 px-4 py-3 text-sm text-red-700">
                  {materialError}
                </p>
              )}
              {materialsSummary && (
                <div className="mb-6 grid gap-4 sm:grid-cols-3">
                  <div className="rounded-xl bg-slate-50 p-4">
                    <p className="text-xs text-slate-500">Total items</p>
                    <p className="text-xl font-bold text-slate-800">{materialsSummary.total_items ?? 0}</p>
                  </div>
                  <div className="rounded-xl bg-slate-50 p-4">
                    <p className="text-xs text-slate-500">Total quantity</p>
                    <p className="text-xl font-bold text-slate-800">{materialsSummary.total_quantity ?? 0}</p>
                  </div>
                  <div className="rounded-xl bg-slate-50 p-4">
                    <p className="text-xs text-slate-500">Total cost</p>
                    <p className="text-xl font-bold text-slate-800">{materialsSummary.total_cost ?? 0}</p>
                  </div>
                </div>
              )}
              {materialsLoading && <p className="text-sm text-slate-500">Loading materials…</p>}
              {!materialsLoading && materials && materials.length === 0 && (
                <p className="text-sm text-slate-500">No materials recorded yet.</p>
              )}
              {!materialsLoading && materials && materials.length > 0 && (
                <div className="overflow-x-auto">
                  <table className="min-w-full text-sm">
                    <thead>
                      <tr className="text-left text-slate-500 border-b">
                        <th className="py-2 pr-4">Name</th>
                        <th className="py-2 pr-4">Qty</th>
                        <th className="py-2 pr-4">Unit</th>
                        <th className="py-2 pr-4">Unit cost</th>
                        <th className="py-2 pr-4">Supplier</th>
                        <th className="py-2 pr-4">Delivery date</th>
                      </tr>
                    </thead>
                    <tbody>
                      {materials.map((m) => (
                        <tr key={m.id} className="border-b border-slate-100">
                          <td className="py-2 pr-4 font-medium text-slate-800">{m.material_name}</td>
                          <td className="py-2 pr-4">{m.quantity}</td>
                          <td className="py-2 pr-4">{m.unit}</td>
                          <td className="py-2 pr-4">{m.unit_cost}</td>
                          <td className="py-2 pr-4">{m.supplier || "—"}</td>
                          <td className="py-2 pr-4">{m.delivery_date || "—"}</td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              )}
            </div>
          )}
          {activeTab === "boq" && (
            <div>
              <p className="text-sm text-slate-500">BOQ management UI (issue #13) will be shown here.</p>
            </div>
          )}
          {activeTab === "reports" && (
            <div>
              <h3 className="text-lg font-bold text-[#000080] mb-4">Daily reports</h3>
              {reportError && (
                <p role="alert" className="mb-4 rounded-xl bg-red-50 border border-red-100 px-4 py-3 text-sm text-red-700">
                  {reportError}
                </p>
              )}

              {reportSummary && (
                <div className="mb-6 grid gap-4 sm:grid-cols-4">
                  <div className="rounded-xl bg-slate-50 p-4">
                    <p className="text-xs text-slate-500">Total reports</p>
                    <p className="text-xl font-bold text-slate-800">{reportSummary.total_reports}</p>
                  </div>
                  <div className="rounded-xl bg-slate-50 p-4">
                    <p className="text-xs text-slate-500">Average workers</p>
                    <p className="text-xl font-bold text-slate-800">{reportSummary.average_workers.toFixed(1)}</p>
                  </div>
                  <div className="rounded-xl bg-slate-50 p-4">
                    <p className="text-xs text-slate-500">Peak workers</p>
                    <p className="text-xl font-bold text-slate-800">{reportSummary.peak_workers}</p>
                  </div>
                  <div className="rounded-xl bg-slate-50 p-4">
                    <p className="text-xs text-slate-500">Latest report</p>
                    <p className="text-xl font-bold text-slate-800">{reportSummary.latest_report_date || "—"}</p>
                  </div>
                </div>
              )}

              <form onSubmit={handleCreateReport} className="mb-6 grid gap-3 rounded-xl border border-slate-200 p-4 sm:grid-cols-2">
                <div>
                  <label className="block text-xs font-medium text-slate-600 mb-1">Date</label>
                  <input
                    type="date"
                    className={inputClasses}
                    value={reportForm.date}
                    onChange={(e) => setReportForm({ ...reportForm, date: e.target.value })}
                    required
                  />
                </div>
                <div>
                  <label className="block text-xs font-medium text-slate-600 mb-1">Workers present</label>
                  <input
                    type="number"
                    min={0}
                    className={inputClasses}
                    value={reportForm.workers}
                    onChange={(e) => setReportForm({ ...reportForm, workers: e.target.value })}
                  />
                </div>
                <div className="sm:col-span-2">
                  <label className="block text-xs font-medium text-slate-600 mb-1">Completed work</label>
                  <textarea
                    className={inputClasses}
                    rows={3}
                    value={reportForm.completed_work}
                    onChange={(e) => setReportForm({ ...reportForm, completed_work: e.target.value })}
                    required
                  />
                </div>
                <div className="sm:col-span-2">
                  <label className="block text-xs font-medium text-slate-600 mb-1">Remarks (optional)</label>
                  <input
                    className={inputClasses}
                    value={reportForm.remarks}
                    onChange={(e) => setReportForm({ ...reportForm, remarks: e.target.value })}
                  />
                </div>
                <div className="sm:col-span-2">
                  <button
                    type="submit"
                    disabled={savingReport}
                    className="rounded-xl bg-[#F28500] px-4 py-2 text-sm font-semibold text-white hover:bg-[#d97706] disabled:opacity-50"
                  >
                    {savingReport ? "Saving…" : "Add report"}
                  </button>
                </div>
              </form>

              {reportsLoading && <p className="text-sm text-slate-500">Loading reports…</p>}
              {!reportsLoading && reports && reports.length === 0 && (
                <p className="text-sm text-slate-500">No daily reports yet.</p>
              )}
              {!reportsLoading && reports && reports.length > 0 && (
                <div className="space-y-3">
                  {reports.map((r) => (
                    <div key={r.id} className="rounded-xl border border-slate-200 p-4">
                      <div className="flex items-start justify-between gap-4">
                        <div>
                          <p className="font-medium text-slate-800">{r.report_date}</p>
                          <p className="text-sm text-slate-500">{r.workers_present} workers present</p>
                        </div>
                        <button
                          type="button"
                          onClick={() => handleDeleteReport(r.id)}
                          className="text-xs text-red-600 hover:underline"
                        >
                          Delete
                        </button>
                      </div>
                      <p className="mt-2 text-sm text-slate-700 whitespace-pre-wrap">{r.completed_work}</p>
                      {r.remarks && <p className="mt-1 text-sm text-slate-500">Remarks: {r.remarks}</p>}
                    </div>
                  ))}
                </div>
              )}
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
