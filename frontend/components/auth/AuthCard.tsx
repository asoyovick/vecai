import Link from "next/link";
import type { ReactNode } from "react";

/**
 * Shared split layout for /register and /login: brand panel on the left,
 * form card on the right. Uses the same navy/orange palette as the Navbar.
 */
export function AuthCard({
  title,
  subtitle,
  children,
  footer,
}: {
  title: string;
  subtitle: string;
  children: ReactNode;
  footer: ReactNode;
}) {
  return (
    <div className="min-h-screen bg-slate-50 flex flex-col lg:flex-row">
      {/* Brand panel */}
      <div className="hidden lg:flex lg:w-2/5 bg-[#000080] p-12 flex-col justify-between">
        <Link href="/" className="text-white text-2xl font-black tracking-tight">
          VECAI
        </Link>
        <div>
          <h2 className="text-3xl font-bold text-white leading-tight">
            Intelligence at every level of your project.
          </h2>
          <p className="mt-4 text-white/70 text-sm leading-relaxed">
            Plan, design, estimate, and execute construction projects with AI —
            from first sketch to final supplier order.
          </p>
        </div>
        <p className="text-white/40 text-xs">
          © {new Date().getFullYear()} VECAI. All rights reserved.
        </p>
      </div>

      {/* Form panel */}
      <div className="flex-1 flex items-center justify-center px-4 py-12">
        <div className="w-full max-w-md">
          <Link
            href="/"
            className="lg:hidden text-[#000080] font-black text-xl mb-8 inline-block"
          >
            VECAI
          </Link>
          <div className="bg-white rounded-2xl border border-slate-200 shadow-sm p-8">
            <h1 className="text-2xl font-bold text-[#000080]">{title}</h1>
            <p className="mt-2 text-sm text-slate-500">{subtitle}</p>
            <div className="mt-8">{children}</div>
          </div>
          <p className="mt-6 text-center text-sm text-slate-500">{footer}</p>
        </div>
        </div>
    </div>
  );
}
