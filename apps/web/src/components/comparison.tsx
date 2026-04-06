const rows = [
  { feature: "Local-first", tene: true, env: true, doppler: false, vault: false, infisical: false },
  { feature: "No server required", tene: true, env: true, doppler: false, vault: false, infisical: false },
  { feature: "No signup required", tene: true, env: true, doppler: false, vault: false, infisical: false },
  { feature: "Encrypted at rest", tene: true, env: false, doppler: true, vault: true, infisical: true },
  { feature: "AI agent auto-detect", tene: true, env: false, doppler: false, vault: false, infisical: false },
  { feature: "Environment injection", tene: true, env: false, doppler: true, vault: true, infisical: true },
  { feature: "100% offline", tene: true, env: true, doppler: false, vault: false, infisical: false },
  { feature: "Open source", tene: true, env: true, doppler: false, vault: false, infisical: true },
  { feature: "Free", tene: true, env: true, doppler: false, vault: false, infisical: false },
];

function Check() {
  return (
    <svg className="h-5 w-5 text-accent" fill="none" stroke="currentColor" viewBox="0 0 24 24" strokeWidth="2.5">
      <path d="M5 13l4 4L19 7" />
    </svg>
  );
}

function Cross() {
  return (
    <svg className="h-4 w-4 text-muted/40" fill="none" stroke="currentColor" viewBox="0 0 24 24" strokeWidth="2">
      <path d="M18 6L6 18M6 6l12 12" />
    </svg>
  );
}

export function Comparison() {
  return (
    <section className="px-6 py-24">
      <div className="mx-auto max-w-4xl">
        <h2 className="text-center text-3xl font-bold sm:text-4xl">
          How Tene compares
        </h2>
        <p className="mx-auto mt-4 max-w-xl text-center text-muted">
          Tene is the only tool that combines local-first encryption with AI agent auto-detection.
        </p>

        <div className="mt-12 overflow-x-auto">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-border">
                <th className="pb-4 pr-4 text-left font-normal text-muted" />
                <th className="pb-4 px-4 text-center font-bold text-accent">Tene</th>
                <th className="pb-4 px-4 text-center font-normal text-muted">.env</th>
                <th className="pb-4 px-4 text-center font-normal text-muted">Doppler</th>
                <th className="pb-4 px-4 text-center font-normal text-muted">Vault</th>
                <th className="pb-4 px-4 text-center font-normal text-muted">Infisical</th>
              </tr>
            </thead>
            <tbody>
              {rows.map((row) => (
                <tr key={row.feature} className="border-b border-border/50">
                  <td className="py-3 pr-4 text-foreground">{row.feature}</td>
                  <td className="py-3 px-4 text-center">
                    <div className="flex justify-center">{row.tene ? <Check /> : <Cross />}</div>
                  </td>
                  <td className="py-3 px-4 text-center">
                    <div className="flex justify-center">{row.env ? <Check /> : <Cross />}</div>
                  </td>
                  <td className="py-3 px-4 text-center">
                    <div className="flex justify-center">{row.doppler ? <Check /> : <Cross />}</div>
                  </td>
                  <td className="py-3 px-4 text-center">
                    <div className="flex justify-center">{row.vault ? <Check /> : <Cross />}</div>
                  </td>
                  <td className="py-3 px-4 text-center">
                    <div className="flex justify-center">{row.infisical ? <Check /> : <Cross />}</div>
                  </td>
                </tr>
              ))}
              <tr>
                <td className="pt-4 pr-4 text-muted">Price</td>
                <td className="pt-4 px-4 text-center font-bold text-accent">$0</td>
                <td className="pt-4 px-4 text-center text-muted">$0</td>
                <td className="pt-4 px-4 text-center text-muted">$21/user/mo</td>
                <td className="pt-4 px-4 text-center text-muted">$1,152+/mo</td>
                <td className="pt-4 px-4 text-center text-muted">$6/user/mo</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </section>
  );
}
