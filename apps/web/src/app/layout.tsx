import type { Metadata } from "next";
import { Geist, Geist_Mono } from "next/font/google";
import "./globals.css";

const geistSans = Geist({
  variable: "--font-geist-sans",
  subsets: ["latin"],
});

const geistMono = Geist_Mono({
  variable: "--font-geist-mono",
  subsets: ["latin"],
});

export const metadata: Metadata = {
  title: "Tene — Secret management that AI agents understand",
  description:
    "Local-first encrypted secret management CLI. AI agents auto-detect your secrets. No server, no signup, no cost.",
  openGraph: {
    title: "Tene — Secret management that AI agents understand",
    description:
      "Local-first encrypted secret management CLI. No server, no signup, free.",
    url: "https://tene.dev",
    siteName: "Tene",
    type: "website",
  },
  twitter: {
    card: "summary_large_image",
    title: "Tene — Secret management that AI agents understand",
    description:
      "Local-first encrypted secret management CLI. No server, no signup, free.",
  },
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html
      lang="en"
      className={`${geistSans.variable} ${geistMono.variable} h-full antialiased`}
    >
      <body className="min-h-full flex flex-col">{children}</body>
    </html>
  );
}
