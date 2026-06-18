import { Plus, Users, Clock } from "lucide-react";

export type TeamNavItem = {
  id: string;
  name: string;
  status: string;
  cadence_minutes: number;
};

type TeamsLeftSidebarProps = {
  teams: TeamNavItem[];
  activeTeamID?: string | null;
  onSelectTeam?: (teamID: string) => void;
};

export function TeamsLeftSidebar({ teams, activeTeamID, onSelectTeam }: TeamsLeftSidebarProps) {
  return (
    <aside className="w-[280px] h-full flex-shrink-0 flex flex-col pr-4 overflow-y-auto custom-scrollbar justify-center">
      <div className="flex items-center justify-between mb-6 px-2 mt-2">
        <h2 className="text-[15px] font-bold tracking-tight text-foreground">AI Teams</h2>
        <button className="w-6 h-6 rounded-full hover:bg-black/5 dark:hover:bg-white/5 flex items-center justify-center text-muted-foreground hover:text-foreground transition-colors" aria-label="Create team">
          <Plus className="w-4 h-4" />
        </button>
      </div>

      <nav className="flex flex-col gap-1 mb-8">
        {teams.length === 0 ? (
          <div className="px-3 py-4 rounded-xl bg-white/40 dark:bg-black/20 border border-black/5 dark:border-white/5">
            <p className="text-[13px] font-semibold text-foreground">No squads configured</p>
            <p className="text-[10px] text-muted-foreground mt-1">Create a team workspace to start delegating tracked tasks.</p>
          </div>
        ) : (
          teams.map((team) => {
            const isActive = team.id === activeTeamID;
            return (
              <button
                key={team.id}
                type="button"
                onClick={() => onSelectTeam?.(team.id)}
                className={`flex items-center gap-3 px-3 py-3 rounded-xl transition-colors text-left ${
                  isActive
                    ? "bg-primary/10 text-primary shadow-sm ring-1 ring-black/5 dark:ring-white/10 bg-white dark:bg-black/40"
                    : "text-muted-foreground hover:bg-black/5 dark:hover:bg-white/5 hover:text-foreground"
                }`}
              >
                <div className={`w-8 h-8 rounded-full flex items-center justify-center flex-shrink-0 ${
                  isActive ? "bg-purple-100 dark:bg-purple-900/30 text-purple-600 dark:text-purple-400" : "bg-black/5 dark:bg-white/5"
                }`}>
                  <Users className="w-4 h-4" />
                </div>
                <div className="flex flex-col min-w-0">
                  <span className={`text-[13px] font-semibold truncate ${isActive ? "text-foreground" : ""}`}>
                    {team.name}
                  </span>
                  <div className="flex items-center gap-1.5 mt-0.5">
                    <div className={`w-1.5 h-1.5 rounded-full ${team.status === "ACTIVE" ? "bg-green-500" : "bg-zinc-400"}`}></div>
                    <span className="text-[10px] text-muted-foreground capitalize">{team.status.toLowerCase()}</span>
                  </div>
                </div>
              </button>
            );
          })
        )}
      </nav>

      <div className="mt-8" />

      <div className="bg-white/40 dark:bg-black/20 backdrop-blur-md rounded-2xl p-4 border border-black/5 dark:border-white/5 mt-6">
        <h4 className="text-[11px] font-semibold text-foreground">Workspace Cadence</h4>
        <p className="text-[10px] text-muted-foreground mt-0.5">Configured schedule</p>
        <div className="flex items-end gap-2 mt-4 mb-2">
          <span className="text-2xl font-bold text-foreground leading-none">{averageCadence(teams)}</span>
          <span className="text-[10px] font-medium text-muted-foreground flex items-center mb-0.5">
            <Clock className="w-3 h-3 mr-1" /> average minutes
          </span>
        </div>
      </div>
    </aside>
  );
}

function averageCadence(teams: TeamNavItem[]) {
  if (teams.length === 0) return "0";
  return Math.round(teams.reduce((sum, team) => sum + team.cadence_minutes, 0) / teams.length).toString();
}
