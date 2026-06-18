import { ShieldCheck, ChevronRight, Settings2, UserPlus, Key, RefreshCcw, Activity } from "lucide-react";
import { Button } from "./ui/button";
import { useEffect, useState } from "react";
import { fetchAPI } from "@/lib/api";

interface AuditEvent {
  id: string;
  actor_email: string;
  action: string;
  entity_type: string;
  created_at: string;
}

export function SettingsInsightsSidebar() {
  const [auditLogs, setAuditLogs] = useState<AuditEvent[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const fetchAuditLogs = async () => {
      try {
        const data = await fetchAPI("/api/v1/audit");
        if (data) setAuditLogs(data as AuditEvent[]);
      } catch (err) {
        console.error("Failed to load audit logs:", err);
      } finally {
        setLoading(false);
      }
    };
    fetchAuditLogs();
  }, []);
  return (
    <div className="flex flex-col h-full w-full">
      
      <div className="flex-1 overflow-y-auto pr-2 custom-scrollbar flex flex-col gap-8 pb-10">
        
        {/* Your Plan */}
        <div>
          <div className="flex items-center justify-between mb-4">
            <h4 className="font-semibold text-[13px] text-zinc-100">Your Plan</h4>
            <button className="text-[10px] text-zinc-400 hover:text-white transition-colors bg-white/5 px-2 py-1 rounded-md">Manage Plan</button>
          </div>
          
          <div className="flex items-center gap-4 mb-5">
            <div className="w-10 h-10 rounded-xl bg-white/10 text-white flex items-center justify-center flex-shrink-0 border border-white/10">
              <span className="font-bold text-lg">E</span>
            </div>
            <div className="flex flex-col">
              <div className="flex items-center gap-2">
                <span className="font-bold text-sm text-white">Enterprise Plan</span>
                <span className="px-1.5 py-0.5 bg-green-500/20 text-green-400 text-[8px] font-bold rounded uppercase tracking-wider">Active</span>
              </div>
              <span className="text-[10px] text-zinc-400 mt-0.5">Renews on May 24, 2025</span>
            </div>
          </div>

          <div className="flex flex-col gap-3">
            <div className="flex flex-col gap-1.5">
              <div className="flex items-center justify-between text-[11px]">
                <span className="text-zinc-400 flex items-center gap-2">AI Teams</span>
                <span className="text-zinc-200 font-medium">8 / 10</span>
              </div>
              <div className="w-full h-1 bg-zinc-800 rounded-full overflow-hidden">
                <div className="h-full bg-white rounded-full" style={{ width: '80%' }}></div>
              </div>
            </div>

            <div className="flex flex-col gap-1.5">
              <div className="flex items-center justify-between text-[11px]">
                <span className="text-zinc-400 flex items-center gap-2">Team Members</span>
                <span className="text-zinc-200 font-medium">24 / 50</span>
              </div>
              <div className="w-full h-1 bg-zinc-800 rounded-full overflow-hidden">
                <div className="h-full bg-white rounded-full" style={{ width: '48%' }}></div>
              </div>
            </div>

            <div className="flex flex-col gap-1.5">
              <div className="flex items-center justify-between text-[11px]">
                <span className="text-zinc-400 flex items-center gap-2">Connected Channels</span>
                <span className="text-zinc-200 font-medium">8 / 20</span>
              </div>
              <div className="w-full h-1 bg-zinc-800 rounded-full overflow-hidden">
                <div className="h-full bg-white rounded-full" style={{ width: '40%' }}></div>
              </div>
            </div>

            <div className="flex flex-col gap-1.5 mt-2">
              <div className="flex items-center justify-between text-[11px]">
                <span className="text-zinc-400 flex items-center gap-2">API Calls</span>
                <span className="text-zinc-200 font-medium">245K / 1M</span>
              </div>
            </div>

            <div className="flex flex-col gap-1.5">
              <div className="flex items-center justify-between text-[11px]">
                <span className="text-zinc-400 flex items-center gap-2">Data Storage</span>
                <span className="text-zinc-200 font-medium">245 GB / 1 TB</span>
              </div>
            </div>
          </div>
        </div>

        {/* Security Status */}
        <div>
          <h4 className="font-semibold text-[13px] text-zinc-100 mb-3">Security Status</h4>
          <div className="bg-white/[0.02] border border-white/5 rounded-xl p-4 flex flex-col items-center justify-center text-center gap-3">
            <div className="w-10 h-10 rounded-full bg-green-500/20 text-green-400 flex items-center justify-center">
              <ShieldCheck className="w-5 h-5" />
            </div>
            <div>
              <h5 className="text-[13px] font-semibold text-white">Your workspace is secure</h5>
              <p className="text-[10px] text-zinc-400 mt-1">All security checks passed</p>
            </div>
            <button className="text-[10px] font-medium text-white bg-white/10 hover:bg-white/20 transition-colors px-4 py-1.5 rounded-lg w-full mt-2">
              View Security Settings
            </button>
          </div>
        </div>

        {/* Recent Activity */}
        <div>
          <h4 className="font-semibold text-[13px] text-zinc-100 mb-4">Recent Activity</h4>
          <div className="flex flex-col gap-4">
            
            {loading ? (
              <div className="flex justify-center py-4">
                <div className="w-4 h-4 border-2 border-white/20 border-t-white rounded-full animate-spin"></div>
              </div>
            ) : auditLogs.length === 0 ? (
              <div className="text-[11px] text-zinc-500 text-center py-2">No recent activity.</div>
            ) : (
              auditLogs.map((log) => (
                <div key={log.id} className="flex items-start justify-between gap-3 group cursor-pointer">
                  {log.entity_type === 'SECURITY' ? <ShieldCheck className="w-3.5 h-3.5 text-zinc-500 mt-0.5 flex-shrink-0" /> : <Activity className="w-3.5 h-3.5 text-zinc-500 mt-0.5 flex-shrink-0" />}
                  <div className="flex flex-col flex-1">
                    <span className="text-[11px] text-zinc-300 group-hover:text-white transition-colors capitalize">
                      {log.action.replace(/_/g, ' ')}
                    </span>
                    <span className="text-[9px] text-zinc-500">{log.actor_email}</span>
                  </div>
                  <span className="text-[9px] text-zinc-500 whitespace-nowrap">
                    {new Date(log.created_at).toLocaleDateString()}
                  </span>
                </div>
              ))
            )}

          </div>
          <button className="text-[10px] font-medium text-zinc-400 hover:text-white transition-colors flex items-center justify-center w-full gap-1 mt-5 bg-white/5 py-2 rounded-xl">
            View all activity <ChevronRight className="w-3 h-3" />
          </button>
        </div>

      </div>
    </div>
  );
}
