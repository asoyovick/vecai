"use client";

import { useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { AuthCard } from "@/components/auth/AuthCard";
import { ApiError } from "@/lib/api";
import { register, roleOptions, type Role } from "@/lib/auth";

const inputClasses =
  "w-full rounded-xl border border-slate-200 bg-white px-4 py-2.5 text-sm text-slate-900 outline-none placeholder:text-slate-400 focus:border-[#F28500] focus:ring-2 focus:ring-[#F28500]/20";

export default function RegisterPage() {
  const router = useRouter();
  const [form, setForm] = useState({
    fullname: "",
    email: "",
    password: "",
    phone: "",
  });
  const [role, setRole] = useState<Role>("owner");
  const [error, setError] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);

  function update(field: keyof typeof form) {
    return (e: React.ChangeEvent<HTMLInputElement>) =>
      setForm((prev) => ({ ...prev, [field]: e.target.value }));
  }

  async function handleSubmit(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault();
    if (submitting) return;
    setSubmitting(true);
    setError(null);

    try {
      await register({
        fullname: form.fullname.trim(),
        email: form.email.trim(),
        password: form.password,
        role,
        phone: form.phone.trim() || undefined,
      });
      router.push("/dashboard");
    } catch (err) {
      setError(
        err instanceof ApiError
          ? err.message
          : "Unable to create your account. Please try again.",
      );
      setSubmitting(false);
    }
  }

  return (
    <AuthCard
      title="Create your account"
      subtitle="Start planning and estimating your first project in minutes."
      footer={
        <>
          Already have an account?{" "}
          <Link href="/login" className="font-semibold text-[#F28500] hover:underline">
            Sign in
          </Link>
        </>
      }
    >
      <form className="space-y-4" onSubmit={handleSubmit} noValidate>
        {error && (
          <p
            role="alert"
            className="rounded-xl bg-red-50 border border-red-100 px-4 py-3 text-sm text-red-700"
          >
            {error}
          </p>
        )}

        <div>
          <label htmlFor="fullname" className="mb-1.5 block text-sm font-medium text-slate-700">
            Full name
          </label>
          <input
            id="fullname"
            type="text"
            required
            autoComplete="name"
            value={form.fullname}
            onChange={update("fullname")}
            placeholder="Ada Lovelace"
            className={inputClasses}
          />
        </div>

        <div>
          <label htmlFor="email" className="mb-1.5 block text-sm font-medium text-slate-700">
            Email
          </label>
          <input
            id="email"
            type="email"
            required
            autoComplete="email"
            value={form.email}
            onChange={update("email")}
            placeholder="you@company.com"
            className={inputClasses}
          />
        </div>

        <div>
          <label htmlFor="password" className="mb-1.5 block text-sm font-medium text-slate-700">
            Password
          </label>
          <input
            id="password"
            type="password"
            required
            minLength={8}
            autoComplete="new-password"
            value={form.password}
            onChange={update("password")}
            placeholder="At least 8 characters"
            className={inputClasses}
          />
        </div>

        <div>
          <label htmlFor="role" className="mb-1.5 block text-sm font-medium text-slate-700">
            I am a…
          </label>
          <select
            id="role"
            value={role}
            onChange={(e) => setRole(e.target.value as Role)}
            className={inputClasses}
          >
            {roleOptions.map((opt) => (
              <option key={opt.value} value={opt.value}>
                {opt.label}
              </option>
            ))}
          </select>
        </div>

        <div>
          <label htmlFor="phone" className="mb-1.5 block text-sm font-medium text-slate-700">
            Phone <span className="font-normal text-slate-400">(optional)</span>
          </label>
          <input
            id="phone"
            type="tel"
            autoComplete="tel"
            value={form.phone}
            onChange={update("phone")}
            placeholder="+234 800 000 0000"
            className={inputClasses}
          />
        </div>

        <button
          type="submit"
          disabled={submitting}
          className="w-full rounded-xl bg-[#F28500] px-5 py-3 text-sm font-bold text-white shadow-md transition-all hover:bg-[#F28500]/90 disabled:cursor-not-allowed disabled:opacity-60"
        >
          {submitting ? "Creating account…" : "Create account"}
        </button>
      </form>
    </AuthCard>
  );
}
