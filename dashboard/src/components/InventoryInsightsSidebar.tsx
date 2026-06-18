import { Package, TrendingUp, History, AlertTriangle, ArrowRight } from "lucide-react";

export function InventoryInsightsSidebar() {
  return (
    <div className="flex flex-col h-full w-full">
      <div className="flex-1 overflow-y-auto pr-2 custom-scrollbar flex flex-col gap-8 pb-10">
        
        {/* Inventory Health */}
        <div>
          <h4 className="font-semibold text-[13px] text-zinc-100 mb-4">Inventory Health</h4>
          <div className="flex flex-col gap-4">
            
            <div className="flex items-center justify-between">
              <span className="text-[11px] text-zinc-400">Stockout Risk</span>
              <div className="flex items-center gap-1.5">
                <div className="w-1.5 h-1.5 rounded-full bg-amber-500"></div>
                <span className="text-[11px] font-semibold text-amber-500">Moderate</span>
              </div>
            </div>

            <div className="flex flex-col gap-2 mt-2">
              <div className="flex items-center justify-between">
                <span className="text-[11px] text-zinc-400">Items Low on Stock</span>
                <span className="text-[11px] font-bold text-red-400">12</span>
              </div>
              <div className="w-full h-1.5 bg-zinc-800 rounded-full overflow-hidden">
                <div className="h-full bg-red-500 w-[15%]"></div>
              </div>
            </div>

            <div className="flex flex-col gap-2 mt-1">
              <div className="flex items-center justify-between">
                <span className="text-[11px] text-zinc-400">Items Overstocked</span>
                <span className="text-[11px] font-bold text-blue-400">45</span>
              </div>
              <div className="w-full h-1.5 bg-zinc-800 rounded-full overflow-hidden">
                <div className="h-full bg-blue-500 w-[40%]"></div>
              </div>
            </div>
            
          </div>
        </div>

        {/* Alerts & Reorder */}
        <div>
          <div className="flex items-center justify-between mb-4">
            <h4 className="font-semibold text-[13px] text-zinc-100">Low Stock Alerts</h4>
            <button className="text-[10px] text-zinc-400 hover:text-white transition-colors">View All</button>
          </div>
          
          <div className="flex flex-col gap-3">
            <div className="bg-red-500/10 border border-red-500/20 rounded-xl p-3 flex gap-3">
              <AlertTriangle className="w-4 h-4 text-red-500 shrink-0 mt-0.5" />
              <div>
                <p className="text-[11px] font-semibold text-red-400">Sony WH-1000XM5</p>
                <p className="text-[10px] text-red-500/70 mt-0.5">0 available across all locations.</p>
                <button className="text-[10px] font-medium text-white mt-2 flex items-center gap-1 hover:underline">
                  Create PO <ArrowRight className="w-3 h-3" />
                </button>
              </div>
            </div>

            <div className="bg-amber-500/10 border border-amber-500/20 rounded-xl p-3 flex gap-3">
              <Package className="w-4 h-4 text-amber-500 shrink-0 mt-0.5" />
              <div>
                <p className="text-[11px] font-semibold text-amber-400">iPhone 15 Pro - Black</p>
                <p className="text-[10px] text-amber-500/70 mt-0.5">Only 12 left in LA Store.</p>
                <button className="text-[10px] font-medium text-white mt-2 flex items-center gap-1 hover:underline">
                  Transfer Stock <ArrowRight className="w-3 h-3" />
                </button>
              </div>
            </div>
          </div>
        </div>

        {/* Recent Adjustments */}
        <div>
          <h4 className="font-semibold text-[13px] text-zinc-100 mb-3">Recent Activity</h4>
          <div className="flex flex-col gap-1 relative before:absolute before:inset-0 before:ml-3.5 before:-translate-x-px before:h-full before:w-0.5 before:bg-gradient-to-b before:from-zinc-800 before:to-transparent">
            
            <div className="relative flex items-start gap-4 mb-4">
              <div className="flex items-center justify-center w-7 h-7 rounded-full border-4 border-[#121212] bg-zinc-800 text-zinc-400 shrink-0 z-10">
                <History className="w-3 h-3" />
              </div>
              <div className="pt-1">
                <p className="text-[11px] font-semibold text-zinc-200">Restock (+50)</p>
                <p className="text-[10px] text-zinc-500 mt-0.5">AirPods Pro (2nd Gen)</p>
                <p className="text-[9px] text-zinc-600 mt-1">2 hours ago</p>
              </div>
            </div>

            <div className="relative flex items-start gap-4">
              <div className="flex items-center justify-center w-7 h-7 rounded-full border-4 border-[#121212] bg-zinc-800 text-zinc-400 shrink-0 z-10">
                <TrendingUp className="w-3 h-3" />
              </div>
              <div className="pt-1">
                <p className="text-[11px] font-semibold text-zinc-200">Shrinkage (-2)</p>
                <p className="text-[10px] text-zinc-500 mt-0.5">MacBook Pro 14 M3</p>
                <p className="text-[9px] text-zinc-600 mt-1">Yesterday</p>
              </div>
            </div>

          </div>
        </div>

      </div>
    </div>
  );
}
