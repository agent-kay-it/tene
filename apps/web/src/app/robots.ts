import type { MetadataRoute } from "next";

// Explicit allow-list for LLM crawlers so conservative bots that default to
// deny-on-missing-rule don't skip us. Also explicitly disallows /.tene/
// (the encrypted vault directory that must never be served) and the
// Bytespider crawler (high-volume, low-signal).
//
// Coverage:
//   - OpenAI: GPTBot, ChatGPT-User, OAI-SearchBot
//   - Anthropic: ClaudeBot, Claude-Web, anthropic-ai
//   - Google: Google-Extended (training), Googlebot (search)
//   - Others: PerplexityBot, CCBot, Applebot-Extended, Meta-ExternalAgent,
//             Amazonbot, YouBot, cohere-ai
export default function robots(): MetadataRoute.Robots {
  const base = "https://tene.sh";
  return {
    rules: [
      {
        userAgent: ["GPTBot", "ChatGPT-User", "OAI-SearchBot"],
        allow: "/",
        disallow: ["/.tene/", "/api/"],
      },
      {
        userAgent: ["ClaudeBot", "Claude-Web", "anthropic-ai"],
        allow: "/",
        disallow: ["/.tene/", "/api/"],
      },
      {
        userAgent: ["Google-Extended", "Googlebot"],
        allow: "/",
      },
      {
        userAgent: [
          "PerplexityBot",
          "CCBot",
          "Applebot-Extended",
          "Meta-ExternalAgent",
          "Amazonbot",
          "YouBot",
          "cohere-ai",
        ],
        allow: "/",
      },
      { userAgent: "Bytespider", disallow: "/" },
      { userAgent: "*", allow: "/" },
    ],
    sitemap: [`${base}/sitemap.xml`, `${base}/blog/rss.xml`],
    host: base,
  };
}
