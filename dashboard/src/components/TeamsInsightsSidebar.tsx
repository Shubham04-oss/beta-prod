import { ChevronRight, Clock, Settings2, ListChecks, ShieldCheck } from "lucide-react";

export type TeamTask = {
  id: string;
  team_id: string;
  title: string;
  status: string;
  priority: string;
  requires_approval: boolean;
  due_at?: string;
  created_at: string;
};

export type TeamInsightRecord = {
  id: string;
  name: string;
  description: string;
  status: string;
  cadence_minutes: number;
  created_at: string;
  updated_at: string;
};

type TeamsInsightsSidebarProps = {
  team?: TeamInsightRecord | null;
  tasks: TeamTask[];
};

export function TeamsInsightsSidebar({ team, tasks }: TeamsInsightsSidebarProps) {
  const teamTasks = team ? tasks.filter((task) => task.team_id === team.id) : [];
  const approvals = teamTasks.filter((task) => task.requires_approval).length;

  return (
    <div className="flex flex-col h-full w-full">
      <div className="flex-1 overflow-y-auto pr-2 custom-scrollbar flex flex-col gap-8 pb-10">
        <div>
          <h4 className="font-semibold text-[13px] text-zinc-100 mb-4">Selected Team</h4>
          <div className="bg-white/[0.02] border border-white/5 p-4 rounded-xl">
            {team ? (
              <>
                <div className="flex items-center gap-3">
                  <div className="w-10 h-10 rounded-xl bg-purple-500/20 text-purple-300 flex items-center justify-center border border-purple-500/20">
                    <ShieldCheck className="w-5 h-5" />
                  </div>
                  <div className="min-w-0">
                    <span className="text-[13px] font-semibold text-white truncate block">{team.name}</span>
                    <span className="text-[10px] text-zinc-400 capitalize">{team.status.toLowerCase()}</span>
                  </div>
                </div>
                <p className="text-[11px] text-zinc-500 mt-4 leading-relaxed">{team.description || "No description configured."}</p>
              </>
            ) : (
              <p className="text-[11px] text-zinc-500">No team selected.</p>
            )}
          </div>
        </div>

        <div>
          <h4 className="font-semibold text-[13px] text-zinc-100 mb-4">Schedule</h4>
          <div className="flex flex-col gap-3">
            <div className="flex items-center gap-2 text-[11px] text-zinc-400">
              <Clock className="w-3.5 h-3.5" />
              <span>{team ? `Every ${team.cadence_minutes} minutes` : "No cadence configured"}</span>
            </div>
            <div>
              <span className="inline-block px-2.5 py-1 rounded bg-purple-500/20 text-purple-400 text-[10px] font-semibold">
                Temporal-ready configuration
              </span>
            </div>
          </div>
        </div>

        <div>
          <div className="flex items-center justify-between mb-4">
            <h4 className="font-semibold text-[13px] text-zinc-100">Task Queue</h4>
            <span className="text-[10px] text-zinc-400 font-medium">{teamTasks.length}</span>
          </div>
          <div className="flex flex-col gap-3">
            {teamTasks.length === 0 ? (
              <div className="text-[11px] text-zinc-500 bg-white/[0.02] border border-white/5 rounded-xl p-4">
                No delegated tasks for this team.
              </div>
            ) : (
              teamTasks.slice(0, 6).map((task) => (
                <div key={task.id} className="flex items-start gap-3 group">
                  <ListChecks className="w-3.5 h-3.5 text-zinc-500 mt-0.5 flex-shrink-0" />
                  <div className="flex flex-col flex-1 min-w-0">
                    <span className="text-[12px] font-semibold text-zinc-200 group-hover:text-white transition-colors truncate">{task.title}</span>
                    <span className="text-[10px] text-zinc-500 capitalize">{task.status.toLowerCase()} • {task.priority.toLowerCase()}</span>
                  </div>
                </div>
              ))
            )}
          </div>
          <button className="text-[10px] font-medium text-zinc-400 hover:text-white transition-colors flex items-center justify-between w-full mt-4">
            View all tasks <ChevronRight className="w-3 h-3" />
          </button>
        </div>

        <div>
          <h4 className="font-semibold text-[13px] text-zinc-100 mb-4">Approval Gates</h4>
          <div className="grid grid-cols-2 gap-2">
            <MetricTile label="Pending" value={approvals.toString()} />
            <MetricTile label="Total Tasks" value={teamTasks.length.toString()} />
          </div>
        </div>

        <div>
          <h4 className="font-semibold text-[13px] text-zinc-100 mb-4">Quick Actions</h4>
          <button className="flex items-center gap-2 p-2 rounded-lg bg-white/5 border border-white/5 hover:bg-white/10 transition-colors text-[10px] text-zinc-300 font-medium justify-center w-full">
            <Settings2 className="w-3.5 h-3.5" />
            Team Settings
          </button>
        </div>
      </div>
    </div>
  );
}

function MetricTile({ label, value }: { label: string; value: string }) {
  return (
    <div className="p-3 rounded-lg bg-white/5 border border-white/5">
      <p className="text-[10px] text-zinc-500">{label}</p>
      <p className="text-lg font-bold text-white mt-1">{value}</p>
    </div>
  );
}
