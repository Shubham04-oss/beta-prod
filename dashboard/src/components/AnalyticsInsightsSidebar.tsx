import { ChevronDown, ChevronRight, TrendingUp, Truck, Clock, AlertTriangle, FileText, Plus, Calendar, Settings2, Download } from "lucide-react";

export function AnalyticsInsightsSidebar() {
  return (
    <div className="flex flex-col h-full w-full">
      <div className="flex-1 overflow-y-auto pr-2 custom-scrollbar flex flex-col gap-8 pb-10">
        
        {/* Insights */}
        <div>
          <div className="flex items-center justify-between mb-4">
            <h4 className="font-semibold text-[13px] text-zinc-100 flex items-center gap-2">
              <TrendingUp className="w-4 h-4 text-purple-500" /> Insights
            </h4>
            <div className="flex items-center gap-1 text-[10px] text-zinc-400 hover:text-white cursor-pointer bg-white/5 px-2 py-1 rounded-md">
              This Week <ChevronDown className="w-3 h-3" />
            </div>
          </div>

          <div className="flex flex-col gap-3">
            {[
              { title: "Revenue is up by 16.6%", sub: "Compared to May 1 – May 5", icon: TrendingUp, color: "text-green-500", bg: "bg-green-500/10" },
              { title: "Shipments are up by 12.4%", sub: "Compared to May 1 – May 5", icon: Truck, color: "text-blue-500", bg: "bg-blue-500/10" },
              { title: "On-time delivery improved", sub: "By 3.2% compared to last week", icon: Clock, color: "text-purple-500", bg: "bg-purple-500/10" },
              { title: "Return rate increased", sub: "By 0.6% compared to last week", icon: AlertTriangle, color: "text-amber-500", bg: "bg-amber-500/10" },
            ].map((insight, i) => (
              <div key={i} className="flex items-center gap-3 p-3 rounded-xl bg-white/5 border border-white/5 hover:bg-white/10 transition-colors cursor-pointer group">
                <div className={`w-8 h-8 rounded-lg flex items-center justify-center flex-shrink-0 ${insight.bg} ${insight.color}`}>
                  <insight.icon className="w-4 h-4" />
                </div>
                <div className="flex flex-col flex-1">
                  <span className="text-[11px] font-medium text-zinc-200 group-hover:text-white transition-colors">{insight.title}</span>
                  <span className="text-[9px] text-zinc-500 mt-0.5">{insight.sub}</span>
                </div>
                <ChevronRight className="w-3.5 h-3.5 text-zinc-600 group-hover:text-zinc-400 transition-colors" />
              </div>
            ))}
          </div>
        </div>

        {/* Top Reports */}
        <div>
          <div className="flex items-center justify-between mb-4">
            <h4 className="font-semibold text-[13px] text-zinc-100">Top Reports</h4>
            <span className="text-[10px] text-purple-400 hover:text-purple-300 font-medium cursor-pointer transition-colors">View all</span>
          </div>

          <div className="flex flex-col gap-3">
            {[
              { name: "Sales Performance Report", val: "$1,250,430", trend: "↑ 16.6%", trendColor: "text-green-500", icon: FileText, color: "text-blue-500", bg: "bg-blue-500/10" },
              { name: "Order Analytics Report", val: "4,782", trend: "↑ 14.3%", trendColor: "text-green-500", icon: FileText, color: "text-green-500", bg: "bg-green-500/10" },
              { name: "Shipment Analytics Report", val: "1,248", trend: "↑ 12.4%", trendColor: "text-green-500", icon: FileText, color: "text-blue-500", bg: "bg-blue-500/10" },
              { name: "Inventory Report", val: "2.4%", trend: "↓ 0.6%", trendColor: "text-red-500", icon: FileText, color: "text-amber-500", bg: "bg-amber-500/10" },
              { name: "Customer Report", val: "1,856", trend: "↑ 8.7%", trendColor: "text-green-500", icon: FileText, color: "text-purple-500", bg: "bg-purple-500/10" },
            ].map((report, i) => (
              <div key={i} className="flex items-center gap-3 group cursor-pointer border-b border-white/5 pb-3 last:border-0 last:pb-0">
                <div className={`w-7 h-7 rounded-lg flex items-center justify-center flex-shrink-0 ${report.bg} ${report.color}`}>
                  <report.icon className="w-3.5 h-3.5" />
                </div>
                <span className="text-[11px] font-medium text-zinc-300 group-hover:text-white transition-colors flex-1 truncate pr-2">{report.name}</span>
                <div className="flex flex-col items-end">
                  <span className="text-[10px] font-semibold text-white">{report.val}</span>
                  <span className={`text-[8px] font-medium ${report.trendColor}`}>{report.trend}</span>
                </div>
              </div>
            ))}
          </div>
        </div>

        {/* Quick Actions */}
        <div>
          <h4 className="font-semibold text-[13px] text-zinc-100 mb-4">Quick Actions</h4>
          <div className="grid grid-cols-2 gap-2">
            <button className="flex items-center gap-2 p-3 rounded-lg bg-white/5 border border-white/5 hover:bg-white/10 transition-colors text-[10px] text-zinc-300 font-medium justify-center text-center">
              <Plus className="w-3.5 h-3.5 flex-shrink-0" />
              <span className="truncate">Create Custom Report</span>
            </button>
            <button className="flex items-center gap-2 p-3 rounded-lg bg-white/5 border border-white/5 hover:bg-white/10 transition-colors text-[10px] text-zinc-300 font-medium justify-center text-center">
              <Calendar className="w-3.5 h-3.5 flex-shrink-0" />
              <span className="truncate">Schedule Report</span>
            </button>
            <button className="flex items-center gap-2 p-3 rounded-lg bg-white/5 border border-white/5 hover:bg-white/10 transition-colors text-[10px] text-zinc-300 font-medium justify-center text-center">
              <FileText className="w-3.5 h-3.5 flex-shrink-0" />
              <span className="truncate">Manage Reports</span>
            </button>
            <button className="flex items-center gap-2 p-3 rounded-lg bg-white/5 border border-white/5 hover:bg-white/10 transition-colors text-[10px] text-zinc-300 font-medium justify-center text-center">
              <Download className="w-3.5 h-3.5 flex-shrink-0" />
              <span className="truncate">Export Data</span>
            </button>
          </div>
        </div>

      </div>
    </div>
  );
}
