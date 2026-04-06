"use client";

import { useState, useEffect, useRef, useCallback } from "react";

// Design Ref: §4.5 — Terminal demo showing .env risk then tene solution
type Line =
  | { type: "command"; text: string }
  | { type: "output"; text: string; style?: "green" | "accent" | "dim" | "red" }
  | { type: "blank" };

const SCRIPT: Line[] = [
  { type: "command", text: "cat .env" },
  { type: "output", text: "  STRIPE_KEY=sk_test_51Hxxxxx", style: "red" },
  { type: "output", text: "  OPENAI_API_KEY=sk-proj-xxxxx", style: "red" },
  { type: "output", text: "  \u2191 AI can see this", style: "dim" },
  { type: "blank" },
  { type: "command", text: "tene import .env" },
  { type: "output", text: "  \u2713 2 secrets imported and encrypted", style: "green" },
  { type: "output", text: "  \u2713 .env can now be deleted", style: "green" },
  { type: "blank" },
  { type: "command", text: "tene run -- claude" },
  { type: "output", text: "  \u2713 2 secrets injected as environment variables", style: "green" },
  { type: "output", text: "  \u2713 Starting: claude", style: "green" },
  { type: "output", text: "  \u2713 AI cannot see secret values", style: "accent" },
];

const TYPING_SPEED = 40;
const LINE_PAUSE = 400;
const RESTART_DELAY = 4000;

export function Terminal() {
  const [lines, setLines] = useState<Line[]>([]);
  const [currentLine, setCurrentLine] = useState(0);
  const [currentChar, setCurrentChar] = useState(0);
  const [isTyping, setIsTyping] = useState(true);
  const containerRef = useRef<HTMLDivElement>(null);

  const scrollToBottom = useCallback(() => {
    const el = containerRef.current;
    if (el) el.scrollTop = el.scrollHeight;
  }, []);

  useEffect(() => {
    if (currentLine >= SCRIPT.length) {
      const timeout = setTimeout(() => {
        setLines([]);
        setCurrentLine(0);
        setCurrentChar(0);
        setIsTyping(true);
      }, RESTART_DELAY);
      return () => clearTimeout(timeout);
    }

    const line = SCRIPT[currentLine];

    if (line.type === "blank") {
      const timeout = setTimeout(() => {
        setLines((prev) => [...prev, line]);
        setCurrentLine((prev) => prev + 1);
        scrollToBottom();
      }, LINE_PAUSE / 2);
      return () => clearTimeout(timeout);
    }

    if (line.type === "output") {
      const timeout = setTimeout(() => {
        setLines((prev) => [...prev, line]);
        setCurrentLine((prev) => prev + 1);
        setIsTyping(true);
        scrollToBottom();
      }, LINE_PAUSE);
      return () => clearTimeout(timeout);
    }

    if (line.type === "command") {
      if (currentChar < line.text.length) {
        const timeout = setTimeout(() => {
          setCurrentChar((prev) => prev + 1);
          scrollToBottom();
        }, TYPING_SPEED);
        return () => clearTimeout(timeout);
      } else {
        const timeout = setTimeout(() => {
          setLines((prev) => [...prev, line]);
          setCurrentLine((prev) => prev + 1);
          setCurrentChar(0);
          setIsTyping(true);
          scrollToBottom();
        }, LINE_PAUSE);
        return () => clearTimeout(timeout);
      }
    }
  }, [currentLine, currentChar, scrollToBottom]);

  const currentScript = currentLine < SCRIPT.length ? SCRIPT[currentLine] : null;

  return (
    <section className="px-4 py-12 sm:px-6">
      <div className="mx-auto max-w-2xl">
        <div className="overflow-hidden rounded-xl border border-border bg-surface shadow-2xl shadow-black/40">
          <div className="flex items-center gap-2 border-b border-border px-4 py-3">
            <span className="h-3 w-3 rounded-full bg-[#ff5f56]" />
            <span className="h-3 w-3 rounded-full bg-[#ffbd2e]" />
            <span className="h-3 w-3 rounded-full bg-[#27c93f]" />
            <span className="ml-2 text-xs text-muted">terminal</span>
          </div>

          <div
            ref={containerRef}
            className="h-[340px] overflow-y-auto p-4 font-mono text-sm leading-6"
          >
            {lines.map((line, i) => {
              if (line.type === "blank") return <div key={i} className="h-3" />;
              if (line.type === "command")
                return (
                  <div key={i}>
                    <span className="text-accent">$ </span>
                    <span>{line.text}</span>
                  </div>
                );
              return (
                <div
                  key={i}
                  className={
                    line.style === "green"
                      ? "text-green-400"
                      : line.style === "accent"
                        ? "text-accent"
                        : line.style === "red"
                          ? "text-red-400"
                          : "text-muted"
                  }
                >
                  {line.text}
                </div>
              );
            })}

            {currentScript?.type === "command" && isTyping && (
              <div>
                <span className="text-accent">$ </span>
                <span>{currentScript.text.slice(0, currentChar)}</span>
                <span className="terminal-cursor" />
              </div>
            )}
          </div>
        </div>

        <p className="mt-4 text-center text-sm text-muted">
          From .env exposure to encrypted runtime injection &mdash; under 1
          minute.
        </p>
      </div>
    </section>
  );
}
