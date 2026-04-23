// Design Ref: ai-discoverability I-3 — visual breadcrumb.
// Pairs with the BreadcrumbList JSON-LD already emitted in article/comparison
// page components. Renders a lightweight trail at the top of deep pages so
// visitors (and crawlers that don't read JSON-LD) can navigate back to hubs.
//
// Positioning note: the global <Nav /> is `fixed top-0 h-14 z-50`, which
// occupies 56px at the top of the viewport. This breadcrumb sits in normal
// flow at the top of <main>, so we add pt-20 (mobile) / sm:pt-24 (desktop)
// to clear the fixed nav. Hero sections that follow this breadcrumb were
// updated in the same change to use a smaller top padding (pt-8 instead of
// pt-28) since the breadcrumb now handles the nav-clearance.

import Link from "next/link";

export type Crumb = {
  label: string;
  href?: string;
};

type Props = {
  items: Crumb[];
};

export function Breadcrumb({ items }: Props) {
  return (
    <nav
      aria-label="Breadcrumb"
      className="mx-auto max-w-3xl px-4 pt-20 pb-3 text-sm sm:px-6 sm:pt-24"
    >
      <ol className="flex flex-wrap items-center gap-2 text-muted">
        {items.map((item, i) => (
          <li key={i} className="flex items-center gap-2 min-w-0">
            {i > 0 && <span className="text-muted/50">/</span>}
            {item.href ? (
              <Link
                href={item.href}
                className="truncate hover:text-accent hover:underline"
              >
                {item.label}
              </Link>
            ) : (
              <span
                aria-current="page"
                className="truncate text-foreground"
                title={item.label}
              >
                {item.label}
              </span>
            )}
          </li>
        ))}
      </ol>
    </nav>
  );
}
