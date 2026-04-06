// Design Ref: §3.5 — Comparison data with "Secrets hidden from AI" as first row
export type ComparisonRow = {
  feature: string;
  tene: boolean;
  env: boolean;
  doppler: boolean;
  vault: boolean;
  infisical: boolean;
};

export const comparisonRows: ComparisonRow[] = [
  { feature: "Secrets hidden from AI", tene: true, env: false, doppler: false, vault: false, infisical: false },
  { feature: "Local-first", tene: true, env: true, doppler: false, vault: false, infisical: false },
  { feature: "No server required", tene: true, env: true, doppler: false, vault: false, infisical: false },
  { feature: "Encrypted at rest", tene: true, env: false, doppler: true, vault: true, infisical: true },
  { feature: "AI agent auto-detect", tene: true, env: false, doppler: false, vault: false, infisical: false },
  { feature: "Runtime injection", tene: true, env: false, doppler: true, vault: true, infisical: true },
  { feature: "No signup required", tene: true, env: true, doppler: false, vault: false, infisical: false },
  { feature: "Open source", tene: true, env: true, doppler: false, vault: false, infisical: true },
];

export const comparisonPricing = {
  tene: "$0",
  env: "$0",
  doppler: "$21/mo",
  vault: "$1,152+",
  infisical: "$6/mo",
};
