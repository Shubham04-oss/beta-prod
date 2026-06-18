"use client";

import { useEffect, useMemo, useState } from "react";
import { ChannelsLeftSidebar, type ChannelsNavItemId } from "@/components/ChannelsLeftSidebar";
import { ChannelsInsightsSidebar } from "@/components/ChannelsInsightsSidebar";
import { 
  Plus, 
  Search, 
  Filter, 
  MoreHorizontal, 
  Settings, 
  Globe, 
  Link2, 
  Activity, 
  CircleDollarSign,
  Layers,
  Puzzle,
  GitMerge,
  Receipt,
  Bell,
  ArrowRightLeft,
  Zap,
  ArrowRight,
  Sparkles
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { fetchAPI } from "@/lib/api";

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

export default function ChannelsPage() {
  const [channels, setChannels] = useState<SalesChannel[]>([]);
  const [connections, setConnections] = useState<CommerceConnection[]>([]);
  const [query, setQuery] = useState("");
  const [statusFilter, setStatusFilter] = useState("all");
  const [activeNavItem, setActiveNavItem] = useState<ChannelsNavItemId>("all");
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let mounted = true;
    Promise.all([
      fetchAPI("/api/v1/channels") as Promise<{ channels?: SalesChannel[] }>,
      fetchAPI("/api/v1/integrations/connections") as Promise<{ connections?: CommerceConnection[] }>,
    ])
      .then(([channelPayload, connectionPayload]) => {
        if (!mounted) return;
        setChannels(channelPayload.channels ?? []);
        setConnections(connectionPayload.connections ?? []);
        setError(null);
      })
      .catch((err: Error) => {
        if (!mounted) return;
        setError(err.message || "Failed to load channels.");
      })
      .finally(() => {
        if (mounted) setLoading(false);
      });
    return () => {
      mounted = false;
    };
  }, []);

  const filteredChannels = useMemo(() => {
    const needle = query.trim().toLowerCase();
    return channels.filter((channel) => {
      const matchesStatus = statusFilter === "all" || (statusFilter === "active" ? channel.active : !channel.active);
      const matchesQuery = needle === "" || channel.name.toLowerCase().includes(needle) || channel.currency.toLowerCase().includes(needle);
      return matchesStatus && matchesQuery;
    });
  }, [channels, query, statusFilter]);

  const activeChannels = channels.filter((channel) => channel.active).length;
  const activeConnections = connections.filter((connection) => connection.status === "ACTIVE").length;
  const currencies = new Set(channels.map((channel) => channel.currency)).size;
  const inactiveChannels = channels.length - activeChannels;
  const failedConnections = connections.filter((connection) => connection.status && connection.status !== "ACTIVE").length;
  const healthSummary = {
    total: channels.length + connections.length,
    healthy: activeChannels + activeConnections,
    warning: inactiveChannels,
    error: error ? 1 : failedConnections,
  };

  const renderMainContent = () => {
    switch (activeNavItem) {
      case "all":
        return (
          <>
            <div className="flex items-center justify-between mt-2">
              <div>
                <h1 className="text-3xl font-bold tracking-tight">Sales Channels</h1>
                <p className="text-muted-foreground mt-1">Connect, manage and optimize your sales channels from one place.</p>
              </div>
              <div className="flex items-center gap-3">
                <Button className="rounded-full shadow-sm">
                  <Plus className="mr-2 h-4 w-4" />
                  Add Channel
                </Button>
                <Button variant="outline" size="icon" className="rounded-full bg-white/40 dark:bg-black/20 backdrop-blur-md border-white/40 dark:border-white/10 shadow-sm hover:bg-white/60 dark:hover:bg-white/10">
                  <MoreHorizontal className="h-4 w-4" />
                </Button>
              </div>
            </div>

            <div className="grid grid-cols-4 gap-4">
              <ChannelKPICard title="Sales Channels" value={channels.length.toLocaleString()} subtext="Tenant configured" icon={Globe} color="text-purple-500" />
              <ChannelKPICard title="Active Channels" value={activeChannels.toLocaleString()} subtext="Available for routing" icon={Activity} color="text-green-500" />
              <ChannelKPICard title="Unified Connections" value={activeConnections.toLocaleString()} subtext="OAuth connections" icon={Link2} color="text-blue-500" />
              <ChannelKPICard title="Currencies" value={currencies.toLocaleString()} subtext="Configured markets" icon={CircleDollarSign} color="text-amber-500" />
            </div>

            <div className="bg-white/20 dark:bg-white/5 rounded-[2rem] p-6 border border-white/30 dark:border-white/10 flex flex-col flex-1">
              <div className="flex items-center justify-between mb-6">
                <div className="relative w-64">
                  <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
                  <input
                    type="text"
                    value={query}
                    onChange={(event) => setQuery(event.target.value)}
                    placeholder="Search channels..."
                    className="w-full h-9 pl-9 pr-4 rounded-full bg-white/40 dark:bg-black/20 border border-black/5 dark:border-white/10 text-xs focus:outline-none focus:ring-2 focus:ring-primary/50"
                  />
                </div>

                <div className="flex items-center gap-3">
                  <select
                    value={statusFilter}
                    onChange={(event) => setStatusFilter(event.target.value)}
                    className="h-9 px-4 rounded-full bg-white/40 dark:bg-black/20 border border-black/5 dark:border-white/10 text-xs focus:outline-none"
                  >
                    <option value="all">All Status</option>
                    <option value="active">Active</option>
                    <option value="inactive">Inactive</option>
                  </select>
                  <Button variant="outline" size="sm" className="rounded-full bg-white/40 dark:bg-black/20 border-black/5 dark:border-white/10 h-9 text-xs font-medium">
                    <Filter className="w-3.5 h-3.5 mr-2" /> Filters
                  </Button>
                </div>
              </div>

              <div className="w-full overflow-x-auto">
                <table className="w-full text-sm text-left">
                  <thead>
                    <tr className="text-[11px] text-muted-foreground uppercase tracking-wider border-b border-black/5 dark:border-white/5">
                      <th className="pb-3 font-semibold px-2">Channel</th>
                      <th className="pb-3 font-semibold px-2">Currency</th>
                      <th className="pb-3 font-semibold px-2">Status</th>
                      <th className="pb-3 font-semibold px-2">Created</th>
                      <th className="pb-3 font-semibold px-2">Updated</th>
                      <th className="pb-3 font-semibold px-2 text-right">Actions</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-black/5 dark:divide-white/5">
                    {loading ? (
                      <tr>
                        <td colSpan={6} className="py-10 text-center text-muted-foreground">Loading sales channels...</td>
                      </tr>
                    ) : error ? (
                      <tr>
                        <td colSpan={6} className="py-10 text-center text-red-600 dark:text-red-400">{error}</td>
                      </tr>
                    ) : filteredChannels.length === 0 ? (
                      <tr>
                        <td colSpan={6} className="py-10 text-center text-muted-foreground">No sales channels match this view.</td>
                      </tr>
                    ) : (
                      filteredChannels.map((channel) => (
                        <tr key={channel.id} className="group hover:bg-black/[0.02] dark:hover:bg-white/[0.02] transition-colors cursor-pointer">
                          <td className="py-3 px-2 w-[240px]">
                            <div className="flex items-center gap-3">
                              <div className="w-8 h-8 rounded-lg bg-white flex items-center justify-center flex-shrink-0 border border-black/5 shadow-sm">
                                <Globe className="w-4 h-4 text-purple-500" />
                              </div>
                              <div className="flex flex-col">
                                <span className="font-semibold text-sm text-foreground">{channel.name}</span>
                                <span className="text-[10px] text-muted-foreground font-mono">{channel.id.slice(0, 8)}</span>
                              </div>
                            </div>
                          </td>
                          <td className="py-3 px-2 text-xs text-foreground font-medium">{channel.currency}</td>
                          <td className="py-3 px-2">
                            <span className={`px-2 py-1 rounded-full text-[10px] font-bold ${channel.active ? "text-green-500 bg-green-500/10" : "text-zinc-500 bg-zinc-500/10"}`}>
                              {channel.active ? "Active" : "Inactive"}
                            </span>
                          </td>
                          <td className="py-3 px-2 text-xs text-muted-foreground">{formatDate(channel.created_at)}</td>
                          <td className="py-3 px-2 text-xs text-muted-foreground">{formatDate(channel.updated_at)}</td>
                          <td className="py-3 px-2 text-right">
                            <div className="flex items-center justify-end gap-2 opacity-50 group-hover:opacity-100 transition-opacity">
                              <button className="text-muted-foreground hover:text-foreground p-1" aria-label="Channel settings">
                                <Settings className="w-3.5 h-3.5" />
                              </button>
                              <button className="text-muted-foreground hover:text-foreground p-1" aria-label="Channel actions">
                                <MoreHorizontal className="w-4 h-4" />
                              </button>
                            </div>
                          </td>
                        </tr>
                      ))
                    )}
                  </tbody>
                </table>
              </div>

              <div className="flex items-center justify-between mt-auto pt-6 text-xs text-muted-foreground">
                <span>Showing {filteredChannels.length} of {channels.length} channels</span>
                <span>{activeConnections} Unified.to connections active</span>
              </div>
            </div>
          </>
        );
      case "groups":
        return <ChannelGroupsView />;
      case "integrations":
        return <IntegrationsView connections={connections} />;
      case "mappings":
        return <MappingBuilderView />;
      case "fees":
        return <FeesCommissionsView />;
      case "notifications":
        return <NotificationsView />;
      case "settings":
        return <ChannelSettingsView onNavigate={setActiveNavItem} />;
      case "automation":
        return <AutomationRulesView />;
      default:
        return (
          <div className="flex-1 flex items-center justify-center text-muted-foreground">
            Select an item from the sidebar.
          </div>
        );
    }
  };

  return (
    <>
      <ChannelsLeftSidebar
        activeItem={activeNavItem}
        onItemChange={setActiveNavItem}
        healthSummary={healthSummary}
      />

      <main className="flex-1 overflow-y-auto custom-scrollbar pr-4 flex flex-col gap-6">
        {renderMainContent()}
      </main>

      <aside className="w-[360px] flex-shrink-0 bg-[#121212] text-white p-6 rounded-3xl relative z-20 flex flex-col h-full shadow-[0_8px_30px_rgb(0,0,0,0.4)] border border-white/5">
        <ChannelsInsightsSidebar
          channels={channels}
          connections={connections}
          loading={loading}
          error={error}
          onNavigate={setActiveNavItem}
        />
      </aside>
    </>
  );
}

function ChannelKPICard({ title, value, subtext, icon: Icon, color }: { title: string; value: string; subtext: string; icon: typeof Globe; color: string }) {
  return (
    <div className="bg-white/20 dark:bg-white/5 rounded-[2rem] p-6 border border-white/30 dark:border-white/10 flex items-center gap-4 min-w-0">
      <div className="w-12 h-12 rounded-2xl flex items-center justify-center border border-black/5 dark:border-white/5 bg-white/60 dark:bg-black/20 flex-shrink-0">
        <Icon className={`w-5 h-5 ${color}`} />
      </div>
      <div className="min-w-0">
        <p className="text-xs font-semibold text-muted-foreground uppercase tracking-wider truncate">{title}</p>
        <h3 className="text-2xl font-bold mt-1">{value}</h3>
        <p className="text-[10px] mt-1 font-semibold text-muted-foreground truncate">{subtext}</p>
      </div>
    </div>
  );
}

function formatDate(value: string) {
  return value ? new Date(value).toLocaleString() : "No timestamp";
}

// SUB-VIEWS FOR SIDEBAR TABS

function ChannelGroupsView() {
  const groups = [
    { id: 1, name: "US Storefronts", channels: 3, status: "Active", volume: "$124,500" },
    { id: 2, name: "EU Marketplaces", channels: 2, status: "Active", volume: "€89,200" },
    { id: 3, name: "B2B Wholesalers", channels: 1, status: "Inactive", volume: "$0" },
  ];
  return (
    <div className="flex flex-col gap-6">
      <div className="flex items-center justify-between mt-2">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Channel Groups</h1>
          <p className="text-muted-foreground mt-1">Organize your sales channels into logical sync groups.</p>
        </div>
        <Button className="rounded-full shadow-sm">
          <Plus className="mr-2 h-4 w-4" /> Create Group
        </Button>
      </div>

      <div className="grid grid-cols-3 gap-4">
        {groups.map((g) => (
          <div key={g.id} className="bg-white/20 dark:bg-white/5 rounded-[2rem] p-6 border border-white/30 dark:border-white/10 flex flex-col justify-between h-48">
            <div>
              <div className="flex items-center justify-between mb-4">
                <span className={`px-2 py-0.5 rounded-full text-[10px] font-bold ${g.status === "Active" ? "text-green-500 bg-green-500/10" : "text-zinc-500 bg-zinc-500/10"}`}>{g.status}</span>
                <span className="text-xs text-muted-foreground font-semibold">{g.channels} Channels</span>
              </div>
              <h3 className="text-lg font-bold text-foreground">{g.name}</h3>
              <p className="text-xs text-muted-foreground mt-1">Volume: {g.volume}</p>
            </div>
            <div className="flex justify-end gap-2">
              <Button variant="outline" size="sm" className="rounded-full h-8 text-xs">Edit</Button>
              <Button variant="outline" size="sm" className="rounded-full h-8 text-xs text-red-500 hover:text-red-600">Delete</Button>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}

function IntegrationsView({ connections }: { connections: CommerceConnection[] }) {
  const availableConnectors = [
    { provider: "shopify", name: "Shopify", desc: "Sync catalog, orders, and fulfillment." },
    { provider: "amazon", name: "Amazon US", desc: "Sync Amazon Merchant Fulfilled/FBA." },
    { provider: "woocommerce", name: "WooCommerce", desc: "Self-hosted store sync." },
    { provider: "bigcommerce", name: "BigCommerce", desc: "Multi-storefront enterprise sync." },
  ];

  return (
    <div className="flex flex-col gap-6">
      <div className="flex items-center justify-between mt-2">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">OAuth Integrations</h1>
          <p className="text-muted-foreground mt-1">Link and authorize third-party platforms via Unified.to secure OAuth flow.</p>
        </div>
        <Button className="rounded-full shadow-sm">
          <Plus className="mr-2 h-4 w-4" /> Connect Integration
        </Button>
      </div>

      <div className="bg-white/20 dark:bg-white/5 rounded-[2rem] p-6 border border-white/30 dark:border-white/10 flex flex-col gap-4">
        <h2 className="text-lg font-bold text-foreground mb-2">Connected Providers</h2>
        {connections.length === 0 ? (
          <p className="text-sm text-muted-foreground">No active OAuth connections found. Connect one below.</p>
        ) : (
          <div className="divide-y divide-black/5 dark:divide-white/5">
            {connections.map((c) => (
              <div key={c.id} className="py-4 flex items-center justify-between">
                <div className="flex items-center gap-3">
                  <div className="w-10 h-10 rounded-xl bg-white flex items-center justify-center border border-black/5 shadow-sm">
                    <Puzzle className="w-5 h-5 text-blue-500" />
                  </div>
                  <div>
                    <h4 className="font-semibold text-sm capitalize">{c.provider}</h4>
                    <p className="text-[10px] text-muted-foreground font-mono">{c.unified_connection_id}</p>
                  </div>
                </div>
                <div className="flex items-center gap-4">
                  <span className={`px-2 py-0.5 rounded-full text-[10px] font-bold ${c.status === "ACTIVE" ? "text-green-500 bg-green-500/10" : "text-amber-500 bg-amber-500/10"}`}>{c.status}</span>
                  <Button variant="outline" size="sm" className="rounded-full h-8 text-xs text-red-500">Disconnect</Button>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>

      <div className="grid grid-cols-2 gap-4">
        {availableConnectors.map((c) => (
          <div key={c.provider} className="bg-white/20 dark:bg-white/5 rounded-[2rem] p-6 border border-white/30 dark:border-white/10 flex items-center justify-between">
            <div>
              <h3 className="font-bold text-foreground">{c.name}</h3>
              <p className="text-xs text-muted-foreground mt-1 pr-4">{c.desc}</p>
            </div>
            <Button variant="outline" className="rounded-full shadow-sm">Connect</Button>
          </div>
        ))}
      </div>
    </div>
  );
}

function MappingBuilderView() {
  const [mappings] = useState([
    { id: "1", source: "SKU", target: "sku", type: "string", required: true },
    { id: "2", source: "Product Title", target: "title", type: "string", required: true },
    { id: "3", source: "Base Price", target: "price", type: "number", required: true },
    { id: "4", source: "Inventory Count", target: "inventory_quantity", type: "number", required: false },
  ]);

  return (
    <div className="flex flex-col gap-6">
      <div className="flex items-center justify-between mt-2">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Field Mapping Builder</h1>
          <p className="text-muted-foreground mt-1">Map your internal product fields to channel-specific properties.</p>
        </div>
        <div className="flex gap-2">
          <Button variant="outline" className="rounded-full shadow-sm bg-white/40 dark:bg-black/20 border-black/5 dark:border-white/10 hover:bg-white/60 dark:hover:bg-white/10">
            <Sparkles className="mr-2 h-4 w-4 text-purple-500" /> Auto-Map (AI)
          </Button>
          <Button className="rounded-full shadow-sm">Save Mappings</Button>
        </div>
      </div>

      <div className="bg-white/20 dark:bg-white/5 rounded-[2rem] p-6 border border-white/30 dark:border-white/10 flex flex-col">
        <div className="grid grid-cols-4 text-xs font-semibold text-muted-foreground uppercase tracking-wider pb-3 border-b border-black/5 dark:border-white/5 mb-4">
          <div>Synq Core Property</div>
          <div>Channel Field Target</div>
          <div>Data Type</div>
          <div className="text-right">Requirement</div>
        </div>
        <div className="flex flex-col gap-4">
          {mappings.map((m) => (
            <div key={m.id} className="grid grid-cols-4 items-center text-sm">
              <div className="font-semibold text-foreground">{m.source}</div>
              <div>
                <select defaultValue={m.target} className="h-9 px-3 w-48 rounded-xl bg-white/40 dark:bg-black/20 border border-black/5 dark:border-white/10 text-xs focus:outline-none">
                  <option value={m.target}>{m.target}</option>
                  <option value="none">-- Ignore Field --</option>
                  <option value="custom">-- Custom Field --</option>
                </select>
              </div>
              <div className="text-xs font-mono text-muted-foreground">{m.type}</div>
              <div className="text-right">
                <span className={`text-[10px] font-bold ${m.required ? "text-purple-500" : "text-muted-foreground"}`}>{m.required ? "REQUIRED" : "OPTIONAL"}</span>
              </div>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}

function FeesCommissionsView() {
  const rules = [
    { id: 1, channel: "Shopify US Store", type: "Referral Fee", rate: "1.5%", active: true },
    { id: 2, channel: "Amazon US Marketplace", type: "Category Fee", rate: "15.0%", active: true },
    { id: 3, channel: "WooCommerce Dev", type: "Payment Overhead", rate: "2.9% + $0.30", active: false },
  ];
  return (
    <div className="flex flex-col gap-6">
      <div className="flex items-center justify-between mt-2">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Fees & Commissions</h1>
          <p className="text-muted-foreground mt-1">Configure marketplace fee rules, referral percentages, and transaction overheads.</p>
        </div>
        <Button className="rounded-full shadow-sm">
          <Plus className="mr-2 h-4 w-4" /> Add Fee Rule
        </Button>
      </div>

      <div className="bg-white/20 dark:bg-white/5 rounded-[2rem] p-6 border border-white/30 dark:border-white/10">
        <table className="w-full text-sm text-left">
          <thead>
            <tr className="text-[11px] text-muted-foreground uppercase tracking-wider border-b border-black/5 dark:border-white/5">
              <th className="pb-3 px-2 font-semibold">Sales Channel</th>
              <th className="pb-3 px-2 font-semibold">Fee Type</th>
              <th className="pb-3 px-2 font-semibold">Calculation Rate</th>
              <th className="pb-3 px-2 font-semibold">Status</th>
              <th className="pb-3 px-2 font-semibold text-right">Actions</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-black/5 dark:divide-white/5">
            {rules.map((r) => (
              <tr key={r.id} className="hover:bg-black/[0.01] dark:hover:bg-white/[0.01]">
                <td className="py-3 px-2 font-semibold">{r.channel}</td>
                <td className="py-3 px-2 text-xs">{r.type}</td>
                <td className="py-3 px-2 text-xs font-mono">{r.rate}</td>
                <td className="py-3 px-2">
                  <span className={`px-2 py-0.5 rounded-full text-[10px] font-bold ${r.active ? "text-green-500 bg-green-500/10" : "text-zinc-500 bg-zinc-500/10"}`}>
                    {r.active ? "Active" : "Disabled"}
                  </span>
                </td>
                <td className="py-3 px-2 text-right">
                  <Button variant="outline" size="sm" className="rounded-full h-8 text-xs mr-2">Edit</Button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}

function NotificationsView() {
  const [webhookUrl, setWebhookUrl] = useState("https://api.merchant.com/v1/synq-webhook");
  const logs = [
    { id: 1, event: "order.created", status: 200, latency: "45ms", time: "2 minutes ago" },
    { id: 2, event: "inventory.updated", status: 200, latency: "122ms", time: "15 minutes ago" },
    { id: 3, event: "order.fulfilled", status: 500, latency: "1200ms", time: "1 hour ago" },
  ];
  return (
    <div className="flex flex-col gap-6">
      <div className="flex items-center justify-between mt-2">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Notifications & Webhooks</h1>
          <p className="text-muted-foreground mt-1">Configure event-driven webhook receivers and review transmission logs.</p>
        </div>
      </div>

      <div className="bg-white/20 dark:bg-white/5 rounded-[2rem] p-6 border border-white/30 dark:border-white/10 flex flex-col gap-4">
        <h3 className="font-bold text-foreground">Webhook Endpoint</h3>
        <div className="flex gap-3">
          <input
            type="text"
            value={webhookUrl}
            onChange={(e) => setWebhookUrl(e.target.value)}
            className="flex-1 h-9 px-4 rounded-full bg-white/40 dark:bg-black/20 border border-black/5 dark:border-white/10 text-xs focus:outline-none focus:ring-2 focus:ring-primary/50"
          />
          <Button className="rounded-full shadow-sm h-9">Update URL</Button>
        </div>
        <div className="flex items-center gap-4 mt-2">
          <label className="flex items-center gap-2 text-xs">
            <input type="checkbox" defaultChecked className="rounded border-gray-300 text-primary focus:ring-primary" /> Order Events
          </label>
          <label className="flex items-center gap-2 text-xs">
            <input type="checkbox" defaultChecked className="rounded border-gray-300 text-primary focus:ring-primary" /> Inventory Events
          </label>
          <label className="flex items-center gap-2 text-xs">
            <input type="checkbox" className="rounded border-gray-300 text-primary focus:ring-primary" /> Product Sync Events
          </label>
        </div>
      </div>

      <div className="bg-white/20 dark:bg-white/5 rounded-[2rem] p-6 border border-white/30 dark:border-white/10 flex flex-col">
        <h3 className="font-bold text-foreground mb-4">Webhook Delivery Logs</h3>
        <table className="w-full text-sm text-left">
          <thead>
            <tr className="text-[11px] text-muted-foreground uppercase tracking-wider border-b border-black/5 dark:border-white/5">
              <th className="pb-3 px-2 font-semibold">Event Topic</th>
              <th className="pb-3 px-2 font-semibold">HTTP Status</th>
              <th className="pb-3 px-2 font-semibold">Latency</th>
              <th className="pb-3 px-2 font-semibold text-right">Delivered</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-black/5 dark:divide-white/5">
            {logs.map((l) => (
              <tr key={l.id} className="hover:bg-black/[0.01] dark:hover:bg-white/[0.01]">
                <td className="py-3 px-2 font-mono text-xs text-foreground">{l.event}</td>
                <td className="py-3 px-2">
                  <span className={`px-2 py-0.5 rounded-full text-[10px] font-mono font-bold ${l.status === 200 ? "text-green-500 bg-green-500/10" : "text-red-500 bg-red-500/10"}`}>
                    {l.status}
                  </span>
                </td>
                <td className="py-3 px-2 text-xs text-muted-foreground font-mono">{l.latency}</td>
                <td className="py-3 px-2 text-xs text-muted-foreground text-right">{l.time}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}

function ChannelSettingsView({ onNavigate }: { onNavigate: (item: ChannelsNavItemId) => void }) {
  const [syncFreq, setSyncFreq] = useState("hourly");
  const [conflict, setConflict] = useState("sor");
  const [threshold, setThreshold] = useState("5");

  return (
    <div className="flex flex-col gap-6">
      <div className="flex items-center justify-between mt-2">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Channel Sync Settings</h1>
          <p className="text-muted-foreground mt-1">Configure global retry patterns, low-stock thresholds, and override priority.</p>
        </div>
        <Button className="rounded-full shadow-sm">Save Settings</Button>
      </div>

      <div className="bg-white/20 dark:bg-white/5 rounded-[2rem] p-6 border border-white/30 dark:border-white/10 flex flex-col gap-6">
        <div className="grid grid-cols-2 gap-6">
          <div className="flex flex-col gap-2">
            <label className="text-xs font-semibold text-foreground">Sync Frequency</label>
            <select
              value={syncFreq}
              onChange={(e) => setSyncFreq(e.target.value)}
              className="h-9 px-4 rounded-xl bg-white/40 dark:bg-black/20 border border-black/5 dark:border-white/10 text-xs focus:outline-none"
            >
              <option value="realtime">Real-time Hook-driven</option>
              <option value="hourly">Hourly Pull Interval</option>
              <option value="daily">Daily Cron Sync</option>
            </select>
            <p className="text-[10px] text-muted-foreground">Select how often catalog and pricing records are updated.</p>
          </div>

          <div className="flex flex-col gap-2">
            <label className="text-xs font-semibold text-foreground">Conflict Resolution Policy</label>
            <select
              value={conflict}
              onChange={(e) => setConflict(e.target.value)}
              className="h-9 px-4 rounded-xl bg-white/40 dark:bg-black/20 border border-black/5 dark:border-white/10 text-xs focus:outline-none"
            >
              <option value="sor">Synq System of Record Wins</option>
              <option value="channel">Channel Overwrite Wins</option>
              <option value="manual">Flag for Human Approval</option>
            </select>
            <p className="text-[10px] text-muted-foreground">Define system precedence when data changes concurrently.</p>
          </div>

          <div className="flex flex-col gap-2">
            <label className="text-xs font-semibold text-foreground">Safety Stock Buffer (Threshold)</label>
            <input
              type="number"
              value={threshold}
              onChange={(e) => setThreshold(e.target.value)}
              className="h-9 px-4 rounded-xl bg-white/40 dark:bg-black/20 border border-black/5 dark:border-white/10 text-xs focus:outline-none"
            />
            <p className="text-[10px] text-muted-foreground">Set out-of-stock buffer on marketplaces to prevent overselling.</p>
          </div>

          <div className="flex flex-col gap-2">
            <label className="text-xs font-semibold text-foreground">Max Workflow Retries</label>
            <input
              type="number"
              defaultValue="5"
              className="h-9 px-4 rounded-xl bg-white/40 dark:bg-black/20 border border-black/5 dark:border-white/10 text-xs focus:outline-none"
            />
            <p className="text-[10px] text-muted-foreground">Maximum Temporal retries before marking synchronization failed.</p>
          </div>
        </div>
      </div>
    </div>
  );
}

function AutomationRulesView() {
  const rules = [
    { id: 1, name: "Shopify Out of Stock Sync", trigger: "Stock < 3", action: "Set Shopify status = Draft", active: true },
    { id: 2, name: "Amazon Pricing Floor Override", trigger: "Competitor Price Match", action: "Adjust Price to Min ($19.99)", active: true },
    { id: 3, name: "High Value Order Routing", trigger: "Order Total > $500", action: "Require Manager Approval", active: false },
  ];
  return (
    <div className="flex flex-col gap-6">
      <div className="flex items-center justify-between mt-2">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Automation Rules</h1>
          <p className="text-muted-foreground mt-1">Define event-first triggers to automate inventory and routing actions.</p>
        </div>
        <Button className="rounded-full shadow-sm">
          <Plus className="mr-2 h-4 w-4" /> Create Rule
        </Button>
      </div>

      <div className="flex flex-col gap-4">
        {rules.map((r) => (
          <div key={r.id} className="bg-white/20 dark:bg-white/5 rounded-[2rem] p-6 border border-white/30 dark:border-white/10 flex items-center justify-between">
            <div className="flex flex-col">
              <h3 className="font-bold text-foreground">{r.name}</h3>
              <p className="text-xs text-muted-foreground mt-1">
                <span className="font-semibold text-primary">IF</span> {r.trigger}{" "}
                <span className="font-semibold text-primary">THEN</span> {r.action}
              </p>
            </div>
            <div className="flex items-center gap-3">
              <span className={`px-2 py-0.5 rounded-full text-[10px] font-bold ${r.active ? "text-green-500 bg-green-500/10" : "text-zinc-500 bg-zinc-500/10"}`}>
                {r.active ? "Active" : "Paused"}
              </span>
              <Button variant="outline" size="sm" className="rounded-full h-8 text-xs">Edit</Button>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
