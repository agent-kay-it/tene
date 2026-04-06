// Design Ref: §3.3 — Pricing tiers for Free vs Cloud
export type PricingTier = {
  name: string;
  price: string;
  period: string;
  description: string;
  features: string[];
  cta: { label: string; action: "install" | "waitlist" };
  highlighted: boolean;
};

export const pricingTiers: PricingTier[] = [
  {
    name: "Free",
    price: "$0",
    period: "forever",
    description: "Local encrypted secrets for individual projects.",
    features: [
      "Unlimited secrets",
      "XChaCha20-Poly1305 encryption",
      "AI runtime injection",
      "OS keychain integration",
      "12-word recovery key",
    ],
    cta: { label: "Install now", action: "install" },
    highlighted: false,
  },
  {
    name: "Cloud",
    price: "$1",
    period: "per user / month",
    description: "Sync secrets across projects and machines.",
    features: [
      "Everything in Free",
      "Cross-project sync",
      "Cross-machine access",
      "Team sharing",
      "No repeated tene init",
    ],
    cta: { label: "Join waitlist", action: "waitlist" },
    highlighted: true,
  },
];
