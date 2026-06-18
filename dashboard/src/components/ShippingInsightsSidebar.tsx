import { ChevronDown, Plus, Calendar, Printer, Settings2, MoreHorizontal } from "lucide-react";
import { PieChart, Pie, Cell, ResponsiveContainer } from "recharts";
import { Avatar, AvatarFallback, AvatarImage } from "./ui/avatar";

const overviewData = [
  { name: "Delivered", value: 1304, color: "#22c55e" },
  { name: "In Transit", value: 482, color: "#3b82f6" },
  { name: "Out for Delivery", value: 156, color: "#eab308" },
  { name: "Exceptions", value: 23, color: "#ef4444" },
];

export function ShippingInsightsSidebar() {
  return (
    <div className="flex flex-col h-full w-full">
      <div className="flex-1 overflow-y-auto pr-2 custom-scrollbar flex flex-col gap-8 pb-10">
        
        {/* Shipment Overview */}
        <div>
          <div className="flex items-center justify-between mb-4">
            <h4 className="font-semibold text-[13px] text-zinc-100">Shipment Overview</h4>
            <div className="flex items-center gap-1 text-[10px] text-zinc-400 hover:text-white cursor-pointer bg-white/5 px-2 py-1 rounded-md">
              This Week <ChevronDown className="w-3 h-3" />
            </div>
          </div>
          
          <div className="flex items-center gap-6 mt-6">
            <div className="w-24 h-24 relative flex-shrink-0">
              <ResponsiveContainer width="100%" height="100%">
                <PieChart>
                  <Pie
                    data={overviewData}
                    cx="50%"
                    cy="50%"
                    innerRadius={35}
                    outerRadius={45}
                    paddingAngle={2}
                    dataKey="value"
                    stroke="none"
                  >
                    {overviewData.map((entry, index) => (
                      <Cell key={`cell-${index}`} fill={entry.color} />
                    ))}
                  </Pie>
                </PieChart>
              </ResponsiveContainer>
              <div className="absolute inset-0 flex flex-col items-center justify-center pointer-events-none">
                <span className="text-[13px] font-bold text-white">2,093</span>
                <span className="text-[8px] text-zinc-400">Total Shipments</span>
              </div>
            </div>

            <div className="flex flex-col gap-2.5 flex-1">
              {overviewData.map((item, i) => (
                <div key={i} className="flex items-center justify-between text-[10px]">
                  <div className="flex items-center gap-2">
                    <div className="w-1.5 h-1.5 rounded-full" style={{ backgroundColor: item.color }}></div>
                    <span className="text-zinc-300">{item.name}</span>
                  </div>
                  <div className="flex items-center gap-2">
                    <span className="text-white font-medium">{item.value.toLocaleString("en-US")}</span>
                    <span className="text-zinc-500 w-8 text-right">({((item.value / 2093) * 100).toFixed(1)}%)</span>
                  </div>
                </div>
              ))}
            </div>
          </div>
        </div>

        {/* Carrier Performance */}
        <div>
          <div className="flex items-center justify-between mb-4">
            <h4 className="font-semibold text-[13px] text-zinc-100">Carrier Performance</h4>
            <div className="flex items-center gap-1 text-[10px] text-zinc-400 hover:text-white cursor-pointer bg-white/5 px-2 py-1 rounded-md">
              This Week <ChevronDown className="w-3 h-3" />
            </div>
          </div>
          
          <div className="flex flex-col gap-4 mt-2">
            {[
              { name: "FedEx", icon: "https://cdn.jsdelivr.net/gh/devicons/devicon@latest/icons/fedex/fedex-original.svg", score: "98.6%", fill: "98%" },
              { name: "UPS", icon: "https://cdn.jsdelivr.net/gh/devicons/devicon@latest/icons/ups/ups-original.svg", score: "97.2%", fill: "97%" },
              { name: "DHL", icon: "https://cdn.jsdelivr.net/gh/devicons/devicon@latest/icons/dhl/dhl-original.svg", score: "95.4%", fill: "95%" },
              { name: "USPS", icon: "https://cdn.jsdelivr.net/gh/devicons/devicon@latest/icons/usps/usps-original.svg", score: "93.1%", fill: "93%" },
            ].map((carrier, i) => (
              <div key={i} className="flex items-center gap-4 group">
                <div className="w-6 h-6 rounded bg-white p-0.5 flex-shrink-0">
                  <img src={carrier.icon} alt={carrier.name} className="w-full h-full object-contain grayscale group-hover:grayscale-0 transition-all" />
                </div>
                <span className="text-[11px] font-medium text-white w-8">{carrier.score}</span>
                <div className="flex-1 h-1.5 bg-white/10 rounded-full overflow-hidden">
                  <div className="h-full bg-purple-500 rounded-full" style={{ width: carrier.fill }}></div>
                </div>
              </div>
            ))}
          </div>
          <button className="text-[10px] font-medium text-zinc-400 hover:text-white transition-colors flex items-center justify-center w-full gap-1 mt-5 bg-white/5 py-2 rounded-xl">
            View full performance <ChevronDown className="w-3 h-3 -rotate-90" />
          </button>
        </div>

        {/* Upcoming Pickups */}
        <div>
          <div className="flex items-center justify-between mb-4">
            <h4 className="font-semibold text-[13px] text-zinc-100">Upcoming Pickups</h4>
            <span className="text-[10px] text-zinc-400 hover:text-white cursor-pointer transition-colors">View all</span>
          </div>

          <div className="flex flex-col gap-3">
            {[
              { name: "FedEx Express Pickup", time: "Today, 3:00 PM", loc: "Warehouse A • Dock 3", icon: "https://cdn.jsdelivr.net/gh/devicons/devicon@latest/icons/fedex/fedex-original.svg" },
              { name: "UPS Ground Pickup", time: "Tomorrow, 10:00 AM", loc: "Warehouse B • Dock 1", icon: "https://cdn.jsdelivr.net/gh/devicons/devicon@latest/icons/ups/ups-original.svg" },
              { name: "DHL Express Pickup", time: "May 9, 2024 • 9:00 AM", loc: "Warehouse A • Dock 2", icon: "https://cdn.jsdelivr.net/gh/devicons/devicon@latest/icons/dhl/dhl-original.svg" },
            ].map((pickup, i) => (
              <div key={i} className="flex items-start gap-3 p-3 rounded-xl bg-white/5 border border-white/5 hover:bg-white/10 transition-colors cursor-pointer group">
                <div className="w-8 h-8 rounded-lg bg-white p-1.5 flex-shrink-0 mt-0.5">
                  <img src={pickup.icon} alt={pickup.name} className="w-full h-full object-contain grayscale group-hover:grayscale-0 transition-all" />
                </div>
                <div className="flex flex-col flex-1">
                  <span className="text-[11px] font-semibold text-white">{pickup.name}</span>
                  <span className="text-[10px] text-zinc-400 mt-0.5">{pickup.time}</span>
                  <span className="text-[9px] text-zinc-500 mt-0.5">{pickup.loc}</span>
                </div>
                <span className="text-[9px] font-semibold text-purple-400 bg-purple-500/10 px-2 py-1 rounded">Scheduled</span>
              </div>
            ))}
          </div>
        </div>

        {/* Quick Actions */}
        <div>
          <h4 className="font-semibold text-[13px] text-zinc-100 mb-4">Quick Actions</h4>
          <div className="grid grid-cols-2 gap-2">
            <button className="flex items-center gap-2 p-2.5 rounded-lg bg-white/5 border border-white/5 hover:bg-white/10 transition-colors text-[10px] text-zinc-300 font-medium justify-center">
              <Plus className="w-3.5 h-3.5" />
              Create Shipment
            </button>
            <button className="flex items-center gap-2 p-2.5 rounded-lg bg-white/5 border border-white/5 hover:bg-white/10 transition-colors text-[10px] text-zinc-300 font-medium justify-center">
              <Calendar className="w-3.5 h-3.5" />
              Schedule Pickup
            </button>
            <button className="flex items-center gap-2 p-2.5 rounded-lg bg-white/5 border border-white/5 hover:bg-white/10 transition-colors text-[10px] text-zinc-300 font-medium justify-center">
              <Printer className="w-3.5 h-3.5" />
              Print Labels
            </button>
            <button className="flex items-center gap-2 p-2.5 rounded-lg bg-white/5 border border-white/5 hover:bg-white/10 transition-colors text-[10px] text-zinc-300 font-medium justify-center">
              <Settings2 className="w-3.5 h-3.5" />
              Manage Carriers
            </button>
          </div>
        </div>

      </div>
    </div>
  );
}
