"use client";

import { useEffect, useState } from "react";
import { SettingsLeftSidebar } from "@/components/SettingsLeftSidebar";
import { SettingsInsightsSidebar } from "@/components/SettingsInsightsSidebar";
import { Activity, Search } from "lucide-react";
import { fetchAPI } from "@/lib/api";

type AuditEvent = {
  id: string;
  actor_email?: string;
  action: string;
  entity_type: string;
  entity_id?: string;
  details?: Record<string, unknown>;
  ip_address?: string;
  created_at: string;
};

export default function SettingsAuditPage() {
  const [events, setEvents] = useState<AuditEvent[]>([]);
  const [query, setQuery] = useState("");
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let mounted = true;
    fetchAPI("/api/v1/audit")
      .then((data) => {
        if (!mounted) return;
        setEvents((data as AuditEvent[]) ?? []);
        setError(null);
      })
      .catch((err: Error) => {
        if (!mounted) return;
        setError(err.message || "Failed to load audit logs.");
      })
      .finally(() => {
        if (mounted) setLoading(false);
      });
    return () => {
      mounted = false;
    };
  }, []);

  const filtered = events.filter((event) => {
    const needle = query.trim().toLowerCase();
    return needle === "" || [event.action, event.entity_type, event.actor_email || ""].some((value) => value.toLowerCase().includes(needle));
  });

  return (
    <>
      <SettingsLeftSidebar />

      <main className="flex-1 overflow-y-auto custom-scrollbar pr-4 flex flex-col gap-8 pb-10">
        <div className="mt-2">
          <h1 className="text-3xl font-bold tracking-tight">Audit Logs</h1>
          <p className="text-muted-foreground mt-1">Tenant-scoped activity and control-plane changes.</p>
        </div>

        <section className="bg-white/20 dark:bg-white/5 rounded-[2rem] p-6 border border-white/30 dark:border-white/10 flex flex-col flex-1">
          <div className="flex items-center justify-between mb-6">
            <div>
              <h3 className="font-semibold text-[15px]">Recent Activity</h3>
              <p className="text-[11px] text-muted-foreground mt-0.5">Latest 50 audit events from Postgres.</p>
            </div>
            <div className="relative w-72">
              <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
              <input
                value={query}
                onChange={(event) => setQuery(event.target.value)}
                placeholder="Search audit logs..."
                className="w-full h-10 pl-9 pr-4 rounded-full bg-white/40 dark:bg-black/20 border border-black/5 dark:border-white/10 text-sm focus:outline-none focus:ring-2 focus:ring-primary/50"
              />
            </div>
          </div>

          <div className="overflow-x-auto">
            <table className="w-full text-sm text-left">
              <thead>
                <tr className="text-xs text-muted-foreground uppercase tracking-wider border-b border-black/5 dark:border-white/5">
                  <th className="pb-3 font-semibold px-2">Action</th>
                  <th className="pb-3 font-semibold px-2">Entity</th>
                  <th className="pb-3 font-semibold px-2">Actor</th>
                  <th className="pb-3 font-semibold px-2">IP Address</th>
                  <th className="pb-3 font-semibold px-2">Time</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-black/5 dark:divide-white/5">
                {loading ? (
                  <tr><td colSpan={5} className="py-10 text-center text-muted-foreground">Loading audit logs...</td></tr>
                ) : error ? (
                  <tr><td colSpan={5} className="py-10 text-center text-red-600 dark:text-red-400">{error}</td></tr>
                ) : filtered.length === 0 ? (
                  <tr><td colSpan={5} className="py-10 text-center text-muted-foreground">No audit events match this view.</td></tr>
                ) : (
                  filtered.map((event) => (
                    <tr key={event.id} className="hover:bg-black/[0.02] dark:hover:bg-white/[0.02] transition-colors">
                      <td className="py-3 px-2">
                        <div className="flex items-center gap-2">
                          <Activity className="w-4 h-4 text-muted-foreground" />
                          <span className="font-medium capitalize">{event.action.replaceAll("_", " ").toLowerCase()}</span>
                        </div>
                      </td>
                      <td className="py-3 px-2">{event.entity_type}</td>
                      <td className="py-3 px-2 text-muted-foreground">{event.actor_email || "system"}</td>
                      <td className="py-3 px-2 text-muted-foreground">{event.ip_address || "not captured"}</td>
                      <td className="py-3 px-2 text-muted-foreground">{event.created_at ? new Date(event.created_at).toLocaleString() : "No timestamp"}</td>
                    </tr>
                  ))
                )}
              </tbody>
            </table>
          </div>
        </section>
      </main>

      <aside className="w-[360px] flex-shrink-0 bg-[#121212] text-white p-6 rounded-3xl relative z-20 flex flex-col h-full shadow-[0_8px_30px_rgb(0,0,0,0.4)] border border-white/5">
        <SettingsInsightsSidebar />
      </aside>
    </>
  );
}
