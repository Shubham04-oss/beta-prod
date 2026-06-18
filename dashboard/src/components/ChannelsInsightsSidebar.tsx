import { AlertCircle, BarChart2, CheckCircle2, ChevronRight, GitMerge, Globe, RefreshCw, Settings } from "lucide-react";
import { Area, AreaChart, ResponsiveContainer } from "recharts";
import type { ChannelsNavItemId } from "./ChannelsLeftSidebar";

type SalesChannel = {
  id: string;
  name: string;
  currency: string;
  active: boolean;
  created_at: string;
  updated_at: string;
};

type CommerceConnection = {
  id: string;
  unified_connection_id: string;
  provider: string;
  status: string;
};

type ChannelsInsightsSidebarProps = {
  channels: SalesChannel[];
  connections: CommerceConnection[];
  loading: boolean;
  error: string | null;
  onNavigate: (item: ChannelsNavItemId) => void;
};

export function ChannelsInsightsSidebar({ channels, connections, loading, error, onNavigate }: ChannelsInsightsSidebarProps) {
  const primaryChannel = channels.find((channel) => channel.active) ?? channels[0] ?? null;
  const activeChannels = channels.filter((channel) => channel.active).length;
  const inactiveChannels = channels.length - activeChannels;
  const activeConnections = connections.filter(isActiveConnection).length;
  const connectionWarnings = connections.length - activeConnections;
  const currencies = new Set(channels.map((channel) => channel.currency).filter(Boolean)).size;
  const activityData = buildActivityData(channels);
  const yAxisLabels = buildYAxisLabels(activityData);
  const alerts = buildAlerts({ channels, connections, loading, error });
  const health = getConnectionHealth({ loading, error, channels, activeConnections, connectionWarnings });
  const HealthIcon = health.tone === "healthy" ? CheckCircle2 : AlertCircle;

  return (
    <div className="flex flex-col h-full w-full">
      <div className="flex items-center justify-between mb-6">
        <h2 className="font-semibold text-sm leading-tight tracking-tight">Channel Overview</h2>
      </div>

      <div className="flex-1 overflow-y-auto pr-2 custom-scrollbar flex flex-col gap-8 pb-10">
        <div className="flex items-center gap-4">
          <div className="w-12 h-12 rounded-xl bg-white flex items-center justify-center p-2 flex-shrink-0 border border-white/10">
            {primaryChannel ? (
              <span className="text-sm font-bold text-zinc-900">{getInitials(primaryChannel.name)}</span>
            ) : (
              <Globe className="w-6 h-6 text-zinc-900" />
            )}
          </div>
          <div className="flex flex-col min-w-0">
            <h3 className="font-semibold text-lg text-white truncate">{primaryChannel?.name ?? (loading ? "Loading channels" : "No channels")}</h3>
            <p className="text-[10px] text-zinc-400 truncate">
              {primaryChannel ? `${primaryChannel.currency || "No currency"} sales channel` : "Tenant channel configuration"}
            </p>
            <div className="mt-1">
              <span className={`px-2 py-0.5 text-[9px] font-semibold rounded-full border ${getChannelStatusClass(primaryChannel)}`}>
                {primaryChannel ? (primaryChannel.active ? "Active" : "Inactive") : loading ? "Loading" : "Not configured"}
              </span>
            </div>
          </div>
        </div>

        <div>
          <h4 className="text-[10px] font-medium text-zinc-400 mb-3">Connection Health</h4>
          <div className="flex flex-col gap-3">
            <div className="flex items-start gap-2">
              <HealthIcon className={`w-4 h-4 mt-0.5 flex-shrink-0 ${health.iconClass}`} />
              <div className="flex flex-col">
                <span className="text-sm font-semibold text-zinc-200">{health.label}</span>
                <span className="text-[10px] text-zinc-500">{health.detail}</span>
              </div>
            </div>
            <button
              type="button"
              onClick={() => onNavigate("integrations")}
              className="text-[11px] font-semibold text-white bg-zinc-900 border border-zinc-800 rounded-lg py-2 hover:bg-zinc-800 transition-colors w-full"
            >
              View Integrations
            </button>
          </div>
        </div>

        <div>
          <h4 className="text-[10px] font-medium text-zinc-400 mb-3">Channel Activity (7 Days)</h4>
          <div className="flex flex-col gap-2 mb-4">
            <MetricRow label="Sales Channels" value={channels.length.toLocaleString()} detail={`${inactiveChannels.toLocaleString()} inactive`} warning={inactiveChannels > 0} />
            <MetricRow label="Active Channels" value={activeChannels.toLocaleString()} detail={`${channels.length.toLocaleString()} total`} />
            <MetricRow label="Connections" value={activeConnections.toLocaleString()} detail={`${connections.length.toLocaleString()} linked`} warning={connectionWarnings > 0} />
            <MetricRow label="Currencies" value={currencies.toLocaleString()} detail="configured" />
          </div>
          
          <div className="h-24 w-full relative">
            <div className="absolute left-0 inset-y-0 flex flex-col justify-between text-[8px] text-zinc-600 z-10 py-1">
              {yAxisLabels.map((label, index) => (
                <span key={`${label}-${index}`}>{label}</span>
              ))}
            </div>
            <div className="absolute inset-0 pl-8">
              <ResponsiveContainer width="100%" height="100%">
                <AreaChart data={activityData} margin={{ top: 5, left: 0, right: 0, bottom: 0 }}>
                  <defs>
                    <linearGradient id="colorValuePerf" x1="0" y1="0" x2="0" y2="1">
                      <stop offset="5%" stopColor="#a855f7" stopOpacity={0.3}/>
                      <stop offset="95%" stopColor="#a855f7" stopOpacity={0}/>
                    </linearGradient>
                  </defs>
                  <Area type="monotone" dataKey="value" stroke="#a855f7" strokeWidth={2} fillOpacity={1} fill="url(#colorValuePerf)" />
                </AreaChart>
              </ResponsiveContainer>
            </div>
          </div>
          <div className="flex justify-between text-[8px] text-zinc-600 mt-1 pl-8">
            {activityData.map((point) => (
              <span key={point.name}>{point.name}</span>
            ))}
          </div>
        </div>

        <div>
          <h4 className="text-[10px] font-medium text-zinc-400 mb-2">Quick Actions</h4>
          <div className="flex flex-col gap-1">
            <QuickAction icon={Settings} label="Edit Channel Settings" onClick={() => onNavigate("settings")} />
            <QuickAction icon={GitMerge} label="Manage Mappings" onClick={() => onNavigate("mappings")} />
            <QuickAction icon={RefreshCw} label="Review Sync Status" onClick={() => onNavigate("integrations")} />
            <QuickAction icon={BarChart2} label="View Channel Analytics" onClick={() => onNavigate("analytics")} />
          </div>
        </div>

        <div>
          <h4 className="text-[10px] font-medium text-zinc-400 mb-3">Recent Alerts</h4>
          <div className="flex flex-col gap-3">
            {alerts.map((alert) => (
              <div key={alert.message} className="flex items-start justify-between gap-4">
                <div className="flex items-center gap-2">
                  <div className={`w-1.5 h-1.5 rounded-full mt-1 flex-shrink-0 ${alert.color}`}></div>
                  <span className="text-[11px] text-zinc-300">{alert.message}</span>
                </div>
                <span className="text-[9px] text-zinc-500 whitespace-nowrap">{alert.time}</span>
              </div>
            ))}
          </div>
          <button
            type="button"
            onClick={() => onNavigate("notifications")}
            className="text-[10px] font-medium text-zinc-400 hover:text-white transition-colors flex items-center justify-end w-full gap-1 mt-3"
          >
            View all alerts <ChevronRight className="w-3 h-3" />
          </button>
        </div>

      </div>
    </div>
  );
}

function MetricRow({ label, value, detail, warning = false }: { label: string; value: string; detail: string; warning?: boolean }) {
  return (
    <div className="flex items-center justify-between">
      <span className="text-[11px] text-zinc-400">{label}</span>
      <div className="flex items-center gap-2">
        <span className="text-xs font-semibold text-white">{value}</span>
        <span className={`text-[9px] font-medium ${warning ? "text-amber-500" : "text-zinc-500"}`}>{detail}</span>
      </div>
    </div>
  );
}

function QuickAction({ icon: Icon, label, onClick }: { icon: typeof Settings; label: string; onClick: () => void }) {
  return (
    <button type="button" onClick={onClick} className="flex items-center justify-between w-full py-2 group">
      <div className="flex items-center gap-3 text-zinc-300 group-hover:text-white transition-colors">
        <Icon className="w-3.5 h-3.5" />
        <span className="text-[11px]">{label}</span>
      </div>
      <ChevronRight className="w-3 h-3 text-zinc-600 group-hover:text-zinc-400" />
    </button>
  );
}

function buildActivityData(channels: SalesChannel[]) {
  const today = new Date();
  today.setHours(0, 0, 0, 0);

  return Array.from({ length: 7 }, (_, index) => {
    const day = new Date(today);
    day.setDate(today.getDate() - (6 - index));
    const nextDay = new Date(day);
    nextDay.setDate(day.getDate() + 1);
    const value = channels.filter((channel) => {
      const activityDate = parseDate(channel.updated_at || channel.created_at);
      return activityDate ? activityDate >= day && activityDate < nextDay : false;
    }).length;

    return {
      name: day.toLocaleDateString(undefined, { month: "short", day: "numeric" }),
      value,
    };
  });
}

function buildYAxisLabels(data: { value: number }[]) {
  const maxValue = Math.max(1, ...data.map((point) => point.value));
  return [maxValue, Math.ceil(maxValue * 0.66), Math.ceil(maxValue * 0.33), 0].map((value) => value.toLocaleString());
}

function buildAlerts({ channels, connections, loading, error }: Pick<ChannelsInsightsSidebarProps, "channels" | "connections" | "loading" | "error">) {
  if (loading) {
    return [{ message: "Loading channel status", time: "Now", color: "bg-amber-500" }];
  }

  if (error) {
    return [{ message: "Channel data failed to load", time: "Now", color: "bg-red-500" }];
  }

  if (channels.length === 0) {
    return [{ message: "No sales channels configured", time: "Current", color: "bg-amber-500" }];
  }

  const alerts = [];
  const inactiveChannels = channels.filter((channel) => !channel.active);
  const latestInactiveDate = getLatestChannelDate(inactiveChannels);
  const connectionWarnings = connections.filter((connection) => !isActiveConnection(connection));

  if (inactiveChannels.length > 0) {
    alerts.push({
      message: `${inactiveChannels.length.toLocaleString()} inactive channel${inactiveChannels.length === 1 ? "" : "s"}`,
      time: latestInactiveDate ? formatRelativeDate(latestInactiveDate) : "Current",
      color: "bg-amber-500",
    });
  }

  if (connectionWarnings.length > 0) {
    alerts.push({
      message: `${connectionWarnings.length.toLocaleString()} connection${connectionWarnings.length === 1 ? "" : "s"} need attention`,
      time: "Current",
      color: "bg-red-500",
    });
  }

  if (connections.length === 0) {
    alerts.push({
      message: "No unified connections linked",
      time: "Current",
      color: "bg-amber-500",
    });
  }

  return alerts.length > 0
    ? alerts
    : [{ message: "No channel alerts", time: "Current", color: "bg-green-500" }];
}

function getConnectionHealth({
  loading,
  error,
  channels,
  activeConnections,
  connectionWarnings,
}: {
  loading: boolean;
  error: string | null;
  channels: SalesChannel[];
  activeConnections: number;
  connectionWarnings: number;
}) {
  if (loading) {
    return { label: "Loading", detail: "Checking channel data", tone: "warning", iconClass: "text-amber-500" };
  }

  if (error) {
    return { label: "Unavailable", detail: error, tone: "error", iconClass: "text-red-500" };
  }

  if (channels.length === 0) {
    return { label: "Not configured", detail: "No sales channels found", tone: "warning", iconClass: "text-amber-500" };
  }

  if (connectionWarnings > 0 || activeConnections === 0) {
    return {
      label: "Needs attention",
      detail: `${activeConnections.toLocaleString()} active unified connection${activeConnections === 1 ? "" : "s"}`,
      tone: "warning",
      iconClass: "text-amber-500",
    };
  }

  return {
    label: "Healthy",
    detail: `${activeConnections.toLocaleString()} active unified connection${activeConnections === 1 ? "" : "s"}`,
    tone: "healthy",
    iconClass: "text-green-500",
  };
}

function getLatestChannelDate(channels: SalesChannel[]) {
  return channels.reduce<Date | null>((latest, channel) => {
    const date = parseDate(channel.updated_at || channel.created_at);
    if (!date) return latest;
    return !latest || date > latest ? date : latest;
  }, null);
}

function getInitials(value: string) {
  return value
    .split(/\s+/)
    .filter(Boolean)
    .slice(0, 2)
    .map((part) => part[0]?.toUpperCase())
    .join("") || "CH";
}

function getChannelStatusClass(channel: SalesChannel | null) {
  if (!channel) {
    return "bg-zinc-500/20 text-zinc-300 border-zinc-500/20";
  }

  return channel.active
    ? "bg-green-500/20 text-green-400 border-green-500/20"
    : "bg-amber-500/20 text-amber-400 border-amber-500/20";
}

function isActiveConnection(connection: CommerceConnection) {
  return connection.status?.toUpperCase() === "ACTIVE";
}

function parseDate(value: string) {
  const date = value ? new Date(value) : null;
  return date && Number.isFinite(date.getTime()) ? date : null;
}

function formatRelativeDate(date: Date) {
  const diffMs = Date.now() - date.getTime();
  const diffMinutes = Math.max(0, Math.floor(diffMs / 60000));

  if (diffMinutes < 1) return "Now";
  if (diffMinutes < 60) return `${diffMinutes}m ago`;

  const diffHours = Math.floor(diffMinutes / 60);
  if (diffHours < 24) return `${diffHours}h ago`;

  const diffDays = Math.floor(diffHours / 24);
  return `${diffDays}d ago`;
}
