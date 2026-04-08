"use client";

import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { api } from "@/lib/api";

const FILTERS = ["All", "push", "pull", "login", "delete", "create"] as const;

export default function AuditPage() {
  const [activeFilter, setActiveFilter] = useState<string>("All");

  const { data: logs, isLoading } = useQuery({
    queryKey: ["audit-logs", activeFilter],
    queryFn: () =>
      api.listAuditLogs({
        action: activeFilter === "All" ? undefined : activeFilter,
        limit: 50,
      }),
  });

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold">Audit Log</h1>
        <p className="text-muted text-sm mt-1">Security activity across your account</p>
      </div>

      <div className="flex gap-2 flex-wrap">
        {FILTERS.map((filter) => (
          <button
            key={filter}
            onClick={() => setActiveFilter(filter)}
            className={`px-3 py-1.5 rounded-lg text-xs border transition-colors ${
              filter === activeFilter
                ? "border-accent/30 bg-accent/10 text-accent"
                : "border-border bg-surface text-muted hover:text-foreground"
            }`}
          >
            {filter === "All" ? "All" : filter.charAt(0).toUpperCase() + filter.slice(1)}
          </button>
        ))}
      </div>

      <div className="rounded-xl border border-border bg-surface overflow-hidden">
        <table className="w-full text-sm" role="table" aria-label="Audit log">
          <thead>
            <tr className="border-b border-border text-left text-xs text-muted uppercase tracking-wider">
              <th className="px-4 py-3" scope="col">Time</th>
              <th className="px-4 py-3" scope="col">Action</th>
              <th className="px-4 py-3 hidden sm:table-cell" scope="col">Detail</th>
              <th className="px-4 py-3 hidden md:table-cell" scope="col">IP</th>
            </tr>
          </thead>
          <tbody>
            {isLoading ? (
              <tr>
                <td colSpan={4} className="px-4 py-12 text-center text-muted animate-pulse">
                  Loading audit logs...
                </td>
              </tr>
            ) : !logs || logs.length === 0 ? (
              <tr>
                <td colSpan={4} className="px-4 py-12 text-center text-muted">
                  <span className="font-mono text-2xl block mb-2">&#x2630;</span>
                  No audit events{activeFilter !== "All" ? ` for "${activeFilter}"` : ""} yet
                </td>
              </tr>
            ) : (
              logs.map((log) => (
                <tr key={log.id} className="border-b border-border last:border-0 hover:bg-surface-2 transition-colors">
                  <td className="px-4 py-3 text-xs text-muted whitespace-nowrap">
                    {new Date(log.created_at).toLocaleString()}
                  </td>
                  <td className="px-4 py-3">
                    <span className={`badge ${actionBadgeClass(log.action)}`}>
                      {log.action}
                    </span>
                  </td>
                  <td className="px-4 py-3 hidden sm:table-cell text-xs text-muted truncate max-w-[200px]">
                    {log.detail || "-"}
                  </td>
                  <td className="px-4 py-3 hidden md:table-cell font-mono text-xs text-muted">
                    {log.ip_address || "-"}
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
}

function actionBadgeClass(action: string): string {
  switch (action) {
    case "push":
      return "badge-active";
    case "pull":
      return "badge-free";
    case "delete":
      return "badge-danger";
    case "login":
      return "badge-pro";
    default:
      return "badge-free";
  }
}
