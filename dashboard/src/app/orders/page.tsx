"use client";

import { useEffect, useMemo, useState, Suspense } from "react";
import { LeftSidebar } from "@/components/LeftSidebar";
import { OrderDetailsSidebar, type OrderSidebarRecord } from "@/components/OrderDetailsSidebar";
import { 
  Download, 
  Plus, 
  Search, 
  Filter, 
  MoreHorizontal, 
  ShoppingBag, 
  Clock, 
  PackageCheck, 
  Truck, 
  Settings,
  RotateCcw,
  Zap,
  Sliders,
  Sparkles,
  type LucideIcon 
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { fetchAPI } from "@/lib/api";
import { useSearchParams } from "next/navigation";

type OrdersResponse = {
  orders: OrderSidebarRecord[];
  limit: number;
  offset: number;
};

const statusStyles: Record<string, string> = {
  pending_payment: "text-amber-500 bg-amber-500/10",
  payment_authorized: "text-blue-500 bg-blue-500/10",
  confirmed: "text-green-500 bg-green-500/10",
  processing: "text-yellow-500 bg-yellow-500/10",
  fulfilled: "text-purple-500 bg-purple-500/10",
  cancelled: "text-red-500 bg-red-500/10",
  failed: "text-red-500 bg-red-500/10",
};

function OrdersPageContent() {
  const [orders, setOrders] = useState<OrderSidebarRecord[]>([]);
  const [selectedOrderID, setSelectedOrderID] = useState<string | null>(null);
  const [query, setQuery] = useState("");
  const [statusFilter, setStatusFilter] = useState("all");
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const searchParams = useSearchParams();
  const activeStatus = searchParams.get("status");
  const activeTab = searchParams.get("tab");

  useEffect(() => {
    let mounted = true;
    fetchAPI("/api/v1/oms/orders?limit=100")
      .then((data) => {
        if (!mounted) return;
        const payload = data as OrdersResponse;
        const nextOrders = payload.orders ?? [];
        setOrders(nextOrders);
        setSelectedOrderID((current) => current ?? nextOrders[0]?.id ?? null);
        setError(null);
      })
      .catch((err: Error) => {
        if (!mounted) return;
        setError(err.message || "Failed to load orders.");
      })
      .finally(() => {
        if (mounted) setLoading(false);
      });
    return () => {
      mounted = false;
    };
  }, []);

  // Sync sidebar active status filter URL param with local filter state
  useEffect(() => {
    if (activeStatus) {
      setStatusFilter(activeStatus);
    } else {
      setStatusFilter("all");
    }
  }, [activeStatus]);

  const filteredOrders = useMemo(() => {
    const needle = query.trim().toLowerCase();
    return orders.filter((order) => {
      const matchesStatus = statusFilter === "all" || order.status === statusFilter;
      const matchesQuery =
        needle === "" ||
        order.id.toLowerCase().includes(needle) ||
        (order.channel || "").toLowerCase().includes(needle) ||
        (order.source_platform || "").toLowerCase().includes(needle) ||
        (order.payment_reference || "").toLowerCase().includes(needle);
      return matchesStatus && matchesQuery;
    });
  }, [orders, query, statusFilter]);

  const selectedOrder = orders.find((order) => order.id === selectedOrderID) ?? null;
  const totalRevenue = orders.reduce((sum, order) => sum + order.total, 0);
  const pendingCount = orders.filter((order) => order.status === "pending_payment").length;
  const processingCount = orders.filter((order) => order.status === "processing" || order.status === "confirmed").length;
  const fulfilledCount = orders.filter((order) => order.status === "fulfilled" || order.status === "completed").length;
  const statuses = Array.from(new Set(orders.map((order) => order.status))).sort();

  const renderMainContent = () => {
    switch (activeTab) {
      case "settings":
        return <OrderSettingsView />;
      case "shipping-rules":
        return <ShippingRulesView />;
      case "return-rules":
        return <ReturnRulesView />;
      case "automation":
        return <OMSAutomationView />;
      default:
        return (
          <>
            <div className="flex items-center justify-between mt-2">
              <div>
                <h1 className="text-3xl font-bold tracking-tight">Orders</h1>
                <p className="text-muted-foreground mt-1">Manage, track and fulfill customer orders in real time.</p>
              </div>
              <div className="flex items-center gap-3">
                <Button variant="outline" className="bg-white/40 dark:bg-black/20 backdrop-blur-md border-white/40 dark:border-white/10 rounded-full shadow-sm hover:bg-white/60 dark:hover:bg-white/10">
                  <Download className="mr-2 h-4 w-4" />
                  Export
                </Button>
                <Button className="rounded-full shadow-sm">
                  <Plus className="mr-2 h-4 w-4" />
                  Create Order
                </Button>
                <Button variant="outline" size="icon" className="rounded-full bg-white/40 dark:bg-black/20 backdrop-blur-md border-white/40 dark:border-white/10 shadow-sm hover:bg-white/60 dark:hover:bg-white/10">
                  <MoreHorizontal className="h-4 w-4" />
                </Button>
              </div>
            </div>

            <div className="bg-white/20 dark:bg-white/5 rounded-[2rem] p-6 border border-white/30 dark:border-white/10 grid grid-cols-4 gap-4">
              <OrderKPICard title="Total Orders" value={orders.length.toLocaleString()} subtext={money(totalRevenue, "USD")} icon={ShoppingBag} color="text-blue-500" />
              <OrderKPICard title="Pending Payment" value={pendingCount.toLocaleString()} subtext="Awaiting authorization" icon={Clock} color="text-amber-500" />
              <OrderKPICard title="Processing" value={processingCount.toLocaleString()} subtext="Confirmed or active" icon={Truck} color="text-yellow-500" />
              <OrderKPICard title="Fulfilled" value={fulfilledCount.toLocaleString()} subtext="Completed flow" icon={PackageCheck} color="text-green-500" />
            </div>

            <div className="bg-white/20 dark:bg-white/5 rounded-[2rem] p-6 border border-white/30 dark:border-white/10 flex flex-col flex-1">
              <div className="flex items-center justify-between mb-6">
                <div className="relative w-64">
                  <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
                  <input
                    type="text"
                    value={query}
                    onChange={(event) => setQuery(event.target.value)}
                    placeholder="Search orders..."
                    className="w-full h-10 pl-9 pr-4 rounded-full bg-white/40 dark:bg-black/20 border border-black/5 dark:border-white/10 text-sm focus:outline-none focus:ring-2 focus:ring-primary/50"
                  />
                </div>

                <div className="flex items-center gap-3">
                  <select
                    value={statusFilter}
                    onChange={(event) => setStatusFilter(event.target.value)}
                    className="h-10 px-4 rounded-full bg-white/40 dark:bg-black/20 border border-black/5 dark:border-white/10 text-sm focus:outline-none"
                  >
                    <option value="all">All Status</option>
                    {statuses.map((status) => (
                      <option key={status} value={status}>{label(status)}</option>
                    ))}
                  </select>
                  <Button variant="outline" className="rounded-full bg-white/40 dark:bg-black/20 border-black/5 dark:border-white/10">
                    <Filter className="w-4 h-4 mr-2" /> Filters
                  </Button>
                </div>
              </div>

              <div className="w-full overflow-x-auto">
                <table className="w-full text-sm text-left">
                  <thead>
                    <tr className="text-xs text-muted-foreground uppercase tracking-wider border-b border-black/5 dark:border-white/5">
                      <th className="pb-3 font-semibold px-2">Order ID</th>
                      <th className="pb-3 font-semibold px-2">Channel</th>
                      <th className="pb-3 font-semibold px-2">Payment</th>
                      <th className="pb-3 font-semibold px-2">Order Value</th>
                      <th className="pb-3 font-semibold px-2">Status</th>
                      <th className="pb-3 font-semibold px-2">Order Date</th>
                      <th className="pb-3 font-semibold px-2 text-right">Actions</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-black/5 dark:divide-white/5">
                    {loading ? (
                      <tr>
                        <td colSpan={7} className="py-10 text-center text-muted-foreground">Loading orders...</td>
                      </tr>
                    ) : error ? (
                      <tr>
                        <td colSpan={7} className="py-10 text-center text-red-600 dark:text-red-400">{error}</td>
                      </tr>
                    ) : filteredOrders.length === 0 ? (
                      <tr>
                        <td colSpan={7} className="py-10 text-center text-muted-foreground">No orders match the current view.</td>
                      </tr>
                    ) : (
                      filteredOrders.map((order) => (
                        <tr
                          key={order.id}
                          onClick={() => setSelectedOrderID(order.id)}
                          className={`group hover:bg-black/[0.02] dark:hover:bg-white/[0.02] transition-colors cursor-pointer ${selectedOrderID === order.id ? "bg-primary/5" : ""}`}
                        >
                          <td className="py-3 px-2 font-medium">{shortOrderID(order.id)}</td>
                          <td className="py-3 px-2 capitalize">{order.channel || order.source_platform || "Direct"}</td>
                          <td className="py-3 px-2">
                            <div className="flex flex-col">
                              <span className="font-medium capitalize">{order.payment_status || "pending"}</span>
                              <span className="text-[10px] text-muted-foreground">{order.payment_provider || "No provider"}</span>
                            </div>
                          </td>
                          <td className="py-3 px-2 font-medium">{money(order.total, order.currency)}</td>
                          <td className="py-3 px-2">
                            <span className={`px-2 py-1 rounded-full text-[10px] font-bold capitalize ${statusStyles[order.status] || "text-zinc-500 bg-zinc-500/10"}`}>
                              {label(order.status)}
                            </span>
                          </td>
                          <td className="py-3 px-2">{order.created_at ? new Date(order.created_at).toLocaleString() : "No timestamp"}</td>
                          <td className="py-3 px-2 text-right">
                            <button className="text-muted-foreground hover:text-foreground" aria-label="Order actions">
                              <MoreHorizontal className="w-4 h-4" />
                            </button>
                          </td>
                        </tr>
                      ))
                    )}
                  </tbody>
                </table>
              </div>

              <div className="flex items-center justify-between mt-auto pt-6 text-xs text-muted-foreground">
                <span>Showing {filteredOrders.length} of {orders.length} orders</span>
                <span>Tenant-scoped OMS feed</span>
              </div>
            </div>
          </>
        );
    }
  };

  return (
    <>
      <LeftSidebar />

      <main className="flex-1 overflow-y-auto custom-scrollbar pr-4 flex flex-col gap-6">
        {renderMainContent()}
      </main>

      <aside className="w-[360px] flex-shrink-0 bg-[#121212] text-white p-6 rounded-3xl relative z-20 flex flex-col h-full shadow-[0_8px_30px_rgb(0,0,0,0.4)] border border-white/5">
        <OrderDetailsSidebar order={selectedOrder} onClose={() => setSelectedOrderID(null)} />
      </aside>
    </>
  );
}

export default function OrdersPage() {
  return (
    <Suspense fallback={<div className="flex-1 flex items-center justify-center text-muted-foreground text-sm">Loading Workspace...</div>}>
      <OrdersPageContent />
    </Suspense>
  );
}

function OrderKPICard({ title, value, subtext, icon: Icon, color }: { title: string; value: string; subtext: string; icon: LucideIcon; color: string }) {
  return (
    <div className="flex items-center gap-4 min-w-0">
      <div className="w-10 h-10 rounded-xl flex items-center justify-center bg-white border border-black/5 dark:border-white/5 shadow-sm flex-shrink-0">
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

function shortOrderID(id: string) {
  return id ? `#${id.slice(0, 8).toUpperCase()}` : "#ORDER";
}

function label(value: string) {
  return value.replaceAll("_", " ");
}

function money(value: number, currency: string) {
  return new Intl.NumberFormat("en-US", { style: "currency", currency: currency || "USD" }).format(value || 0);
}

// CONFIGURATION SUB-VIEWS FOR ORDERS PAGE

function OrderSettingsView() {
  const [delay, setDelay] = useState("15");
  const [fraud, setFraud] = useState("medium");

  return (
    <div className="flex flex-col gap-6">
      <div className="flex items-center justify-between mt-2">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Order Settings</h1>
          <p className="text-muted-foreground mt-1">Configure global order processing rules, fulfillment delay policies, and validation gates.</p>
        </div>
        <Button className="rounded-full shadow-sm">Save Config</Button>
      </div>

      <div className="bg-white/20 dark:bg-white/5 rounded-[2rem] p-6 border border-white/30 dark:border-white/10 flex flex-col gap-6">
        <div className="grid grid-cols-2 gap-6">
          <div className="flex flex-col gap-2">
            <label className="text-xs font-semibold text-foreground">Fulfillment Delay Buffer (Minutes)</label>
            <input
              type="number"
              value={delay}
              onChange={(e) => setDelay(e.target.value)}
              className="h-9 px-4 rounded-xl bg-white/40 dark:bg-black/20 border border-black/5 dark:border-white/10 text-xs focus:outline-none"
            />
            <p className="text-[10px] text-muted-foreground">Grace period allowing customer edits before syncing to logistics.</p>
          </div>

          <div className="flex flex-col gap-2">
            <label className="text-xs font-semibold text-foreground">Fraud Analysis Level</label>
            <select
              value={fraud}
              onChange={(e) => setFraud(e.target.value)}
              className="h-9 px-4 rounded-xl bg-white/40 dark:bg-black/20 border border-black/5 dark:border-white/10 text-xs focus:outline-none"
            >
              <option value="low">Low Security (Skip IP matches)</option>
              <option value="medium">Medium Security (Verify address + card match)</option>
              <option value="high">High Security (Flag all billing mismatches)</option>
            </select>
            <p className="text-[10px] text-muted-foreground">Level of strictness applied to inbound payment verification checks.</p>
          </div>

          <div className="flex flex-col gap-2">
            <label className="text-xs font-semibold text-foreground">Payment Authorization Mode</label>
            <select className="h-9 px-4 rounded-xl bg-white/40 dark:bg-black/20 border border-black/5 dark:border-white/10 text-xs focus:outline-none">
              <option value="auto">Auto-capture on Confirmation</option>
              <option value="manual">Authorize Only (Manual Capture)</option>
            </select>
            <p className="text-[10px] text-muted-foreground">Select when to charge customer payment cards.</p>
          </div>

          <div className="flex flex-col gap-2">
            <label className="text-xs font-semibold text-foreground">Large Order Verification Threshold</label>
            <input
              type="text"
              defaultValue="$1,000.00"
              className="h-9 px-4 rounded-xl bg-white/40 dark:bg-black/20 border border-black/5 dark:border-white/10 text-xs focus:outline-none"
            />
            <p className="text-[10px] text-muted-foreground">Hold orders above this value for explicit manager approval.</p>
          </div>
        </div>
      </div>
    </div>
  );
}

function ShippingRulesView() {
  const rules = [
    { id: 1, name: "US Domestic routing", trigger: "Country == 'US'", carrier: "UPS Ground", priority: "High" },
    { id: 2, name: "EU International delivery", trigger: "Continent == 'EU'", carrier: "DHL Express", priority: "Medium" },
    { id: 3, name: "Free shipping threshold", trigger: "Total >= $100", carrier: "USPS Priority", priority: "Low" },
  ];
  return (
    <div className="flex flex-col gap-6">
      <div className="flex items-center justify-between mt-2">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Shipping Rules</h1>
          <p className="text-muted-foreground mt-1">Establish delivery provider priorities and zone routing configurations.</p>
        </div>
        <Button className="rounded-full shadow-sm">
          <Plus className="mr-2 h-4 w-4" /> Add Route Rule
        </Button>
      </div>

      <div className="flex flex-col gap-4">
        {rules.map((r) => (
          <div key={r.id} className="bg-white/20 dark:bg-white/5 rounded-[2rem] p-6 border border-white/30 dark:border-white/10 flex items-center justify-between">
            <div className="flex flex-col">
              <h3 className="font-bold text-foreground">{r.name}</h3>
              <p className="text-xs text-muted-foreground mt-1">
                <span className="font-semibold text-primary">IF</span> {r.trigger}{" "}
                <span className="font-semibold text-primary">THEN Route via</span> {r.carrier}
              </p>
            </div>
            <div className="flex items-center gap-3">
              <span className={`px-2 py-0.5 rounded-full text-[10px] font-bold ${r.priority === "High" ? "text-purple-500 bg-purple-500/10" : "text-zinc-500 bg-zinc-500/10"}`}>
                Priority: {r.priority}
              </span>
              <Button variant="outline" size="sm" className="rounded-full h-8 text-xs">Configure</Button>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}

function ReturnRulesView() {
  return (
    <div className="flex flex-col gap-6">
      <div className="flex items-center justify-between mt-2">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Return Policies & Rules</h1>
          <p className="text-muted-foreground mt-1">Configure automated approval boundaries and restocking fees.</p>
        </div>
        <Button className="rounded-full shadow-sm">Save Policies</Button>
      </div>

      <div className="bg-white/20 dark:bg-white/5 rounded-[2rem] p-6 border border-white/30 dark:border-white/10 flex flex-col gap-6">
        <div className="grid grid-cols-2 gap-6">
          <div className="flex flex-col gap-2">
            <label className="text-xs font-semibold text-foreground">Return Window (Days)</label>
            <select className="h-9 px-4 rounded-xl bg-white/40 dark:bg-black/20 border border-black/5 dark:border-white/10 text-xs focus:outline-none">
              <option value="30">30 Days after Delivery</option>
              <option value="60">60 Days after Delivery</option>
              <option value="90">90 Days after Delivery</option>
            </select>
            <p className="text-[10px] text-muted-foreground">Maximum limit for customers to trigger a return request.</p>
          </div>

          <div className="flex flex-col gap-2">
            <label className="text-xs font-semibold text-foreground">Restocking Fee</label>
            <select className="h-9 px-4 rounded-xl bg-white/40 dark:bg-black/20 border border-black/5 dark:border-white/10 text-xs focus:outline-none">
              <option value="0">No Restocking Fee</option>
              <option value="5">Flat $5.00 Fee</option>
              <option value="percentage">10% Order Value Fee</option>
            </select>
            <p className="text-[10px] text-muted-foreground">Select penalty applied to customer refund totals.</p>
          </div>

          <div className="flex flex-col gap-2">
            <label className="text-xs font-semibold text-foreground">Auto-Approval Threshold</label>
            <select className="h-9 px-4 rounded-xl bg-white/40 dark:bg-black/20 border border-black/5 dark:border-white/10 text-xs focus:outline-none">
              <option value="never">Require Manual Review (Never Auto-Approve)</option>
              <option value="50">Auto-Approve if Order Value &lt; $50</option>
              <option value="100">Auto-Approve if Order Value &lt; $100</option>
            </select>
            <p className="text-[10px] text-muted-foreground">Automatically issue return slips below this transaction size.</p>
          </div>
        </div>
      </div>
    </div>
  );
}

function OMSAutomationView() {
  const rules = [
    { id: 1, trigger: "payment.captured", action: "Execute OrderFulfillmentWorkflow" },
    { id: 2, trigger: "fulfillment.failed", action: "Retry alternative carrier & flag customer success" },
    { id: 3, trigger: "return.requested", action: "Trigger email notification via SMTP" },
  ];
  return (
    <div className="flex flex-col gap-6">
      <div className="flex items-center justify-between mt-2">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">OMS Automation</h1>
          <p className="text-muted-foreground mt-1">Deploy event-first automation triggers for order workflows.</p>
        </div>
        <Button className="rounded-full shadow-sm">
          <Plus className="mr-2 h-4 w-4" /> Add Trigger
        </Button>
      </div>

      <div className="flex flex-col gap-4">
        {rules.map((r) => (
          <div key={r.id} className="bg-white/20 dark:bg-white/5 rounded-[2rem] p-6 border border-white/30 dark:border-white/10 flex items-center justify-between">
            <div className="flex flex-col">
              <span className="text-xs text-muted-foreground font-semibold">Trigger ID #{r.id}</span>
              <p className="text-sm font-semibold text-foreground mt-1">
                <span className="font-semibold text-primary">WHEN</span> {r.trigger}{" "}
                <span className="font-semibold text-primary">THEN</span> {r.action}
              </p>
            </div>
            <Button variant="outline" size="sm" className="rounded-full h-8 text-xs">Edit Trigger</Button>
          </div>
        ))}
      </div>
    </div>
  );
}
