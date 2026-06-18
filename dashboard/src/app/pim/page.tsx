"use client";

import { PIMLeftSidebar } from "@/components/PIMLeftSidebar";
import { PIMInsightsSidebar } from "@/components/PIMInsightsSidebar";
import { Plus, Search, Edit } from "lucide-react";
import { Button } from "@/components/ui/button";
import { useRouter } from "next/navigation";
import { usePIMStore } from "@/store/pimStore";
import { useGetProducts, useCreateProduct, useGetPIMStats } from "@/hooks/useProducts";

export default function PIMPage() {
  const router = useRouter();
  
  // Local UI State (Zustand)
  const { searchQuery, setSearchQuery } = usePIMStore();
  
  // Server State (React Query)
  const { data: products = [], isLoading: loading, error } = useGetProducts();
  const { data: stats } = useGetPIMStats();
  const createProductMutation = useCreateProduct();

  const formatCurrency = (val: number) => {
    return new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD', maximumFractionDigits: 0 }).format(val);
  };

  const handleAddProduct = () => {
    router.push('/pim/product/new');
  };

  return (
    <>
      {/* Left Sidebar - Specific for PIM */}
      <PIMLeftSidebar />

      {/* Main Content Area */}
      <main className="flex-1 overflow-y-auto custom-scrollbar pr-4 flex flex-col gap-6">
        
        {/* Header */}
        <div className="flex items-center justify-between mt-2">
          <div>
            <h1 className="text-3xl font-bold tracking-tight">PIM & Inventory</h1>
            <p className="text-muted-foreground mt-1">Manage your products, variants and stock across all locations.</p>
          </div>
        </div>

        {/* Large KPI Cards Container */}
        <div className="grid grid-cols-4 gap-4">
          <PIMKPICard title="Total Products" value={(stats?.totalProducts || 0).toString()} subtext="Active SKUs" color="text-green-500" sparkline="M0,15 Q5,10 10,12 T20,8 T30,5" stroke="#22c55e" iconBg="bg-blue-500/10" iconColor="text-blue-500" />
          <PIMKPICard title="Low Stock" value={(stats?.lowStockVariants || 0).toString()} subtext="Need attention" color="text-amber-500" sparkline="M0,15 Q5,20 15,10 T30,0" stroke="#f59e0b" iconBg="bg-amber-500/10" iconColor="text-amber-500" />
          <PIMKPICard title="Out of Stock" value={(stats?.outOfStockVariants || 0).toString()} subtext="Immediate action" color="text-red-500" sparkline="M0,5 Q10,5 15,15 T30,20" stroke="#ef4444" iconBg="bg-red-500/10" iconColor="text-red-500" />
          <PIMKPICard title="Inventory Value" value={formatCurrency(stats?.totalInventoryValue || 0)} subtext="Total valuation" color="text-green-500" sparkline="M0,20 Q10,5 15,10 T30,5" stroke="#22c55e" iconBg="bg-green-500/10" iconColor="text-green-500" />
        </div>

        {/* Tabs & Table Container */}
        <div className="bg-white/20 dark:bg-white/5 rounded-[2rem] p-6 border border-white/30 dark:border-white/10 flex flex-col flex-1">
          
          {/* Sub Navigation Tabs */}
          <div className="flex items-center gap-6 mb-6 px-2">
            <div className="flex items-center gap-2">
              <h2 className="text-lg font-bold">Products</h2>
              <span className="text-xs font-semibold text-muted-foreground bg-black/5 dark:bg-white/5 px-2 py-0.5 rounded-full">{products.length}</span>
            </div>
            
            <div className="ml-auto flex items-center gap-3">
              <div className="relative w-64 mr-2">
                <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
                <input 
                  type="text" 
                  placeholder="Search products, SKU, or category..." 
                  className="w-full h-9 pl-9 pr-4 rounded-full bg-white/40 dark:bg-black/20 border border-black/5 dark:border-white/10 text-xs focus:outline-none focus:ring-2 focus:ring-primary/50"
                />
              </div>
              <Button onClick={handleAddProduct} size="sm" className="rounded-full h-9 text-xs font-medium shadow-sm">
                <Plus className="w-3.5 h-3.5 mr-2" /> Quick Add Product
              </Button>
            </div>
          </div>

          {/* Table */}
          <div className="w-full overflow-x-auto">
            {loading ? (
              <div className="text-sm text-muted-foreground p-4">Loading products from backend...</div>
            ) : error ? (
              <div className="text-sm text-red-500 p-4">Error loading products: {error.message}</div>
            ) : (
            <table className="w-full text-sm text-left">
              <thead>
                <tr className="text-[11px] text-muted-foreground uppercase tracking-wider border-b border-black/5 dark:border-white/5">
                  <th className="pb-3 font-semibold px-2 w-10">
                    <input type="checkbox" className="rounded border-black/10 dark:border-white/10 bg-transparent" />
                  </th>
                  <th className="pb-3 font-semibold px-2">Product</th>
                  <th className="pb-3 font-semibold px-2">Category</th>
                  <th className="pb-3 font-semibold px-2">Status</th>
                  <th className="pb-3 font-semibold px-2">Updated</th>
                  <th className="pb-3 font-semibold px-2 text-right">Actions</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-black/5 dark:divide-white/5">
                {products.map((product, idx) => (
                  <tr 
                    key={idx} 
                    onClick={() => router.push(`/pim/product/${product.id}`)}
                    className="group hover:bg-black/[0.02] dark:hover:bg-white/[0.02] transition-colors cursor-pointer"
                  >
                    <td className="py-3 px-2 w-10">
                      <input type="checkbox" className="rounded border-black/10 dark:border-white/10 bg-transparent opacity-50 group-hover:opacity-100" />
                    </td>
                    <td className="py-3 px-2 w-[240px]">
                      <div className="flex items-center gap-3">
                        <div className="flex flex-col">
                          <span className="font-medium text-sm text-foreground">{product.title}</span>
                          <span className="text-[10px] text-muted-foreground text-ellipsis overflow-hidden whitespace-nowrap max-w-[200px]">{product.id}</span>
                        </div>
                      </div>
                    </td>
                    <td className="py-3 px-2 text-xs text-foreground font-medium">{product.category || "N/A"}</td>
                    <td className="py-3 px-2">
                      <span className="text-[10px] font-semibold px-2 py-0.5 rounded-full bg-green-500/10 text-green-500">{product.status}</span>
                    </td>
                    <td className="py-3 px-2 text-xs text-muted-foreground">{new Date(product.updatedAt).toLocaleDateString()}</td>
                    <td className="py-3 px-2 text-right">
                      <div className="flex items-center justify-end gap-2 opacity-50 group-hover:opacity-100 transition-opacity">
                        <button className="text-muted-foreground hover:text-foreground p-1">
                          <Edit className="w-3.5 h-3.5" />
                        </button>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
            )}
          </div>
          
          {/* Pagination Footer */}
          <div className="flex items-center justify-between mt-auto pt-6 text-xs text-muted-foreground">
            <span>Showing 1 to 8 of 2,450 products</span>
            <div className="flex items-center gap-1">
              <button className="w-6 h-6 rounded-md flex items-center justify-center bg-black/5 dark:bg-white/5 text-foreground font-medium">1</button>
              <button className="w-6 h-6 rounded-md flex items-center justify-center hover:bg-black/5 dark:hover:bg-white/5 transition-colors">2</button>
              <button className="w-6 h-6 rounded-md flex items-center justify-center hover:bg-black/5 dark:hover:bg-white/5 transition-colors">3</button>
              <button className="w-6 h-6 rounded-md flex items-center justify-center hover:bg-black/5 dark:hover:bg-white/5 transition-colors">4</button>
              <button className="w-6 h-6 rounded-md flex items-center justify-center hover:bg-black/5 dark:hover:bg-white/5 transition-colors">5</button>
              <span className="px-1">...</span>
              <button className="w-6 h-6 rounded-md flex items-center justify-center hover:bg-black/5 dark:hover:bg-white/5 transition-colors">307</button>
            </div>
          </div>

        </div>
      </main>

      {/* Right Matte Black Area (PIM Insights) */}
      <aside className="w-[360px] flex-shrink-0 bg-[#121212] text-white p-6 rounded-3xl relative z-20 flex flex-col h-full shadow-[0_8px_30px_rgb(0,0,0,0.4)] border border-white/5">
        <PIMInsightsSidebar stats={stats} />
      </aside>
    </>
  );
}

interface PIMKPICardProps {
  title: string;
  value: string;
  subtext: string;
  color: string;
  sparkline: string;
  stroke: string;
  iconBg: string;
  iconColor: string;
}

function PIMKPICard({ title, value, subtext, color, sparkline, stroke, iconBg, iconColor }: PIMKPICardProps) {
  return (
    <div className="bg-white/20 dark:bg-white/5 rounded-[2rem] p-6 border border-white/30 dark:border-white/10 flex flex-col relative overflow-hidden group">
      <div className="flex items-start gap-4 z-10 relative">
        <div className={`w-12 h-12 rounded-2xl flex items-center justify-center border border-black/5 dark:border-white/5 ${iconBg}`}>
          <div className={`w-6 h-6 rounded ${iconColor.replace('text-', 'bg-')} opacity-80`}></div>
        </div>
        <div className="flex flex-col flex-1">
          <p className="text-xs font-semibold text-muted-foreground uppercase tracking-wider">{title}</p>
          <h3 className="text-2xl font-bold mt-1">{value}</h3>
          <p className={`text-[10px] mt-1 font-semibold text-muted-foreground`}>
            {subtext}
          </p>
        </div>
      </div>
      <div className="absolute right-4 bottom-4 w-20 h-10 opacity-50 group-hover:opacity-100 transition-opacity">
        <svg viewBox="0 0 30 20" fill="none" xmlns="http://www.w3.org/2000/svg" className="w-full h-full">
          <path d={sparkline} stroke={stroke} strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" />
        </svg>
      </div>
    </div>
  );
}
