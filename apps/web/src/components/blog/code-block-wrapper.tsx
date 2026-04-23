"use client";

// Design Ref: §2.4 T4 — Wraps shiki-rendered <pre> with hover copy button.
// Fires blog_copy_code on click (FR-32).
//
// Width/overflow behavior (fix for O3 + O4):
//   - shiki injects `<pre style="background-color:#24292e">` as inline style;
//     inline > className, so `bg-surface-2` was losing. We strip shiki's inline
//     background here so our Tailwind class wins. globals.css also enforces it
//     with !important as a safety net for any shiki output elsewhere.
//   - The wrapper `<div>` gets `min-w-0` so when this wrapper sits inside a
//     CSS grid cell, the grid track cannot grow to accommodate a long code
//     line. Combined with `overflow-x-auto` on the <pre>, long lines scroll
//     *inside* the block rather than blowing out the page width.

import { useRef, useState, type CSSProperties } from "react";
import { track } from "@/lib/track";

type Props = {
  children?: React.ReactNode;
  slug?: string;
  "data-language"?: string;
} & React.HTMLAttributes<HTMLPreElement>;

function stripBackground(style?: CSSProperties): CSSProperties | undefined {
  if (!style) return undefined;
  // Strip shiki-injected background so our Tailwind `bg-surface-2` wins.
  // We keep color/other tokens so shiki's foreground theming stays intact.
  const { background: _bg, backgroundColor: _bgc, ...rest } = style;
  return rest;
}

export function CodeBlockWrapper({
  children,
  slug,
  style,
  className,
  ...rest
}: Props) {
  const ref = useRef<HTMLPreElement>(null);
  const [copied, setCopied] = useState(false);

  async function copy() {
    const code = ref.current?.innerText ?? "";

    // Fire analytics regardless of clipboard success — intent-to-copy is the
    // engagement signal we care about, and clipboard can fail in non-secure
    // contexts or headless automation even when the user clicked.
    const lang =
      ref.current?.className.match(/language-(\w+)/)?.[1] ??
      ref.current?.querySelector("[class*=language-]")
        ?.className.match(/language-(\w+)/)?.[1] ??
      "plaintext";
    if (slug) {
      track("blog_copy_code", { slug, language: lang });
    }

    try {
      await navigator.clipboard.writeText(code);
      setCopied(true);
      setTimeout(() => setCopied(false), 1500);
    } catch {
      // Clipboard unavailable (e.g. non-secure context). Fail silently.
    }
  }

  // Merge shiki's className (`shiki github-dark ...`) with our layout
  // classes — keep shiki for theme scoping AND get our overflow/border/bg.
  // Without this, `{...rest}` spread would overwrite our className prop
  // and kill `overflow-x-auto`, letting long code lines blow out the page.
  const mergedClassName = [
    "overflow-x-auto rounded-lg border border-border bg-surface-2 p-4 text-sm",
    className,
  ]
    .filter(Boolean)
    .join(" ");

  return (
    <div className="group relative my-6 min-w-0">
      <pre
        ref={ref}
        {...rest}
        className={mergedClassName}
        style={stripBackground(style)}
      >
        {children}
      </pre>
      <button
        type="button"
        onClick={copy}
        aria-label="Copy code"
        className="absolute right-3 top-3 rounded border border-border bg-surface/80 px-2 py-1 text-xs text-muted opacity-0 backdrop-blur-sm transition-opacity hover:border-accent/40 hover:text-foreground group-hover:opacity-100 focus:opacity-100"
      >
        {copied ? "✓ Copied" : "Copy"}
      </button>
    </div>
  );
}
