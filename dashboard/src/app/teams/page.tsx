"use client";

import { useEffect, useMemo, useState } from "react";
import { TeamsLeftSidebar, type TeamNavItem } from "@/components/TeamsLeftSidebar";
import { TeamsInsightsSidebar, type TeamInsightRecord, type TeamTask } from "@/components/TeamsInsightsSidebar";
import { Megaphone, MoreHorizontal, CheckCircle2, Circle, Users, ListChecks, ShieldCheck, Clock } from "lucide-react";
import { fetchAPI } from "@/lib/api";

type TeamWorkspaceResponse = {
  teams: TeamInsightRecord[];
  tasks: TeamTask[];
};

export default function TeamsPage() {
  const [teams, setTeams] = useState<TeamInsightRecord[]>([]);
  const [tasks, setTasks] = useState<TeamTask[]>([]);
  const [activeTeamID, setActiveTeamID] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let mounted = true;
    fetchAPI<TeamWorkspaceResponse>("/api/v1/team-workspace")
      .then((data) => {
        if (!mounted) return;
        const nextTeams = data.teams ?? [];
        setTeams(nextTeams);
        setTasks(data.tasks ?? []);
        setActiveTeamID((current) => current ?? nextTeams[0]?.id ?? null);
        setError(null);
      })
      .catch((err: Error) => {
        if (!mounted) return;
        setError(err.message || "Failed to load team workspace.");
      })
      .finally(() => {
        if (mounted) setLoading(false);
      });
    return () => {
      mounted = false;
    };
  }, []);

  const activeTeam = teams.find((team) => team.id === activeTeamID) ?? null;
  const activeTasks = activeTeam ? tasks.filter((task) => task.team_id === activeTeam.id) : [];
  const pendingApprovals = tasks.filter((task) => task.requires_approval && task.status !== "COMPLETED").length;
  const completedTasks = tasks.filter((task) => task.status === "COMPLETED").length;
  const navTeams: TeamNavItem[] = teams.map((team) => ({
    id: team.id,
    name: team.name,
    status: team.status,
    cadence_minutes: team.cadence_minutes,
  }));

  const kpis = useMemo(() => [
    { label: "Configured Teams", value: teams.length.toLocaleString(), icon: Users, subtext: "Tenant-scoped squads" },
    { label: "Delegated Tasks", value: tasks.length.toLocaleString(), icon: ListChecks, subtext: "Tracked in Postgres" },
    { label: "Approval Gates", value: pendingApprovals.toLocaleString(), icon: ShieldCheck, subtext: "Awaiting review" },
    { label: "Completed Tasks", value: completedTasks.toLocaleString(), icon: CheckCircle2, subtext: "Closed work items" },
  ], [completedTasks, pendingApprovals, tasks.length, teams.length]);

  return (
    <>
      <TeamsLeftSidebar teams={navTeams} activeTeamID={activeTeamID} onSelectTeam={setActiveTeamID} />

      <main className="flex-1 overflow-y-auto custom-scrollbar pr-4 flex flex-col gap-6">
        <div className="flex items-start justify-between mt-2">
          <div className="flex gap-4 items-center">
            <div className="w-12 h-12 rounded-2xl bg-purple-100 dark:bg-purple-900/30 text-purple-600 dark:text-purple-400 flex items-center justify-center flex-shrink-0 shadow-sm border border-purple-200 dark:border-purple-800/30">
              <Megaphone className="w-6 h-6" />
            </div>
            <div className="flex flex-col">
              <h1 className="text-2xl font-bold tracking-tight">Team Workspace</h1>
              <p className="text-[11px] text-muted-foreground mt-0.5">Human-managed squads and delegated task queues backed by tenant-scoped Postgres records.</p>
            </div>
          </div>
          <div className="flex items-center gap-3">
            <div className="flex items-center gap-1.5 px-3 py-1.5 rounded-full bg-green-500/10 text-green-600 dark:text-green-500 text-[11px] font-semibold border border-green-500/20">
              {loading ? "Loading" : "Synced"}
            </div>
            <button className="w-8 h-8 rounded-full border border-black/5 dark:border-white/10 flex items-center justify-center text-muted-foreground hover:bg-black/5 dark:hover:bg-white/5 transition-colors" aria-label="Team workspace actions">
              <MoreHorizontal className="w-4 h-4" />
            </button>
          </div>
        </div>

        {error && (
          <div className="p-3 rounded-xl text-sm bg-red-500/10 text-red-600 dark:text-red-400 border border-red-500/20">
            {error}
          </div>
        )}

        <div className="grid grid-cols-4 gap-4 mt-2">
          {kpis.map((kpi) => (
            <div key={kpi.label} className="bg-white/40 dark:bg-black/20 rounded-[1.5rem] p-5 border border-black/5 dark:border-white/5 flex items-center gap-4 min-w-0">
              <div className="w-10 h-10 rounded-xl bg-white/60 dark:bg-white/5 border border-black/5 dark:border-white/5 flex items-center justify-center flex-shrink-0">
                <kpi.icon className="w-5 h-5 text-purple-500" />
              </div>
              <div className="min-w-0">
                <span className="text-[11px] font-semibold text-muted-foreground truncate block">{kpi.label}</span>
                <span className="text-xl font-bold text-foreground mt-1 block">{kpi.value}</span>
                <span className="text-[9px] font-medium text-muted-foreground truncate block">{kpi.subtext}</span>
              </div>
            </div>
          ))}
        </div>

        <div className="grid grid-cols-2 gap-4">
          <section className="bg-white/40 dark:bg-black/20 rounded-[1.5rem] p-6 border border-black/5 dark:border-white/5 flex flex-col h-full">
            <h3 className="font-semibold text-[15px] mb-6">Configured Teams</h3>
            <div className="flex flex-col gap-4 flex-1">
              {loading ? (
                <p className="text-sm text-muted-foreground">Loading team workspace...</p>
              ) : teams.length === 0 ? (
                <EmptyState title="No teams configured" text="Create a team workspace through the API to start tracking delegated work." />
              ) : (
                teams.map((team) => (
                  <button key={team.id} onClick={() => setActiveTeamID(team.id)} className="flex gap-4 group text-left">
                    <div className="w-10 h-10 rounded-full border border-black/5 dark:border-white/5 shadow-sm bg-white dark:bg-black flex items-center justify-center">
                      <Users className="w-4 h-4 text-purple-500" />
                    </div>
                    <div className="flex flex-col flex-1 min-w-0">
                      <span className="text-[13px] font-bold text-foreground truncate">{team.name}</span>
                      <span className="text-[11px] text-muted-foreground mt-0.5 group-hover:text-foreground transition-colors line-clamp-1">{team.description || "No description configured."}</span>
                      <div className="flex items-center gap-1.5 mt-1.5">
                        <div className={`w-1.5 h-1.5 rounded-full ${team.status === "ACTIVE" ? "bg-green-500" : "bg-zinc-400"}`}></div>
                        <span className="text-[10px] text-muted-foreground capitalize">{team.status.toLowerCase()} • every {team.cadence_minutes} mins</span>
                      </div>
                    </div>
                  </button>
                ))
              )}
            </div>
          </section>

          <section className="bg-white/40 dark:bg-black/20 rounded-[1.5rem] p-6 border border-black/5 dark:border-white/5 flex flex-col h-full">
            <div className="flex items-center justify-between mb-4">
              <h3 className="font-semibold text-[15px]">Delegated Tasks</h3>
              <span className="text-[10px] font-medium text-muted-foreground">{activeTasks.length} active view</span>
            </div>
            <div className="flex flex-col gap-4 flex-1">
              {activeTeam ? (
                activeTasks.length === 0 ? (
                  <EmptyState title="No tasks delegated" text="Tasks created through the Team Workspace API will appear here." />
                ) : (
                  activeTasks.map((task) => (
                    <div key={task.id} className="flex gap-3 group">
                      <div className="mt-0.5 text-muted-foreground group-hover:text-foreground transition-colors flex-shrink-0">
                        {task.status === "COMPLETED" ? <CheckCircle2 className="w-4 h-4 text-green-500" /> : <Circle className="w-4 h-4" />}
                      </div>
                      <div className="flex flex-col flex-1 min-w-0">
                        <span className="text-[12px] font-medium text-foreground group-hover:text-primary truncate">{task.title}</span>
                        <span className="text-[10px] text-muted-foreground mt-0.5 capitalize">
                          {task.status.toLowerCase()} • {task.priority.toLowerCase()} {task.requires_approval ? "• approval required" : ""}
                        </span>
                      </div>
                    </div>
                  ))
                )
              ) : (
                <EmptyState title="Select a team" text="Choose a configured team to inspect its task queue." />
              )}
            </div>
          </section>
        </div>

        <section className="bg-white/40 dark:bg-black/20 rounded-[1.5rem] p-6 border border-black/5 dark:border-white/5 mb-6">
          <div className="flex items-center justify-between mb-6">
            <h3 className="font-semibold text-[15px]">Execution Readiness</h3>
            <div className="flex items-center gap-1.5 px-2.5 py-1 rounded-full bg-purple-500/10 text-purple-600 dark:text-purple-400 text-[10px] font-semibold">
              <Clock className="w-3 h-3" />
              Temporal orchestration ready
            </div>
          </div>
          <div className="grid grid-cols-3 gap-4">
            <ReadinessTile label="Tenant Isolation" value="Enforced" />
            <ReadinessTile label="Org Boundary" value="Enforced" />
            <ReadinessTile label="Audit Path" value="Settings tracked" />
          </div>
        </section>
      </main>

      <aside className="w-[360px] flex-shrink-0 bg-[#121212] text-white p-6 rounded-3xl relative z-20 flex flex-col h-full shadow-[0_8px_30px_rgb(0,0,0,0.4)] border border-white/5">
        <TeamsInsightsSidebar team={activeTeam} tasks={tasks} />
      </aside>
    </>
  );
}

function EmptyState({ title, text }: { title: string; text: string }) {
  return (
    <div className="rounded-xl bg-white/40 dark:bg-white/5 border border-black/5 dark:border-white/5 p-4">
      <p className="text-[13px] font-semibold text-foreground">{title}</p>
      <p className="text-[11px] text-muted-foreground mt-1">{text}</p>
    </div>
  );
}

function ReadinessTile({ label, value }: { label: string; value: string }) {
  return (
    <div className="bg-white/60 dark:bg-white/5 rounded-xl p-4 border border-black/5 dark:border-white/5">
      <span className="text-[11px] font-semibold text-muted-foreground">{label}</span>
      <span className="block text-xl font-bold text-foreground mt-2">{value}</span>
    </div>
  );
}
