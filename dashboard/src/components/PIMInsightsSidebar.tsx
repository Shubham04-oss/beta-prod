import { ChevronDown, Plus, Upload, Edit, RefreshCw, Box } from "lucide-react";
import { PieChart, Pie, Cell, ResponsiveContainer } from "recharts";
import { Button } from "./ui/button";
import { PIMStats } from "@/hooks/useProducts";
import { useRouter } from "next/navigation";

interface PIMInsightsSidebarProps {
  stats?: PIMStats;
}

export function PIMInsightsSidebar({ stats }: PIMInsightsSidebarProps) {
  const router = useRouter();
  const inStock = Math.max(0, (stats?.totalProducts || 0) - (stats?.lowStockVariants || 0) - (stats?.outOfStockVariants || 0));
  
  const stockData = [
    { name: "In Stock", value: inStock, color: "#22c55e" },
    { name: "Low Stock", value: stats?.lowStockVariants || 0, color: "#f59e0b" },
    { name: "Out of Stock", value: stats?.outOfStockVariants || 0, color: "#ef4444" },
    { name: "Discontinued", value: 0, color: "#71717a" },
  ];

  const total = stockData.reduce((acc, curr) => acc + curr.value, 0);
  const pct = (val: number) => total > 0 ? Math.round((val / total) * 100) : 0;

  return (
    <div className="flex flex-col h-full w-full">
      {/* Header */}
      <div className="flex items-center justify-between mb-8">
        <h2 className="font-semibold text-lg leading-tight tracking-tight">Inventory Overview</h2>
      </div>

      <div className="flex-1 overflow-y-auto pr-2 custom-scrollbar flex flex-col gap-8 pb-10">
        
        {/* Stock by Status */}
        <div className="bg-zinc-900/50 rounded-2xl p-5 border border-zinc-800">
          <h4 className="text-xs font-medium text-zinc-300 mb-4">Stock by Status</h4>
          
          <div className="flex items-center gap-4">
            {/* Donut Chart */}
            <div className="w-24 h-24 relative flex-shrink-0">
              <ResponsiveContainer width="100%" height="100%">
                <PieChart>
                  <Pie
                    data={stockData}
                    innerRadius={32}
                    outerRadius={44}
                    paddingAngle={2}
                    dataKey="value"
                    stroke="none"
                  >
                    {stockData.map((entry, index) => (
                      <Cell key={`cell-${index}`} fill={entry.color} />
                    ))}
                  </Pie>
                </PieChart>
              </ResponsiveContainer>
              <div className="absolute inset-0 flex flex-col items-center justify-center">
                <span className="text-base font-bold text-white">{total.toLocaleString()}</span>
                <span className="text-[9px] text-zinc-500">Total</span>
              </div>
            </div>

            {/* Legend */}
            <div className="flex flex-col gap-2.5 flex-1 pl-2">
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-2">
                  <div className="w-2 h-2 rounded-full bg-green-500"></div>
                  <span className="text-[10px] text-zinc-400">In Stock</span>
                </div>
                <span className="text-[10px] text-zinc-300">{stockData[0].value.toLocaleString()} <span className="text-zinc-500">({pct(stockData[0].value)}%)</span></span>
              </div>
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-2">
                  <div className="w-2 h-2 rounded-full bg-amber-500"></div>
                  <span className="text-[10px] text-zinc-400">Low Stock</span>
                </div>
                <span className="text-[10px] text-zinc-300">{stockData[1].value.toLocaleString()} <span className="text-zinc-500">({pct(stockData[1].value)}%)</span></span>
              </div>
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-2">
                  <div className="w-2 h-2 rounded-full bg-red-500"></div>
                  <span className="text-[10px] text-zinc-400">Out of Stock</span>
                </div>
                <span className="text-[10px] text-zinc-300">{stockData[2].value.toLocaleString()} <span className="text-zinc-500">({pct(stockData[2].value)}%)</span></span>
              </div>
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-2">
                  <div className="w-2 h-2 rounded-full bg-zinc-500"></div>
                  <span className="text-[10px] text-zinc-400">Discontinued</span>
                </div>
                <span className="text-[10px] text-zinc-300">{stockData[3].value.toLocaleString()} <span className="text-zinc-500">({pct(stockData[3].value)}%)</span></span>
              </div>
            </div>
          </div>
        </div>

        {/* Top Low Stock Products */}
        <div>
          <div className="flex items-center justify-between mb-4">
            <h4 className="text-xs font-semibold text-zinc-400">Top Low Stock Products</h4>
            <button className="text-[10px] text-zinc-400 hover:text-white transition-colors">View all</button>
          </div>
          <div className="flex flex-col gap-3">
            {stats?.topLowStock && stats.topLowStock.length > 0 ? (
              stats.topLowStock.map((item, idx) => (
                <div key={idx} className="flex items-center justify-between p-1 group cursor-pointer" onClick={() => router.push(`/pim/product/${item.product_id}`)}>
                  <div className="flex items-center gap-3">
                    <div className="w-8 h-8 bg-zinc-100 rounded-lg p-1 flex items-center justify-center">
                      <Box className="w-4 h-4 text-zinc-400" />
                    </div>
                    <p className="text-xs font-medium text-zinc-200">{item.product_title}</p>
                  </div>
                  <p className="text-[10px] text-zinc-400">{item.stock_left} left</p>
                </div>
              ))
            ) : (
              <p className="text-xs text-zinc-500 text-center py-4">No low stock items</p>
            )}
          </div>
        </div>

        {/* Quick Actions */}
        <div>
          <h4 className="text-xs font-semibold text-zinc-400 mb-3">Quick Actions</h4>
          <div className="grid grid-cols-2 gap-2">
            <Button onClick={() => router.push('/pim/product/new')} variant="outline" className="bg-zinc-900 border-zinc-800 text-zinc-300 hover:text-white hover:bg-zinc-800 h-10 rounded-xl justify-start px-3 text-[11px]">
              <Plus className="w-3.5 h-3.5 mr-2 text-zinc-400" /> Add Product
            </Button>
            <Button onClick={() => router.push('/pim/import')} variant="outline" className="bg-zinc-900 border-zinc-800 text-zinc-300 hover:text-white hover:bg-zinc-800 h-10 rounded-xl justify-start px-3 text-[11px]">
              <Upload className="w-3.5 h-3.5 mr-2 text-zinc-400" /> Import Products
            </Button>
            <Button onClick={() => router.push('/pim/bulk')} variant="outline" className="bg-zinc-900 border-zinc-800 text-zinc-300 hover:text-white hover:bg-zinc-800 h-10 rounded-xl justify-start px-3 text-[11px]">
              <Edit className="w-3.5 h-3.5 mr-2 text-zinc-400" /> Bulk Update
            </Button>
            <Button onClick={() => router.push('/pim/inventory')} variant="outline" className="bg-zinc-900 border-zinc-800 text-zinc-300 hover:text-white hover:bg-zinc-800 h-10 rounded-xl justify-start px-3 text-[11px]">
              <RefreshCw className="w-3.5 h-3.5 mr-2 text-zinc-400" /> Stock Adjustment
            </Button>
          </div>
        </div>

      </div>
    </div>
  );
}
