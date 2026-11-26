"use client";

import { useState } from "react";
import { Container, Section } from "@/components/ui/Layout";
import { ApiError, apiPost } from "@/lib/api";

type Status = "idle" | "submitting" | "success" | "error";

/**
 * Blog "Get new posts by email" form. Posts to the Go backend's
 * POST /api/v1/newsletter/subscribe and surfaces success/error feedback
 * inline, replacing the previous placeholder that swallowed submissions.
 */
export function Newsletter() {
  const [email, setEmail] = useState("");
  const [status, setStatus] = useState<Status>("idle");
  const [feedback, setFeedback] = useState<string | null>(null);

  async function handleSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (status === "submitting") return;

    setStatus("submitting");
    setFeedback(null);

    try {
      await apiPost<{ email: string }>("/api/v1/newsletter/subscribe", {
        email: email.trim(),
      });
      setStatus("success");
      setFeedback("You're subscribed! Check your inbox to confirm.");
      setEmail("");
    } catch (error) {
      setStatus("error");
      setFeedback(
        error instanceof ApiError
          ? error.message
          : "Something went wrong. Please try again.",
      );
    }
  }

  return (
    <Section className="py-20">
      <Container>
        <div className="flex flex-col items-center gap-6 rounded-[var(--radius-card)] bg-[var(--color-background)] border border-[var(--color-border)] p-10 text-center sm:flex-row sm:justify-between sm:text-left">
          <div>
            <h3 className="text-[length:var(--text-heading-sm)] font-semibold text-[var(--color-forest)]">
              Get new posts by email
            </h3>
            <p className="mt-1.5 text-sm text-[var(--color-muted)]">
              One email a month. No noise, unsubscribe any time.
            </p>
          </div>
          <form
            className="flex w-full max-w-sm flex-col gap-2 sm:w-auto"
            onSubmit={handleSubmit}
            noValidate
          >
            <div className="flex gap-2">
              <input
                type="email"
                required
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                disabled={status === "submitting"}
                aria-label="Email address"
                aria-invalid={status === "error"}
                placeholder="you@company.com"
                className="w-full rounded-full border border-[var(--color-border)] bg-white px-4 py-2.5 text-sm text-[var(--color-text)] outline-none placeholder:text-[var(--color-muted)] focus-visible:border-[var(--color-primary)] disabled:opacity-60"
              />
              <button
                type="submit"
                disabled={status === "submitting"}
                className="shrink-0 rounded-full bg-[var(--color-primary)] px-5 py-2.5 text-sm font-medium text-white transition-colors hover:bg-[var(--color-forest)] disabled:cursor-not-allowed disabled:opacity-60"
              >
                {status === "submitting" ? "Subscribing…" : "Subscribe"}
              </button>
            </div>
            {feedback && (
              <p
                role="status"
                className={`text-xs ${
                  status === "success"
                    ? "text-[var(--color-forest)]"
                    : "text-red-600"
                }`}
              >
                {feedback}
              </p>
            )}
          </form>
        </div>
      </Container>
    </Section>
  );
}
