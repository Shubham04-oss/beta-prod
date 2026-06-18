"use client";

import { LeftSidebar } from "@/components/LeftSidebar";
import { ShippingInsightsSidebar } from "@/components/ShippingInsightsSidebar";
import { PackageOpen, Truck, ArrowRight, CheckCircle2, AlertTriangle, Search, Filter, Columns, Download, Eye, MoreHorizontal, ChevronLeft, ChevronRight } from "lucide-react";

const kpiData = [
  { label: "Ready to Ship", value: "128", trend: "↑ 18.6% vs yesterday", icon: PackageOpen, color: "text-blue-500", bg: "bg-blue-500/10", stroke: "stroke-blue-500" },
  { label: "In Transit", value: "482", trend: "↑ 12.4% vs yesterday", icon: Truck, color: "text-green-500", bg: "bg-green-500/10", stroke: "stroke-green-500" },
  { label: "Out for Delivery", value: "156", trend: "↑ 7.8% vs yesterday", icon: ArrowRight, color: "text-amber-500", bg: "bg-amber-500/10", stroke: "stroke-amber-500" },
  { label: "Delivered", value: "1,304", trend: "↑ 15.3% vs yesterday", icon: CheckCircle2, color: "text-purple-500", bg: "bg-purple-500/10", stroke: "stroke-purple-500" },
  { label: "Exceptions", value: "23", trend: "↓ 8.1% vs yesterday", icon: AlertTriangle, color: "text-red-500", bg: "bg-red-500/10", stroke: "stroke-red-500" },
];

const shipmentsData = [
  { id: "SHP-78291", order: "#ORD-78291", customer: "Emma Johnson", email: "emma.j@example.com", carrier: "FedEx", service: "FedEx Express", status: "In Transit", tracking: "FDX1234567890", date: "May 9, 2024", time: "by 8:00 PM", carrierLogo: "https://cdn.jsdelivr.net/gh/devicons/devicon@latest/icons/fedex/fedex-original.svg" },
  { id: "SHP-78290", order: "#ORD-78290", customer: "James Carter", email: "james.c@example.com", carrier: "UPS", service: "UPS Ground", status: "Out for Delivery", tracking: "1Z999AA1234567890", date: "May 8, 2024", time: "by 7:00 PM", carrierLogo: "https://cdn.jsdelivr.net/gh/devicons/devicon@latest/icons/ups/ups-original.svg" },
  { id: "SHP-78289", order: "#ORD-78289", customer: "Sophia Martinez", email: "sophia.m@example.com", carrier: "DHL", service: "DHL Express", status: "Delivered", tracking: "DHL1234567890", date: "May 7, 2024", time: "2:15 PM", carrierLogo: "https://cdn.jsdelivr.net/gh/devicons/devicon@latest/icons/dhl/dhl-original.svg" },
  { id: "SHP-78288", order: "#ORD-78288", customer: "Liam Anderson", email: "liam.a@example.com", carrier: "USPS", service: "USPS Priority", status: "In Transit", tracking: "940551020082962", date: "May 10, 2024", time: "by 8:00 PM", carrierLogo: "https://cdn.jsdelivr.net/gh/devicons/devicon@latest/icons/usps/usps-original.svg" },
  { id: "SHP-78287", order: "#ORD-78287", customer: "Olivia Brown", email: "olivia.b@example.com", carrier: "FedEx", service: "FedEx Ground", status: "Exception", tracking: "FDX9876543210", date: "May 8, 2024", time: "Delayed", carrierLogo: "https://cdn.jsdelivr.net/gh/devicons/devicon@latest/icons/fedex/fedex-original.svg" },
  { id: "SHP-78286", order: "#ORD-78286", customer: "Noah Wilson", email: "noah.w@example.com", carrier: "UPS", service: "UPS 2nd Day Air", status: "Ready to Ship", tracking: "-", date: "May 9, 2024", time: "by 10:30 AM", carrierLogo: "https://cdn.jsdelivr.net/gh/devicons/devicon@latest/icons/ups/ups-original.svg" },
  { id: "SHP-78285", order: "#ORD-78285", customer: "Ava Taylor", email: "ava.t@example.com", carrier: "DHL", service: "DHL Express", status: "Ready to Ship", tracking: "-", date: "May 9, 2024", time: "by 12:00 PM", carrierLogo: "https://cdn.jsdelivr.net/gh/devicons/devicon@latest/icons/dhl/dhl-original.svg" },
  { id: "SHP-78284", order: "#ORD-78284", customer: "Ethan Thomas", email: "ethan.t@example.com", carrier: "USPS", service: "USPS Priority", status: "In Transit", tracking: "940551020082963", date: "May 11, 2024", time: "by 8:00 PM", carrierLogo: "https://cdn.jsdelivr.net/gh/devicons/devicon@latest/icons/usps/usps-original.svg" },
];

export default function ShippingPickupPage() {
  return (
    <>
      {/* Left Sidebar uses the generic Orders one */}
      <LeftSidebar />

      {/* Main Content Area */}
      <main className="flex-1 overflow-y-auto custom-scrollbar pr-4 flex flex-col gap-6">
        
        {/* Header */}
        <div className="mt-2">
          <h1 className="text-2xl font-bold tracking-tight">Shipping & Pickup</h1>
          <p className="text-[12px] text-muted-foreground mt-0.5">Manage shipments, tracking and pickup operations in one place.</p>
        </div>

        {/* 5 KPI Cards */}
        <div className="grid grid-cols-5 gap-4">
          {kpiData.map((kpi, idx) => (
            <div key={idx} className="bg-white/40 dark:bg-black/20 rounded-2xl p-4 border border-black/5 dark:border-white/5 flex flex-col hover:bg-white/60 dark:hover:bg-black/40 transition-colors">
              <div className="flex items-center gap-3 mb-4">
                <div className={`w-8 h-8 rounded-full flex items-center justify-center flex-shrink-0 ${kpi.bg} ${kpi.color}`}>
                  <kpi.icon className="w-4 h-4" />
                </div>
                <span className="text-[11px] font-semibold text-muted-foreground">{kpi.label}</span>
              </div>
              <span className="text-2xl font-bold text-foreground mb-1">{kpi.value}</span>
              <span className={`text-[9px] font-medium ${kpi.color}`}>{kpi.trend}</span>
              
              <div className="mt-3 w-full h-6 opacity-70">
                <svg viewBox="0 0 100 20" preserveAspectRatio="none" className={`w-full h-full fill-none ${kpi.stroke}`} strokeWidth="1.5">
                  <path d="M0 15 Q 15 5, 30 12 T 60 8 T 85 14 T 100 5" strokeLinecap="round" strokeLinejoin="round" />
                </svg>
              </div>
            </div>
          ))}
        </div>

        {/* Table Controls */}
        <div className="flex flex-col gap-4 mt-2">
          <div className="flex items-center justify-between border-b border-black/5 dark:border-white/5 pb-0">
            <div className="flex items-center gap-6 text-[12px] font-medium text-muted-foreground">
              <button className="pb-3 text-foreground font-semibold border-b-2 border-primary -mb-[2px]">All Shipments</button>
              <button className="pb-3 hover:text-foreground transition-colors -mb-[2px]">Ready to Ship</button>
              <button className="pb-3 hover:text-foreground transition-colors -mb-[2px]">In Transit</button>
              <button className="pb-3 hover:text-foreground transition-colors -mb-[2px]">Out for Delivery</button>
              <button className="pb-3 hover:text-foreground transition-colors -mb-[2px]">Delivered</button>
              <button className="pb-3 hover:text-foreground transition-colors -mb-[2px]">Exceptions</button>
              <button className="pb-3 hover:text-foreground transition-colors -mb-[2px]">Pickups</button>
            </div>
            
            <div className="flex items-center gap-2 pb-2">
              <button className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg bg-white/40 dark:bg-black/20 hover:bg-white/60 dark:hover:bg-white/5 border border-black/5 dark:border-white/10 text-[11px] font-medium transition-colors">
                <Filter className="w-3.5 h-3.5" /> Filters
              </button>
              <button className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg bg-white/40 dark:bg-black/20 hover:bg-white/60 dark:hover:bg-white/5 border border-black/5 dark:border-white/10 text-[11px] font-medium transition-colors">
                <Columns className="w-3.5 h-3.5" /> Columns
              </button>
              <button className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg bg-white/40 dark:bg-black/20 hover:bg-white/60 dark:hover:bg-white/5 border border-black/5 dark:border-white/10 text-[11px] font-medium transition-colors">
                <Download className="w-3.5 h-3.5" /> Export
              </button>
            </div>
          </div>

          {/* Shipments Table */}
          <div className="bg-white/40 dark:bg-black/20 rounded-[1.5rem] border border-black/5 dark:border-white/5 overflow-hidden">
            <table className="w-full text-left border-collapse">
              <thead>
                <tr className="border-b border-black/5 dark:border-white/5 text-[10px] font-semibold text-muted-foreground">
                  <th className="py-3 px-4 w-10 text-center"><input type="checkbox" className="rounded border-black/20 bg-transparent" /></th>
                  <th className="py-3 px-4">Shipment ID</th>
                  <th className="py-3 px-4">Order ID</th>
                  <th className="py-3 px-4">Customer</th>
                  <th className="py-3 px-4">Carrier</th>
                  <th className="py-3 px-4">Service</th>
                  <th className="py-3 px-4">Status</th>
                  <th className="py-3 px-4">Tracking ID</th>
                  <th className="py-3 px-4">Est. Delivery</th>
                  <th className="py-3 px-4 text-right">Actions</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-black/5 dark:divide-white/5">
                {shipmentsData.map((ship, idx) => (
                  <tr key={idx} className="group hover:bg-black/[0.02] dark:hover:bg-white/[0.02] transition-colors">
                    <td className="py-3 px-4 text-center">
                      <input type="checkbox" className="rounded border-black/10 dark:border-white/10 bg-transparent opacity-50 group-hover:opacity-100" />
                    </td>
                    <td className="py-3 px-4 text-[11px] font-semibold text-foreground">{ship.id}</td>
                    <td className="py-3 px-4 text-[11px] font-medium text-muted-foreground">{ship.order}</td>
                    <td className="py-3 px-4">
                      <div className="flex flex-col">
                        <span className="text-[11px] font-semibold text-foreground">{ship.customer}</span>
                        <span className="text-[9px] text-muted-foreground">{ship.email}</span>
                      </div>
                    </td>
                    <td className="py-3 px-4">
                      <div className="flex items-center gap-2">
                        <div className="w-5 h-5 rounded bg-white p-0.5 border border-black/5 flex-shrink-0">
                          <img src={ship.carrierLogo} alt={ship.carrier} className="w-full h-full object-contain" />
                        </div>
                        <span className="text-[11px] font-medium text-foreground">{ship.carrier}</span>
                      </div>
                    </td>
                    <td className="py-3 px-4 text-[11px] text-muted-foreground">{ship.service}</td>
                    <td className="py-3 px-4">
                      <span className={`inline-flex items-center px-2 py-0.5 rounded text-[9px] font-bold ${
                        ship.status === "In Transit" ? "bg-blue-500/10 text-blue-500" :
                        ship.status === "Out for Delivery" ? "bg-amber-500/10 text-amber-500" :
                        ship.status === "Delivered" ? "bg-green-500/10 text-green-500" :
                        ship.status === "Exception" ? "bg-red-500/10 text-red-500" :
                        "bg-purple-500/10 text-purple-500"
                      }`}>
                        {ship.status}
                      </span>
                    </td>
                    <td className="py-3 px-4 text-[11px] font-medium text-foreground">{ship.tracking}</td>
                    <td className="py-3 px-4">
                      <div className="flex flex-col">
                        <span className="text-[11px] font-medium text-foreground">{ship.date}</span>
                        <span className="text-[9px] text-muted-foreground">{ship.time}</span>
                      </div>
                    </td>
                    <td className="py-3 px-4 text-right">
                      <div className="flex items-center justify-end gap-2 opacity-50 group-hover:opacity-100 transition-opacity">
                        <button className="text-muted-foreground hover:text-foreground p-1">
                          <Eye className="w-3.5 h-3.5" />
                        </button>
                        <button className="text-muted-foreground hover:text-foreground p-1">
                          <MoreHorizontal className="w-4 h-4" />
                        </button>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
            
            <div className="flex items-center justify-between px-6 py-3 border-t border-black/5 dark:border-white/5 bg-black/[0.01] dark:bg-white/[0.01]">
              <span className="text-[10px] text-muted-foreground">Showing 1 to 8 of 1,842 shipments</span>
              <div className="flex items-center gap-1">
                <button className="w-6 h-6 rounded flex items-center justify-center text-muted-foreground hover:bg-black/5 dark:hover:bg-white/5 disabled:opacity-50">
                  <ChevronLeft className="w-3.5 h-3.5" />
                </button>
                <button className="w-6 h-6 rounded flex items-center justify-center bg-foreground text-background text-[10px] font-medium">1</button>
                <button className="w-6 h-6 rounded flex items-center justify-center text-muted-foreground hover:bg-black/5 dark:hover:bg-white/5 text-[10px] font-medium">2</button>
                <button className="w-6 h-6 rounded flex items-center justify-center text-muted-foreground hover:bg-black/5 dark:hover:bg-white/5 text-[10px] font-medium">3</button>
                <button className="w-6 h-6 rounded flex items-center justify-center text-muted-foreground hover:bg-black/5 dark:hover:bg-white/5 text-[10px] font-medium">4</button>
                <button className="w-6 h-6 rounded flex items-center justify-center text-muted-foreground hover:bg-black/5 dark:hover:bg-white/5 text-[10px] font-medium">5</button>
                <span className="text-muted-foreground px-1 text-[10px]">...</span>
                <button className="w-6 h-6 rounded flex items-center justify-center text-muted-foreground hover:bg-black/5 dark:hover:bg-white/5 text-[10px] font-medium">231</button>
                <button className="w-6 h-6 rounded flex items-center justify-center text-muted-foreground hover:bg-black/5 dark:hover:bg-white/5">
                  <ChevronRight className="w-3.5 h-3.5" />
                </button>
              </div>
            </div>
          </div>
        </div>

      </main>

      {/* Right Matte Black Area */}
      <aside className="w-[360px] flex-shrink-0 bg-[#121212] text-white p-6 rounded-3xl relative z-20 flex flex-col h-full shadow-[0_8px_30px_rgb(0,0,0,0.4)] border border-white/5">
        <ShippingInsightsSidebar />
      </aside>
    </>
  );
}
