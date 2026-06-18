import { 
  Globe, 
  Layers, 
  Puzzle, 
  GitMerge, 
  Receipt, 
  Bell, 
  BarChart2, 
  ArrowRightLeft, 
  Settings, 
  Zap, 
  ArrowRight 
} from "lucide-react";
import type { LucideIcon } from "lucide-react";

const managementNav = [
  { id: "all", name: "All Channels", icon: Globe },
  { id: "groups", name: "Channel Groups", icon: Layers },
  { id: "integrations", name: "Integrations", icon: Puzzle },
  { id: "mappings", name: "Mappings", icon: GitMerge },
  { id: "fees", name: "Fees & Commissions", icon: Receipt },
  { id: "notifications", name: "Notifications", icon: Bell },
] as const;

const performanceNav = [
  { id: "analytics", name: "Channel Analytics", icon: BarChart2 },
  { id: "comparison", name: "Comparison", icon: ArrowRightLeft },
] as const;

const configNav = [
  { id: "settings", name: "Settings", icon: Settings },
  { id: "automation", name: "Automation Rules", icon: Zap },
] as const;

export type ChannelsNavItemId =
  | (typeof managementNav)[number]["id"]
  | (typeof performanceNav)[number]["id"]
  | (typeof configNav)[number]["id"];

type ChannelNavItem = {
  id: ChannelsNavItemId;
  name: string;
  icon: LucideIcon;
};

type ChannelsLeftSidebarProps = {
  activeItem: ChannelsNavItemId;
  onItemChange: (item: ChannelsNavItemId) => void;
  healthSummary: {
    total: number;
    healthy: number;
    warning: number;
    error: number;
  };
};

export function ChannelsLeftSidebar({ activeItem, onItemChange, healthSummary }: ChannelsLeftSidebarProps) {
  const healthyPercent = getDonutPercent(healthSummary.healthy, healthSummary.total);
  const warningPercent = getDonutPercent(healthSummary.warning, healthSummary.total);
  const errorPercent = getDonutPercent(healthSummary.error, healthSummary.total);

  return (
    <aside className="w-[240px] h-full flex-shrink-0 flex flex-col pr-4 overflow-y-auto custom-scrollbar justify-center">
      
      {/* Management Section */}
      <div className="mb-8">
        <h3 className="text-[10px] font-bold text-muted-foreground uppercase tracking-wider mb-4 px-2">Channel Management</h3>
        <nav className="flex flex-col gap-1">
          {managementNav.map((item) => (
            <ChannelsNavButton
              key={item.name} 
              item={item}
              active={activeItem === item.id}
              onClick={() => onItemChange(item.id)}
            />
          ))}
        </nav>
      </div>

      {/* Performance Section */}
      <div className="mb-8">
        <h3 className="text-[10px] font-bold text-muted-foreground uppercase tracking-wider mb-4 px-2">Performance</h3>
        <nav className="flex flex-col gap-1">
          {performanceNav.map((item) => (
            <ChannelsNavButton
              key={item.name} 
              item={item}
              active={activeItem === item.id}
              onClick={() => onItemChange(item.id)}
            />
          ))}
        </nav>
      </div>

      {/* Configuration Section */}
      <div className="mb-8">
        <h3 className="text-[10px] font-bold text-muted-foreground uppercase tracking-wider mb-4 px-2">Configuration</h3>
        <nav className="flex flex-col gap-1">
          {configNav.map((item) => (
            <ChannelsNavButton
              key={item.name} 
              item={item}
              active={activeItem === item.id}
              onClick={() => onItemChange(item.id)}
            />
          ))}
        </nav>
      </div>

      {/* Spacer removed for vertical centering */}
      <div className="mt-8" />

      {/* Channel Health Card */}
      <div className="bg-white/40 dark:bg-black/20 backdrop-blur-md rounded-2xl p-4 border border-white/40 dark:border-white/10 mt-6 flex flex-col">
        <h4 className="text-xs font-semibold text-foreground mb-4">Channel Health</h4>
        
        <div className="flex items-center gap-4 mb-4">
          <div className="relative w-14 h-14 flex-shrink-0">
            {/* Simple CSS Donut representation */}
            <svg viewBox="0 0 36 36" className="w-full h-full transform -rotate-90">
              <path
                className="text-black/5 dark:text-white/5"
                d="M18 2.0845 a 15.9155 15.9155 0 0 1 0 31.831 a 15.9155 15.9155 0 0 1 0 -31.831"
                fill="none"
                stroke="currentColor"
                strokeWidth="4"
              />
              <path
                className="text-green-500"
                strokeDasharray={`${healthyPercent}, 100`}
                d="M18 2.0845 a 15.9155 15.9155 0 0 1 0 31.831 a 15.9155 15.9155 0 0 1 0 -31.831"
                fill="none"
                stroke="currentColor"
                strokeWidth="4"
              />
              <path
                className="text-amber-500"
                strokeDasharray={`${warningPercent}, 100`}
                strokeDashoffset={`-${healthyPercent}`}
                d="M18 2.0845 a 15.9155 15.9155 0 0 1 0 31.831 a 15.9155 15.9155 0 0 1 0 -31.831"
                fill="none"
                stroke="currentColor"
                strokeWidth="4"
              />
              <path
                className="text-red-500"
                strokeDasharray={`${errorPercent}, 100`}
                strokeDashoffset={`-${healthyPercent + warningPercent}`}
                d="M18 2.0845 a 15.9155 15.9155 0 0 1 0 31.831 a 15.9155 15.9155 0 0 1 0 -31.831"
                fill="none"
                stroke="currentColor"
                strokeWidth="4"
              />
            </svg>
            <div className="absolute inset-0 flex flex-col items-center justify-center">
              <span className="text-sm font-bold leading-none">{healthSummary.total}</span>
              <span className="text-[8px] text-muted-foreground">Total</span>
            </div>
          </div>
          
          <div className="flex flex-col gap-1.5 flex-1">
            <div className="flex items-center justify-between text-[10px]">
              <div className="flex items-center gap-1.5"><div className="w-1.5 h-1.5 rounded-full bg-green-500"></div><span className="text-muted-foreground">Healthy</span></div>
              <span className="font-semibold text-foreground">{healthSummary.healthy}</span>
            </div>
            <div className="flex items-center justify-between text-[10px]">
              <div className="flex items-center gap-1.5"><div className="w-1.5 h-1.5 rounded-full bg-amber-500"></div><span className="text-muted-foreground">Warning</span></div>
              <span className="font-semibold text-foreground">{healthSummary.warning}</span>
            </div>
            <div className="flex items-center justify-between text-[10px]">
              <div className="flex items-center gap-1.5"><div className="w-1.5 h-1.5 rounded-full bg-red-500"></div><span className="text-muted-foreground">Error</span></div>
              <span className="font-semibold text-foreground">{healthSummary.error}</span>
            </div>
          </div>
        </div>

        <button
          type="button"
          onClick={() => onItemChange("integrations")}
          className="flex items-center justify-center gap-2 w-full py-1.5 text-xs font-medium text-foreground border border-black/5 dark:border-white/10 rounded-lg hover:bg-black/5 dark:hover:bg-white/5 transition-colors"
        >
          View Health Report <ArrowRight className="w-3 h-3" />
        </button>
      </div>
      
    </aside>
  );
}

function ChannelsNavButton({ item, active, onClick }: { item: ChannelNavItem; active: boolean; onClick: () => void }) {
  return (
    <button
      type="button"
      onClick={onClick}
      aria-current={active ? "page" : undefined}
      className={`flex items-center gap-3 px-3 py-2 rounded-xl text-sm transition-colors text-left ${
        active
          ? "bg-primary/10 text-primary font-medium"
          : "text-muted-foreground hover:bg-black/5 dark:hover:bg-white/5 hover:text-foreground"
      }`}
    >
      <item.icon className={`w-4 h-4 ${active ? "text-primary" : "text-muted-foreground"}`} />
      {item.name}
    </button>
  );
}

function getDonutPercent(value: number, total: number) {
  return total > 0 ? (value / total) * 100 : 0;
}
