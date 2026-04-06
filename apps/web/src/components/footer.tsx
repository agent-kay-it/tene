export function Footer() {
  return (
    <footer className="border-t border-border px-6 py-8">
      <div className="mx-auto flex max-w-5xl flex-col items-center justify-between gap-4 sm:flex-row">
        <div className="flex items-center gap-2 font-mono text-sm text-muted">
          <span className="text-accent">$</span>
          <span>tene</span>
          <span className="text-border">|</span>
          <span>Agentic Secret Runtime</span>
        </div>
        <div className="flex items-center gap-6 text-sm text-muted">
          <a
            href="https://github.com/tomo-kay/tene"
            target="_blank"
            rel="noopener noreferrer"
            className="transition-colors hover:text-foreground"
          >
            GitHub
          </a>
          <a
            href="https://github.com/tomo-kay/tene/issues"
            target="_blank"
            rel="noopener noreferrer"
            className="transition-colors hover:text-foreground"
          >
            Issues
          </a>
          <span className="text-border">MIT License</span>
        </div>
      </div>
    </footer>
  );
}
