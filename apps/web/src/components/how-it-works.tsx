const steps = [
  {
    step: "01",
    title: "Install",
    command: "npm install -g @tene/cli",
    description: "One command. No account creation, no API keys, no server setup.",
  },
  {
    step: "02",
    title: "Initialize",
    command: "tene init",
    description:
      "Creates an encrypted vault, generates CLAUDE.md for AI agents, and issues a 12-word recovery key.",
  },
  {
    step: "03",
    title: "Store secrets",
    command: "tene set STRIPE_KEY sk_test_xxx",
    description:
      "Secrets are encrypted with XChaCha20-Poly1305 and stored in a local SQLite vault. Never leaves your machine.",
  },
  {
    step: "04",
    title: "Develop with secrets",
    command: "tene run -- cursor .",
    description:
      "Injects all secrets as environment variables into any command. Your AI agent reads CLAUDE.md and knows the rest.",
  },
];

export function HowItWorks() {
  return (
    <section id="how-it-works" className="px-4 py-24 sm:px-6">
      <div className="mx-auto max-w-3xl">
        <h2 className="text-center text-3xl font-bold sm:text-4xl">
          Up and running in{" "}
          <span className="text-accent">3 minutes</span>
        </h2>

        <div className="mt-16 space-y-12">
          {steps.map((s, i) => (
            <div key={s.step} className="relative flex gap-6">
              <div className="flex flex-col items-center">
                <div className="flex h-10 w-10 shrink-0 items-center justify-center rounded-full border border-accent/40 bg-accent/10 text-sm font-bold text-accent">
                  {s.step}
                </div>
                {i < steps.length - 1 && (
                  <div className="mt-2 h-full w-px bg-border" />
                )}
              </div>
              <div className="pb-12">
                <h3 className="text-xl font-semibold">{s.title}</h3>
                <div className="mt-3 inline-flex items-center gap-2 rounded-lg border border-border bg-surface px-4 py-2 font-mono text-sm">
                  <span className="text-accent">$</span>
                  <code>{s.command}</code>
                </div>
                <p className="mt-3 text-sm leading-relaxed text-muted">
                  {s.description}
                </p>
              </div>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}
