"use client";

import { useState } from "react";
import { PIMInsightsSidebar } from "@/components/PIMInsightsSidebar";
import { PIMLeftSidebar } from "@/components/PIMLeftSidebar";
import { Search, History, Filter } from "lucide-react";
import { Button } from "@/components/ui/button";
import { useGetPIMStats } from "@/hooks/useProducts";
import useSWR from "swr";
import { fetchAPI } from "@/lib/api";

export default function AuditLogPage() {
  const { data: stats } = useGetPIMStats();
  const [searchQuery, setSearchQuery] = useState("");

  const { data: logs, mutate, isValidating } = useSWR("/api/v1/pim/audit", fetchAPI);

  return (
    <>
      <PIMLeftSidebar />
      <main className="flex-1 flex flex-col h-full relative z-10 overflow-hidden pr-4">
        {/* Header */}
        <div className="flex items-center justify-between px-8 py-6 flex-shrink-0">
          <div>
            <h1 className="text-2xl font-bold text-foreground flex items-center gap-2">
              <History className="w-6 h-6 text-primary" />
              Data Audit Log
            </h1>
            <p className="text-sm text-muted-foreground mt-1">Track every change across your product catalog.</p>
          </div>
          <div className="flex items-center gap-3">
            <div className="relative">
              <Search className="w-4 h-4 absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground" />
              <input 
                type="text"
                placeholder="Search logs..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="w-64 h-10 pl-9 pr-4 rounded-xl bg-black/5 dark:bg-white/5 border-none text-sm focus:outline-none focus:ring-2 focus:ring-primary/20 transition-all"
              />
            </div>
            <Button variant="outline" className="rounded-full h-10 px-4 bg-transparent border border-black/10 dark:border-white/10 hover:bg-black/5 dark:hover:bg-white/5 gap-2">
              <Filter className="w-4 h-4" />
              Filters
            </Button>
          </div>
        </div>

        {/* Content Area */}
        <div className="flex-1 overflow-y-auto px-8 pb-8 custom-scrollbar">
          <div className="bg-white dark:bg-[#121212] rounded-3xl border border-black/5 dark:border-white/5 shadow-sm overflow-hidden min-h-[400px]">
            {(!logs || (logs as any[]).length === 0) ? (
              <div className="h-[400px] flex flex-col items-center justify-center text-center">
                <div className="w-16 h-16 rounded-full bg-black/5 dark:bg-white/5 flex items-center justify-center mb-4">
                  <History className="w-8 h-8 text-muted-foreground" />
                </div>
                <h3 className="text-lg font-semibold text-foreground">No logs recorded</h3>
                <p className="text-sm text-muted-foreground max-w-sm mt-2 mb-6">Start managing your catalog to see activity logs here.</p>
              </div>
            ) : (
              <table className="w-full text-left border-collapse">
                <thead>
                  <tr className="border-b border-black/5 dark:border-white/5">
                    <th className="py-4 px-6 text-xs font-semibold text-muted-foreground uppercase tracking-wider">Timestamp</th>
                    <th className="py-4 px-6 text-xs font-semibold text-muted-foreground uppercase tracking-wider">Actor</th>
                    <th className="py-4 px-6 text-xs font-semibold text-muted-foreground uppercase tracking-wider">Action</th>
                    <th className="py-4 px-6 text-xs font-semibold text-muted-foreground uppercase tracking-wider">Entity</th>
                    <th className="py-4 px-6 text-xs font-semibold text-muted-foreground uppercase tracking-wider">Details</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-black/5 dark:divide-white/5">
                  {(logs as any[]).map((log: any) => (
                    <tr key={log.id} className="hover:bg-black/5 dark:hover:bg-white/5 transition-colors group">
                      <td className="py-4 px-6">
                        <span className="text-sm font-medium text-foreground">{new Date(log.created_at).toLocaleString()}</span>
                      </td>
                      <td className="py-4 px-6">
                        <span className="text-sm text-muted-foreground">{log.actor_email || 'System'}</span>
                      </td>
                      <td className="py-4 px-6">
                        <span className="text-[10px] font-bold px-2 py-1 rounded-md bg-black/5 dark:bg-white/5 uppercase tracking-wider text-foreground">
                          {log.action}
                        </span>
                      </td>
                      <td className="py-4 px-6">
                        <span className="text-sm text-muted-foreground">{log.entity_type}</span>
                      </td>
                      <td className="py-4 px-6">
                        <span className="text-xs font-mono text-muted-foreground bg-black/5 dark:bg-white/5 px-2 py-1 rounded-lg line-clamp-1 max-w-[200px]">
                          {JSON.stringify(log.details)}
                        </span>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            )}
          </div>
        </div>
      </main>

      <aside className="w-[360px] flex-shrink-0 bg-[#121212] text-white p-6 rounded-3xl relative z-20 flex flex-col h-full shadow-[0_8px_30px_rgb(0,0,0,0.4)] border border-white/5">
        <PIMInsightsSidebar stats={stats} />
      </aside>
    </>
  );
}
