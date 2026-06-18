"use client";

import { useEffect, useState } from "react";
import { SettingsLeftSidebar } from "@/components/SettingsLeftSidebar";
import { SettingsInsightsSidebar } from "@/components/SettingsInsightsSidebar";
import UnifiedDirectory from "@unified-api/react-directory";
import { CheckCircle2, XCircle, Loader2 } from "lucide-react";
import { fetchAPI } from "@/lib/api";

type CommerceConnection = {
  id: string;
  unified_connection_id: string;
  provider: string;
  status: string;
  created_at: string;
  updated_at: string;
};

export default function IntegrationsPage() {
  const [connections, setConnections] = useState<CommerceConnection[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const workspaceId = process.env.NEXT_PUBLIC_UNIFIED_WORKSPACE_ID;

  useEffect(() => {
    let mounted = true;
    fetchAPI("/api/v1/integrations/connections")
      .then((data) => {
        if (!mounted) return;
        const payload = data as { connections?: CommerceConnection[] };
        setConnections(payload.connections ?? []);
        setError(null);
      })
      .catch((err: Error) => {
        if (!mounted) return;
        setError(err.message || "Failed to load connections.");
      })
      .finally(() => {
        if (mounted) setLoading(false);
      });
    return () => {
      mounted = false;
    };
  }, []);

  return (
    <>
      {/* Left Sidebar */}
      <SettingsLeftSidebar />

      {/* Main Content Area */}
      <main className="flex-1 overflow-y-auto custom-scrollbar pr-4 flex flex-col gap-8 pb-10">
        
        {/* Header */}
        <div className="mt-2">
          <h1 className="text-3xl font-bold tracking-tight">Integrations</h1>
          <p className="text-muted-foreground mt-1">Manage your connected third-party platforms and apps.</p>
        </div>

        {/* Active Connections Container */}
        <div className="bg-white/20 dark:bg-white/5 rounded-[2rem] p-6 border border-white/30 dark:border-white/10 flex flex-col gap-6">
          <div>
            <h3 className="font-semibold text-[15px]">Active Connections</h3>
            <p className="text-[11px] text-muted-foreground mt-0.5">Platforms currently connected to your workspace.</p>
          </div>
          
          {loading ? (
            <div className="flex items-center justify-center py-8">
              <Loader2 className="w-6 h-6 animate-spin text-muted-foreground" />
            </div>
          ) : error ? (
            <div className="text-center py-8 border border-red-500/20 rounded-2xl bg-red-500/10">
              <p className="text-sm text-red-600 dark:text-red-400">{error}</p>
            </div>
          ) : connections.length === 0 ? (
            <div className="text-center py-8 border border-black/5 dark:border-white/10 rounded-2xl bg-white/40 dark:bg-black/20">
              <p className="text-sm text-muted-foreground">No active connections yet.</p>
              <p className="text-xs text-muted-foreground mt-1">Browse the directory below to connect an app.</p>
            </div>
          ) : (
            <div className="grid grid-cols-2 gap-4">
              {connections.map((conn) => (
                <div key={conn.id} className="bg-white/40 dark:bg-black/20 rounded-2xl p-4 border border-black/5 dark:border-white/10 flex items-center justify-between">
                  <div className="flex items-center gap-4">
                    <div className="w-10 h-10 rounded-xl bg-white flex items-center justify-center shadow-sm">
                      <span className="font-bold text-xs capitalize">{conn.provider.charAt(0) || "C"}</span>
                    </div>
                    <div>
                      <h4 className="font-semibold text-sm capitalize">{conn.provider}</h4>
                      <p className="text-[10px] text-muted-foreground font-mono mt-0.5">ID: {conn.unified_connection_id.slice(0, 12)}...</p>
                    </div>
                  </div>
                  <div className="flex items-center gap-2">
                    {conn.status === "ACTIVE" ? (
                      <div className="flex items-center gap-1.5 px-2 py-1 rounded-full bg-green-500/10 text-green-600 dark:text-green-400">
                        <CheckCircle2 className="w-3 h-3" />
                        <span className="text-[10px] font-semibold uppercase tracking-wider">Healthy</span>
                      </div>
                    ) : (
                      <div className="flex items-center gap-1.5 px-2 py-1 rounded-full bg-red-500/10 text-red-600 dark:text-red-400">
                        <XCircle className="w-3 h-3" />
                        <span className="text-[10px] font-semibold uppercase tracking-wider">Error</span>
                      </div>
                    )}
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>

        {/* Unified Directory Embedded Component */}
        <div className="bg-white/20 dark:bg-white/5 rounded-[2rem] p-6 border border-white/30 dark:border-white/10">
          <div className="mb-6">
            <h3 className="font-semibold text-[15px]">App Directory</h3>
            <p className="text-[11px] text-muted-foreground mt-0.5">Discover and connect supported platforms.</p>
          </div>
          
          <div className="w-full rounded-2xl overflow-hidden border border-black/5 dark:border-white/10 shadow-sm mix-blend-luminosity">
            {workspaceId ? (
              <UnifiedDirectory 
                workspaceId={workspaceId}
                categories={["commerce", "accounting"]}
              />
            ) : (
              <div className="p-6 text-sm text-muted-foreground bg-white/40 dark:bg-black/20">
                Unified.to workspace is not configured for this dashboard environment.
              </div>
            )}
          </div>
        </div>

      </main>

      {/* Right Matte Black Area */}
      <aside className="w-[360px] flex-shrink-0 bg-[#121212] text-white p-6 rounded-3xl relative z-20 flex flex-col h-full shadow-[0_8px_30px_rgb(0,0,0,0.4)] border border-white/5">
        <SettingsInsightsSidebar />
      </aside>
    </>
  );
}
