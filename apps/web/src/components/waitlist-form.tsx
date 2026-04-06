"use client";

import { useState, useCallback } from "react";

// Design Ref: §6.2 — Formspree waitlist with mailto fallback
export function WaitlistForm({ className = "" }: { className?: string }) {
  const [email, setEmail] = useState("");
  const [submitted, setSubmitted] = useState(false);
  const [error, setError] = useState("");

  const formspreeId = process.env.NEXT_PUBLIC_FORMSPREE_ID;

  const handleSubmit = useCallback(
    async (e: React.FormEvent) => {
      e.preventDefault();
      setError("");

      if (!email || !email.includes("@")) {
        setError("Please enter a valid email.");
        return;
      }

      if (!formspreeId) {
        window.location.href = `mailto:hello@tene.sh?subject=Cloud%20Waitlist&body=Email:%20${encodeURIComponent(email)}`;
        setSubmitted(true);
        return;
      }

      try {
        const res = await fetch(`https://formspree.io/f/${formspreeId}`, {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ email }),
        });
        if (res.ok) {
          setSubmitted(true);
        } else {
          setError("Something went wrong. Please try again.");
        }
      } catch {
        setError("Network error. Please try again.");
      }
    },
    [email, formspreeId],
  );

  if (submitted) {
    return (
      <div className={`text-center ${className}`}>
        <div className="inline-flex items-center gap-2 rounded-lg bg-accent/10 px-4 py-2 text-sm text-accent">
          <svg className="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24" strokeWidth="2">
            <path d="M5 13l4 4L19 7" />
          </svg>
          You&apos;re on the list!
        </div>
      </div>
    );
  }

  return (
    <form onSubmit={handleSubmit} className={`flex flex-col gap-2 sm:flex-row ${className}`}>
      <input
        type="email"
        value={email}
        onChange={(e) => setEmail(e.target.value)}
        placeholder="you@email.com"
        className="flex-1 rounded-lg border border-border bg-surface px-4 py-2.5 text-sm text-foreground placeholder:text-muted focus:border-accent/50 focus:outline-none"
      />
      <button
        type="submit"
        className="rounded-lg bg-accent px-5 py-2.5 text-sm font-semibold text-background transition-colors hover:bg-accent-dim"
      >
        Join waitlist
      </button>
      {error && <p className="text-xs text-red-400 sm:col-span-2">{error}</p>}
    </form>
  );
}
