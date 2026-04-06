"use client";

import { useState } from "react";
import { faqs } from "@/data/faq";

// Design Ref: §4.7 — FAQ with data import, .env risk questions first
export function FAQ() {
  const [openIndex, setOpenIndex] = useState<number | null>(null);

  return (
    <section id="faq" className="px-4 py-24 sm:px-6">
      <div className="mx-auto max-w-2xl">
        <h2 className="text-center text-3xl font-bold sm:text-4xl">FAQ</h2>

        <div className="mt-12 divide-y divide-border">
          {faqs.map((faq, i) => (
            <div key={i}>
              <button
                onClick={() => setOpenIndex(openIndex === i ? null : i)}
                className="flex w-full items-center justify-between py-5 text-left"
              >
                <span className="text-sm font-medium sm:text-base">
                  {faq.question}
                </span>
                <svg
                  className={`h-4 w-4 shrink-0 text-muted transition-transform ${
                    openIndex === i ? "rotate-180" : ""
                  }`}
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                  strokeWidth="2"
                >
                  <path d="M19 9l-7 7-7-7" />
                </svg>
              </button>
              {openIndex === i && (
                <p className="pb-5 text-sm leading-relaxed text-muted">
                  {faq.answer}
                </p>
              )}
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}
