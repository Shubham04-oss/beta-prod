"use client";

import { useState } from "react";
import { PIMLeftSidebar } from "@/components/PIMLeftSidebar";
import { Search, Plus, ArrowDownToLine, ArrowUpFromLine, RefreshCw } from "lucide-react";
import { Button } from "@/components/ui/button";
import { InventoryInsightsSidebar } from "@/components/InventoryInsightsSidebar";

export default function InventoryPage() {
  const [searchQuery, setSearchQuery] = useState("");
  const [isAdjusting, setIsAdjusting] = useState(false);
  const [selectedVariant, setSelectedVariant] = useState<any>(null);
  const [adjustQuantity, setAdjustQuantity] = useState(0);
  const [adjustReason, setAdjustReason] = useState("RESTOCK");
  const [adjustCost, setAdjustCost] = useState(0);

  // Mock data for UI presentation until backend endpoint is confirmed
  const mockInventory = [
    { id: "inv-1", sku: "AIRPODS-PRO-2", title: "AirPods Pro (2nd Gen)", location: "Main Warehouse", available: 1450, reserved: 200, unitCost: 189.50 },
    { id: "inv-2", sku: "AIRPODS-PRO-2", title: "AirPods Pro (2nd Gen)", location: "NY Store (POS)", available: 45, reserved: 5, unitCost: 189.50 },
    { id: "inv-3", sku: "MACBOOK-M3-PRO", title: "MacBook Pro 14 M3", location: "Main Warehouse", available: 210, reserved: 45, unitCost: 1540.00 },
    { id: "inv-4", sku: "SONY-WH1000XM5", title: "Sony WH-1000XM5", location: "Main Warehouse", available: 0, reserved: 0, unitCost: 295.00 },
    { id: "inv-5", sku: "IPHONE-15-PRO-BLK", title: "iPhone 15 Pro - Black Titanium", location: "LA Store (POS)", available: 12, reserved: 2, unitCost: 890.00 },
  ];

  const filteredInventory = mockInventory.filter(item => 
    item.title.toLowerCase().includes(searchQuery.toLowerCase()) || 
    item.sku.toLowerCase().includes(searchQuery.toLowerCase())
  );

  const handleAdjustClick = (item: any) => {
    setSelectedVariant(item);
    setAdjustQuantity(0);
    setAdjustCost(item.unitCost);
    setIsAdjusting(true);
  };

  const submitAdjustment = async () => {
    // Here we would call useAdjustInventory mutation
    alert(`Adjusted ${adjustQuantity} units for ${selectedVariant.sku} at $${adjustCost} (Reason: ${adjustReason})`);
    setIsAdjusting(false);
  };

  const formatCurrency = (val: number) => {
    return new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD' }).format(val);
  };

  return (
    <>
      <PIMLeftSidebar />

      <main className="flex-1 overflow-y-auto custom-scrollbar pr-4 flex flex-col gap-6">
        
        {/* Header */}
        <div className="flex items-center justify-between mt-2">
          <div>
            <h1 className="text-3xl font-bold tracking-tight">Inventory Management</h1>
            <p className="text-muted-foreground mt-1">Track multi-location stock, reservations, and perform adjustments.</p>
          </div>
        </div>

        {/* Large KPI Cards */}
        <div className="grid grid-cols-3 gap-4">
          <div className="bg-white/20 dark:bg-white/5 rounded-[2rem] p-6 border border-white/30 dark:border-white/10 flex flex-col relative overflow-hidden group">
             <p className="text-xs font-semibold text-muted-foreground uppercase tracking-wider">Total On-Hand</p>
             <h3 className="text-3xl font-bold mt-2">1,717</h3>
             <p className="text-[10px] mt-1 font-semibold text-green-500">Across all locations</p>
          </div>
          <div className="bg-white/20 dark:bg-white/5 rounded-[2rem] p-6 border border-white/30 dark:border-white/10 flex flex-col relative overflow-hidden group">
             <p className="text-xs font-semibold text-muted-foreground uppercase tracking-wider">Reserved (Orders)</p>
             <h3 className="text-3xl font-bold mt-2">252</h3>
             <p className="text-[10px] mt-1 font-semibold text-amber-500">Allocated for fulfillment</p>
          </div>
          <div className="bg-white/20 dark:bg-white/5 rounded-[2rem] p-6 border border-white/30 dark:border-white/10 flex flex-col relative overflow-hidden group">
             <p className="text-xs font-semibold text-muted-foreground uppercase tracking-wider">WAC Inventory Value</p>
             <h3 className="text-3xl font-bold mt-2">$712,450.00</h3>
             <p className="text-[10px] mt-1 font-semibold text-blue-500">Weighted Average Cost</p>
          </div>
        </div>

        {/* Inventory Table Container */}
        <div className="bg-white/20 dark:bg-white/5 rounded-[2rem] p-6 border border-white/30 dark:border-white/10 flex flex-col flex-1">
          
          <div className="flex items-center gap-6 mb-6 px-2">
            <h2 className="text-lg font-bold">Stock Levels</h2>
            <div className="ml-auto flex items-center gap-3">
              <div className="relative w-64 mr-2">
                <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
                <input 
                  type="text" 
                  placeholder="Search SKU or Product..." 
                  value={searchQuery}
                  onChange={e => setSearchQuery(e.target.value)}
                  className="w-full h-9 pl-9 pr-4 rounded-full bg-white/40 dark:bg-black/20 border border-black/5 dark:border-white/10 text-xs focus:outline-none focus:ring-2 focus:ring-primary/50"
                />
              </div>
            </div>
          </div>

          <div className="w-full overflow-x-auto">
            <table className="w-full text-sm text-left">
              <thead>
                <tr className="text-[11px] text-muted-foreground uppercase tracking-wider border-b border-black/5 dark:border-white/5">
                  <th className="pb-3 font-semibold px-2">SKU / Product</th>
                  <th className="pb-3 font-semibold px-2">Location</th>
                  <th className="pb-3 font-semibold px-2 text-right">Available</th>
                  <th className="pb-3 font-semibold px-2 text-right">Reserved</th>
                  <th className="pb-3 font-semibold px-2 text-right">WAC (Unit Cost)</th>
                  <th className="pb-3 font-semibold px-2 text-right">Actions</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-black/5 dark:divide-white/5">
                {filteredInventory.map((item) => (
                  <tr key={item.id} className="group hover:bg-black/[0.02] dark:hover:bg-white/[0.02] transition-colors">
                    <td className="py-3 px-2">
                      <div className="flex flex-col">
                        <span className="font-medium text-sm text-foreground">{item.sku}</span>
                        <span className="text-[10px] text-muted-foreground">{item.title}</span>
                      </div>
                    </td>
                    <td className="py-3 px-2">
                      <span className="text-xs px-2 py-1 bg-black/5 dark:bg-white/5 rounded-md font-medium">{item.location}</span>
                    </td>
                    <td className="py-3 px-2 text-right">
                      <span className={`font-bold ${item.available === 0 ? "text-red-500" : "text-foreground"}`}>
                        {item.available}
                      </span>
                    </td>
                    <td className="py-3 px-2 text-right">
                      <span className="font-medium text-amber-500">{item.reserved}</span>
                    </td>
                    <td className="py-3 px-2 text-right font-medium text-muted-foreground">
                      {formatCurrency(item.unitCost)}
                    </td>
                    <td className="py-3 px-2 text-right">
                      <div className="flex items-center justify-end gap-2">
                        <Button onClick={() => handleAdjustClick(item)} size="sm" variant="outline" className="h-7 text-[10px] rounded-md bg-white/40 dark:bg-black/20 border-black/10 dark:border-white/10 shadow-sm">
                          Adjust
                        </Button>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>

      </main>

      {/* Right Matte Black Area */}
      <aside className="w-[360px] flex-shrink-0 bg-[#121212] text-white p-6 rounded-3xl relative z-20 flex flex-col h-full shadow-[0_8px_30px_rgb(0,0,0,0.4)] border border-white/5">
        <InventoryInsightsSidebar />
      </aside>

      {/* Adjust Stock Overlay (Simple visual modal) */}
      {isAdjusting && selectedVariant && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 backdrop-blur-sm">
          <div className="w-[400px] bg-white dark:bg-[#121212] rounded-3xl p-6 shadow-2xl border border-black/10 dark:border-white/10 flex flex-col gap-6 animate-in fade-in zoom-in-95 duration-200">
            <div>
              <h2 className="text-xl font-bold">Adjust Stock</h2>
              <p className="text-xs text-muted-foreground mt-1">
                {selectedVariant.title} ({selectedVariant.sku}) at {selectedVariant.location}
              </p>
            </div>

            <div className="flex flex-col gap-4">
              <div className="flex flex-col gap-1.5">
                <label className="text-[11px] font-medium text-muted-foreground">Adjustment Reason</label>
                <select 
                  value={adjustReason}
                  onChange={(e) => setAdjustReason(e.target.value)}
                  className="w-full h-10 px-3 rounded-xl bg-black/5 dark:bg-white/5 border border-black/5 dark:border-white/10 text-[13px] focus:outline-none"
                >
                  <option value="RESTOCK">Restock (Add)</option>
                  <option value="SHRINKAGE">Shrinkage / Damage (Remove)</option>
                  <option value="COUNT_CORRECTION">Cycle Count Correction</option>
                </select>
              </div>

              <div className="flex gap-4">
                <div className="flex flex-col gap-1.5 flex-1">
                  <label className="text-[11px] font-medium text-muted-foreground">Quantity (+/-)</label>
                  <input 
                    type="number" 
                    value={adjustQuantity}
                    onChange={(e) => setAdjustQuantity(Number(e.target.value))}
                    className="w-full h-10 px-3 rounded-xl bg-black/5 dark:bg-white/5 border border-black/5 dark:border-white/10 text-[13px] focus:outline-none" 
                  />
                </div>
                <div className="flex flex-col gap-1.5 flex-1">
                  <label className="text-[11px] font-medium text-muted-foreground">Unit Cost (For WAC)</label>
                  <input 
                    type="number" 
                    disabled={adjustReason === "SHRINKAGE"}
                    value={adjustCost}
                    onChange={(e) => setAdjustCost(Number(e.target.value))}
                    className="w-full h-10 px-3 rounded-xl bg-black/5 dark:bg-white/5 border border-black/5 dark:border-white/10 text-[13px] focus:outline-none disabled:opacity-50" 
                  />
                </div>
              </div>

            </div>

            <div className="flex items-center justify-end gap-3 mt-4">
              <Button onClick={() => setIsAdjusting(false)} variant="ghost" className="text-xs">Cancel</Button>
              <Button onClick={submitAdjustment} className="rounded-full shadow-sm bg-black text-white hover:bg-zinc-800 dark:bg-white dark:text-black dark:hover:bg-zinc-200">
                Confirm Adjustment
              </Button>
            </div>
          </div>
        </div>
      )}

    </>
  );
}
