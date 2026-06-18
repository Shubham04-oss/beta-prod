"use client";

import { AnalyticsLeftSidebar } from "@/components/AnalyticsLeftSidebar";
import { AnalyticsInsightsSidebar } from "@/components/AnalyticsInsightsSidebar";
import { Download, Calendar, ChevronDown, Package, ShoppingCart, TrendingUp, Truck, Clock, FileText, Info, Box, Users } from "lucide-react";
import { Area, AreaChart, Pie, PieChart, Cell, ResponsiveContainer, Tooltip, YAxis } from "recharts";

const kpiData = [
  { label: "Total Revenue", value: "$1,250,430", trend: "↑ 16.6% vs May 1 - May 5", icon: Package, color: "text-purple-500", bg: "bg-purple-500/10", stroke: "stroke-purple-500" },
  { label: "Total Orders", value: "4,782", trend: "↑ 14.3% vs May 1 - May 5", icon: ShoppingCart, color: "text-green-500", bg: "bg-green-500/10", stroke: "stroke-green-500" },
  { label: "Avg. Order Value", value: "$261.56", trend: "↑ 8.6% vs May 1 - May 5", icon: TrendingUp, color: "text-amber-500", bg: "bg-amber-500/10", stroke: "stroke-amber-500" },
  { label: "Total Shipments", value: "1,248", trend: "↑ 12.4% vs May 1 - May 5", icon: Truck, color: "text-blue-500", bg: "bg-blue-500/10", stroke: "stroke-blue-500" },
  { label: "On-time Delivery", value: "96.2%", trend: "↑ 3.2% vs May 1 - May 5", icon: Clock, color: "text-purple-500", bg: "bg-purple-500/10", stroke: "stroke-purple-500" },
];

const salesTrendData = [
  { name: "May 6", thisPeriod: 120, prevPeriod: 90 },
  { name: "May 7", thisPeriod: 150, prevPeriod: 110 },
  { name: "May 8", thisPeriod: 110, prevPeriod: 80 },
  { name: "May 9", thisPeriod: 180, prevPeriod: 140 },
  { name: "May 10", thisPeriod: 220, prevPeriod: 160 },
  { name: "May 11", thisPeriod: 190, prevPeriod: 130 },
  { name: "May 12", thisPeriod: 240, prevPeriod: 180 },
];

const ordersByStatusData = [
  { name: "Delivered", value: 3724, color: "#22c55e" },
  { name: "Processing", value: 312, color: "#3b82f6" },
  { name: "Pending", value: 64, color: "#eab308" },
  { name: "Cancelled", value: 32, color: "#ef4444" },
  { name: "Returned", value: 45, color: "#8b5cf6" },
];

const revenueByChannelData = [
  { name: "Amazon", value: 520430, color: "#22c55e" },
  { name: "Shopify", value: 320210, color: "#8b5cf6" },
  { name: "Flipkart", value: 210540, color: "#eab308" },
  { name: "Myntra", value: 110250, color: "#ef4444" },
  { name: "Others", value: 89000, color: "#a855f7" },
];

export default function AnalyticsPage() {
  return (
    <>
      <AnalyticsLeftSidebar />

      <main className="flex-1 overflow-y-auto custom-scrollbar pr-4 flex flex-col gap-6">
        
        {/* Header */}
        <div className="flex items-start justify-between mt-2">
          <div>
            <h1 className="text-3xl font-bold tracking-tight">Reports</h1>
            <p className="text-[12px] text-muted-foreground mt-1">Real-time insights and detailed reports across your operations.</p>
          </div>
          
          <div className="flex items-center gap-3">
            <div className="flex items-center gap-2 px-3 py-1.5 bg-white/40 dark:bg-black/20 border border-black/5 dark:border-white/10 rounded-lg cursor-pointer hover:bg-white/60 dark:hover:bg-white/5 transition-colors">
              <Calendar className="w-4 h-4 text-muted-foreground" />
              <span className="text-[11px] font-medium text-foreground">May 6 - May 12, 2024</span>
              <ChevronDown className="w-3.5 h-3.5 text-muted-foreground" />
            </div>
            
            <div className="flex items-center gap-2 px-3 py-1.5 bg-white/40 dark:bg-black/20 border border-black/5 dark:border-white/10 rounded-lg cursor-pointer hover:bg-white/60 dark:hover:bg-white/5 transition-colors">
              <span className="text-[11px] font-medium text-muted-foreground">Compare: <span className="text-foreground">Previous 7 days</span></span>
              <ChevronDown className="w-3.5 h-3.5 text-muted-foreground" />
            </div>

            <button className="flex items-center gap-1.5 px-4 py-1.5 bg-white/40 dark:bg-black/20 border border-black/5 dark:border-white/10 rounded-lg hover:bg-white/60 dark:hover:bg-white/5 transition-colors text-[11px] font-medium text-foreground">
              <Download className="w-3.5 h-3.5" /> Export
            </button>
          </div>
        </div>

        {/* Top KPI Cards */}
        <div className="grid grid-cols-5 gap-4">
          {kpiData.map((kpi, idx) => (
            <div key={idx} className="bg-white/40 dark:bg-black/20 rounded-2xl p-4 border border-black/5 dark:border-white/5 flex flex-col hover:bg-white/60 dark:hover:bg-black/40 transition-colors">
              <div className="flex items-center gap-3 mb-3">
                <div className={`w-8 h-8 rounded-lg flex items-center justify-center flex-shrink-0 ${kpi.bg} ${kpi.color}`}>
                  <kpi.icon className="w-4 h-4" />
                </div>
                <span className="text-[11px] font-semibold text-muted-foreground leading-tight">{kpi.label}</span>
              </div>
              <span className="text-xl font-bold text-foreground mb-1">{kpi.value}</span>
              <span className={`text-[9px] font-medium text-green-600 dark:text-green-500`}>{kpi.trend}</span>
              
              <div className="mt-3 w-full h-6 opacity-70">
                <svg viewBox="0 0 100 20" preserveAspectRatio="none" className={`w-full h-full fill-none ${kpi.stroke}`} strokeWidth="1.5">
                  <path d="M0 15 Q 15 5, 30 12 T 60 8 T 85 14 T 100 5" strokeLinecap="round" strokeLinejoin="round" />
                </svg>
              </div>
            </div>
          ))}
        </div>

        {/* Charts Grid Row 1 */}
        <div className="grid grid-cols-12 gap-6">
          
          {/* Sales Trend (2/3 width) */}
          <div className="col-span-7 bg-white/40 dark:bg-black/20 rounded-[1.5rem] p-6 border border-black/5 dark:border-white/5">
            <div className="flex items-center justify-between mb-6">
              <div>
                <h3 className="font-semibold text-[15px] flex items-center gap-2">Sales Trend <Info className="w-3.5 h-3.5 text-muted-foreground" /></h3>
                <div className="mt-2">
                  <span className="text-2xl font-bold text-foreground">$1,250,430</span>
                  <div className="flex items-center gap-1.5 mt-1">
                    <span className="text-[10px] font-medium text-green-600 dark:text-green-500">↑ 16.6%</span>
                    <span className="text-[10px] text-muted-foreground">vs May 1 - May 5</span>
                  </div>
                </div>
              </div>
              <div className="flex items-center gap-2 px-3 py-1.5 bg-black/5 dark:bg-white/5 rounded-lg cursor-pointer text-[11px] font-medium text-foreground">
                Daily <ChevronDown className="w-3.5 h-3.5 text-muted-foreground" />
              </div>
            </div>
            
            <div className="h-48 w-full mt-4">
              <ResponsiveContainer width="100%" height="100%">
                <AreaChart data={salesTrendData} margin={{ top: 10, right: 0, left: -20, bottom: 0 }}>
                  <YAxis axisLine={false} tickLine={false} tick={{ fontSize: 10, fill: '#71717a' }} tickFormatter={(val) => `$${val}K`} />
                  <defs>
                    <linearGradient id="colorThis" x1="0" y1="0" x2="0" y2="1">
                      <stop offset="5%" stopColor="#8b5cf6" stopOpacity={0.1}/>
                      <stop offset="95%" stopColor="#8b5cf6" stopOpacity={0}/>
                    </linearGradient>
                  </defs>
                  <Tooltip contentStyle={{ borderRadius: '8px', fontSize: '12px' }} />
                  <Area type="monotone" dataKey="prevPeriod" stroke="#8b5cf6" strokeWidth={1.5} strokeDasharray="4 4" fill="none" />
                  <Area type="monotone" dataKey="thisPeriod" stroke="#8b5cf6" strokeWidth={2} fillOpacity={1} fill="url(#colorThis)" />
                </AreaChart>
              </ResponsiveContainer>
            </div>
            <div className="flex items-center gap-6 mt-4 justify-start">
              <div className="flex items-center gap-2">
                <div className="w-3 h-[2px] bg-purple-500"></div>
                <span className="text-[10px] text-muted-foreground font-medium">This Period</span>
              </div>
              <div className="flex items-center gap-2">
                <div className="w-3 h-[2px] border-t border-dashed border-purple-500"></div>
                <span className="text-[10px] text-muted-foreground font-medium">Previous Period</span>
              </div>
            </div>
          </div>

          {/* Orders by Status (1/3 width) */}
          <div className="col-span-5 bg-white/40 dark:bg-black/20 rounded-[1.5rem] p-6 border border-black/5 dark:border-white/5 flex flex-col">
            <h3 className="font-semibold text-[15px] mb-6">Orders by Status</h3>
            <div className="flex items-center justify-between flex-1">
              <div className="w-36 h-36 relative">
                <ResponsiveContainer width="100%" height="100%">
                  <PieChart>
                    <Pie
                      data={ordersByStatusData}
                      cx="50%" cy="50%"
                      innerRadius={50} outerRadius={65}
                      paddingAngle={2} dataKey="value" stroke="none"
                    >
                      {ordersByStatusData.map((entry, index) => (
                        <Cell key={`cell-${index}`} fill={entry.color} />
                      ))}
                    </Pie>
                  </PieChart>
                </ResponsiveContainer>
                <div className="absolute inset-0 flex flex-col items-center justify-center pointer-events-none">
                  <span className="text-xl font-bold text-foreground">4,782</span>
                  <span className="text-[9px] text-muted-foreground mt-0.5">Total Orders</span>
                </div>
              </div>

              <div className="flex flex-col gap-3 flex-1 pl-6">
                {ordersByStatusData.map((item, i) => (
                  <div key={i} className="flex items-center justify-between text-[11px]">
                    <div className="flex items-center gap-2">
                      <div className="w-2 h-2 rounded-sm" style={{ backgroundColor: item.color }}></div>
                      <span className="text-foreground font-medium">{item.name}</span>
                    </div>
                    <div className="flex items-center gap-2">
                      <span className="text-foreground font-medium">{item.value.toLocaleString("en-US")}</span>
                      <span className="text-muted-foreground w-10 text-right">({((item.value / 4782) * 100).toFixed(1)}%)</span>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          </div>

        </div>

        {/* Charts Grid Row 2 */}
        <div className="grid grid-cols-2 gap-6">
          
          {/* Revenue by Channel */}
          <div className="bg-white/40 dark:bg-black/20 rounded-[1.5rem] p-6 border border-black/5 dark:border-white/5">
            <h3 className="font-semibold text-[15px] mb-6">Revenue by Channel</h3>
            <div className="flex items-center justify-between">
              <div className="w-32 h-32 relative">
                <ResponsiveContainer width="100%" height="100%">
                  <PieChart>
                    <Pie
                      data={revenueByChannelData}
                      cx="50%" cy="50%"
                      innerRadius={45} outerRadius={60}
                      paddingAngle={2} dataKey="value" stroke="none"
                    >
                      {revenueByChannelData.map((entry, index) => (
                        <Cell key={`cell-${index}`} fill={entry.color} />
                      ))}
                    </Pie>
                  </PieChart>
                </ResponsiveContainer>
                <div className="absolute inset-0 flex flex-col items-center justify-center pointer-events-none">
                  <span className="text-[13px] font-bold text-foreground">$1.25M</span>
                  <span className="text-[8px] text-muted-foreground mt-0.5">Total Revenue</span>
                </div>
              </div>

              <div className="flex flex-col gap-2.5 flex-1 pl-8">
                {revenueByChannelData.map((item, i) => (
                  <div key={i} className="flex items-center justify-between text-[11px]">
                    <div className="flex items-center gap-2">
                      <div className="w-2 h-2 rounded-sm" style={{ backgroundColor: item.color }}></div>
                      <span className="text-foreground font-medium">{item.name}</span>
                    </div>
                    <div className="flex items-center gap-2">
                      <span className="text-foreground font-medium">${(item.value).toLocaleString("en-US")}</span>
                      <span className="text-muted-foreground w-10 text-right">({((item.value / 1250430) * 100).toFixed(1)}%)</span>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          </div>

          {/* Shipments Performance */}
          <div className="bg-white/40 dark:bg-black/20 rounded-[1.5rem] p-6 border border-black/5 dark:border-white/5 flex flex-col">
            <h3 className="font-semibold text-[15px] mb-4 flex items-center gap-2">Shipments Performance <Info className="w-3.5 h-3.5 text-muted-foreground" /></h3>
            
            <div className="grid grid-cols-2 gap-4 mb-4">
              <div>
                <span className="text-[10px] text-muted-foreground font-medium">On-time Delivery</span>
                <div className="flex items-end gap-2 mt-1">
                  <span className="text-2xl font-bold text-foreground">96.2%</span>
                </div>
                <span className="text-[9px] font-medium text-green-600 dark:text-green-500 mt-1 block">↑ 3.2% vs May 1 - May 5</span>
              </div>
              <div>
                <span className="text-[10px] text-muted-foreground font-medium">Avg. Delivery Time</span>
                <div className="flex items-end gap-2 mt-1">
                  <span className="text-2xl font-bold text-foreground">2.6 days</span>
                </div>
                <span className="text-[9px] font-medium text-red-500 mt-1 block">↓ 0.4 days vs May 1 - May 5</span>
              </div>
            </div>

            <div className="h-16 w-full mt-auto">
              <ResponsiveContainer width="100%" height="100%">
                <AreaChart data={salesTrendData} margin={{ top: 5, right: 0, left: 0, bottom: 0 }}>
                  <YAxis domain={['dataMin - 10', 'dataMax + 10']} hide />
                  <Area type="monotone" dataKey="thisPeriod" stroke="#8b5cf6" strokeWidth={2} fill="none" />
                </AreaChart>
              </ResponsiveContainer>
            </div>
          </div>

        </div>

        {/* Reports Library */}
        <div className="mt-2 mb-6">
          <div className="flex items-center justify-between mb-4">
            <h3 className="font-semibold text-[15px]">Reports Library</h3>
            <button className="text-[11px] font-semibold text-primary hover:text-primary/80 transition-colors">
              View all
            </button>
          </div>
          
          <div className="grid grid-cols-5 gap-4">
            {[
              { name: "Sales Performance Report", desc: "Track revenue, orders and AOV", freq: "Daily", icon: FileText, color: "text-purple-500", bg: "bg-purple-500/10" },
              { name: "Order Analytics Report", desc: "Detailed analysis of orders", freq: "Daily", icon: ShoppingCart, color: "text-green-500", bg: "bg-green-500/10" },
              { name: "Shipment Analytics Report", desc: "Shipment performance and ETA", freq: "Daily", icon: Truck, color: "text-blue-500", bg: "bg-blue-500/10" },
              { name: "Inventory Report", desc: "Stock levels and inventory turns", freq: "Daily", icon: Box, color: "text-amber-500", bg: "bg-amber-500/10" },
              { name: "Customer Report", desc: "Customer insights and trends", freq: "Weekly", icon: Users, color: "text-purple-500", bg: "bg-purple-500/10" },
            ].map((report, i) => (
              <div key={i} className="bg-white/40 dark:bg-black/20 rounded-2xl p-4 border border-black/5 dark:border-white/5 hover:bg-white/60 dark:hover:bg-white/5 transition-colors cursor-pointer group">
                <div className={`w-8 h-8 rounded-lg flex items-center justify-center flex-shrink-0 ${report.bg} ${report.color} mb-3`}>
                  <report.icon className="w-4 h-4" />
                </div>
                <h4 className="text-[11px] font-semibold text-foreground group-hover:text-primary transition-colors line-clamp-1">{report.name}</h4>
                <p className="text-[9px] text-muted-foreground mt-1 line-clamp-2 min-h-[1.5rem]">{report.desc}</p>
                <div className="mt-4 text-[9px] font-medium text-muted-foreground">
                  {report.freq}
                </div>
              </div>
            ))}
          </div>
        </div>

      </main>

      {/* Right Matte Black Area */}
      <aside className="w-[360px] flex-shrink-0 bg-[#121212] text-white p-6 rounded-3xl relative z-20 flex flex-col h-full shadow-[0_8px_30px_rgb(0,0,0,0.4)] border border-white/5">
        <AnalyticsInsightsSidebar />
      </aside>
    </>
  );
}
