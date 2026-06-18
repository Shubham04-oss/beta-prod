"use client";

import { useState } from "react";
import { PIMInsightsSidebar } from "@/components/PIMInsightsSidebar";
import { PIMLeftSidebar } from "@/components/PIMLeftSidebar";
import { Search, AlertCircle, RefreshCw } from "lucide-react";
import { Button } from "@/components/ui/button";
import { useGetPIMStats } from "@/hooks/useProducts";
import useSWR from "swr";
import { fetchAPI } from "@/lib/api";

export default function ValidationIssuesPage() {
  const { data: stats } = useGetPIMStats();
  const [searchQuery, setSearchQuery] = useState("");

  const { data: issues, mutate, isValidating } = useSWR("/api/v1/pim/validation", fetchAPI);

  return (
    <>
      <PIMLeftSidebar />
      <main className="flex-1 flex flex-col h-full relative z-10 overflow-hidden pr-4">
        {/* Header */}
        <div className="flex items-center justify-between px-8 py-6 flex-shrink-0">
          <div>
            <h1 className="text-2xl font-bold text-foreground flex items-center gap-2">
              <AlertCircle className="w-6 h-6 text-red-500" />
              Validation Issues
            </h1>
            <p className="text-sm text-muted-foreground mt-1">Review and resolve data quality issues across your catalog.</p>
          </div>
          <div className="flex items-center gap-3">
            <div className="relative">
              <Search className="w-4 h-4 absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground" />
              <input 
                type="text"
                placeholder="Search issues..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="w-64 h-10 pl-9 pr-4 rounded-xl bg-black/5 dark:bg-white/5 border-none text-sm focus:outline-none focus:ring-2 focus:ring-primary/20 transition-all"
              />
            </div>
            <Button onClick={() => mutate()} disabled={isValidating} className="rounded-full shadow-sm bg-black text-white hover:bg-zinc-800 dark:bg-white dark:text-black dark:hover:bg-zinc-200 gap-2">
              <RefreshCw className={`w-4 h-4 ${isValidating ? 'animate-spin' : ''}`} />
              Run Validator
            </Button>
          </div>
        </div>

        {/* Content Area */}
        <div className="flex-1 overflow-y-auto px-8 pb-8 custom-scrollbar">
          <div className="bg-white dark:bg-[#121212] rounded-3xl border border-black/5 dark:border-white/5 shadow-sm overflow-hidden min-h-[400px]">
            {(!issues || (issues as any[]).length === 0) ? (
              <div className="h-[400px] flex flex-col items-center justify-center text-center">
                <div className="w-16 h-16 rounded-full bg-green-500/10 flex items-center justify-center mb-4">
                  <AlertCircle className="w-8 h-8 text-green-500" />
                </div>
                <h3 className="text-lg font-semibold text-foreground">All clear!</h3>
                <p className="text-sm text-muted-foreground max-w-sm mt-2 mb-6">Your data validator found zero issues. Your catalog data quality is looking great.</p>
                <Button onClick={() => mutate()} className="rounded-full bg-black text-white hover:bg-zinc-800 dark:bg-white dark:text-black dark:hover:bg-zinc-200">
                  <RefreshCw className={`w-4 h-4 mr-2 ${isValidating ? 'animate-spin' : ''}`} />
                  Scan Again
                </Button>
              </div>
            ) : (
              <table className="w-full text-left border-collapse">
                <thead>
                  <tr className="border-b border-black/5 dark:border-white/5">
                    <th className="py-4 px-6 text-xs font-semibold text-muted-foreground uppercase tracking-wider">Severity</th>
                    <th className="py-4 px-6 text-xs font-semibold text-muted-foreground uppercase tracking-wider">Issue Type</th>
                    <th className="py-4 px-6 text-xs font-semibold text-muted-foreground uppercase tracking-wider">Message</th>
                    <th className="py-4 px-6 text-xs font-semibold text-muted-foreground uppercase tracking-wider">Created</th>
                    <th className="py-4 px-6 text-xs font-semibold text-muted-foreground uppercase tracking-wider text-right">Actions</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-black/5 dark:divide-white/5">
                  {(issues as any[]).map((issue: any) => (
                    <tr key={issue.id} className="hover:bg-black/5 dark:hover:bg-white/5 transition-colors group">
                      <td className="py-4 px-6">
                        <span className={`text-[10px] font-bold px-2 py-1 rounded-md uppercase tracking-wider ${
                          issue.severity === 'ERROR' ? 'bg-red-500/10 text-red-500' :
                          issue.severity === 'WARNING' ? 'bg-amber-500/10 text-amber-500' :
                          'bg-blue-500/10 text-blue-500'
                        }`}>
                          {issue.severity}
                        </span>
                      </td>
                      <td className="py-4 px-6">
                        <span className="font-medium text-sm text-foreground">{issue.issue_type}</span>
                      </td>
                      <td className="py-4 px-6">
                        <span className="text-sm text-muted-foreground">{issue.message}</span>
                      </td>
                      <td className="py-4 px-6">
                        <span className="text-sm text-muted-foreground">{new Date(issue.created_at).toLocaleDateString()}</span>
                      </td>
                      <td className="py-4 px-6 text-right">
                        <Button size="sm" variant="ghost" className="h-8 text-xs rounded-lg opacity-0 group-hover:opacity-100 transition-opacity text-primary hover:text-primary">Resolve</Button>
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
