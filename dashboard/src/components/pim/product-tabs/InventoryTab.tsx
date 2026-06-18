export function InventoryTab() {
  return (
    <div className="flex flex-col gap-6 pb-10 animate-in fade-in duration-200">
      <div className="flex items-center justify-between">
        <div>
          <h3 className="font-semibold text-[15px]">Inventory Allocations</h3>
          <p className="text-[13px] text-muted-foreground mt-0.5">Track multi-location stock levels and SKU settings.</p>
        </div>
      </div>

      <div className="flex gap-8">
        <div className="flex-1 flex flex-col gap-4">
          <div className="flex flex-col gap-1.5">
            <label className="text-[11px] font-medium text-muted-foreground">SKU (Stock Keeping Unit)</label>
            <input type="text" defaultValue="PRD-1029-A" className="w-full h-10 px-3 rounded-xl bg-white/40 dark:bg-black/20 border border-black/5 dark:border-white/10 text-[13px] focus:outline-none" />
          </div>
          <div className="flex flex-col gap-1.5">
            <label className="text-[11px] font-medium text-muted-foreground">Barcode (ISBN, UPC, GTIN)</label>
            <input type="text" placeholder="e.g. 012345678905" className="w-full h-10 px-3 rounded-xl bg-white/40 dark:bg-black/20 border border-black/5 dark:border-white/10 text-[13px] focus:outline-none" />
          </div>
        </div>

        <div className="flex-1 flex flex-col gap-4">
          <div className="flex items-center justify-between">
            <label className="text-[12px] font-medium text-foreground">Track quantity</label>
            <div className="w-8 h-4 bg-green-500 rounded-full relative cursor-pointer">
              <div className="absolute right-0.5 top-0.5 w-3 h-3 bg-white rounded-full shadow-sm"></div>
            </div>
          </div>
          <div className="flex items-center justify-between">
            <label className="text-[12px] font-medium text-foreground">Continue selling when out of stock</label>
            <div className="w-8 h-4 bg-black/10 dark:bg-white/10 rounded-full relative cursor-pointer">
              <div className="absolute left-0.5 top-0.5 w-3 h-3 bg-white rounded-full shadow-sm"></div>
            </div>
          </div>
        </div>
      </div>

      <h4 className="font-semibold text-[13px] mt-4">Stock by Location</h4>
      <div className="bg-white/40 dark:bg-black/20 border border-black/5 dark:border-white/10 rounded-2xl overflow-hidden">
        <table className="w-full text-sm text-left">
          <thead>
            <tr className="text-[11px] text-muted-foreground uppercase tracking-wider border-b border-black/5 dark:border-white/5">
              <th className="py-3 px-4 font-semibold">Location</th>
              <th className="py-3 px-4 font-semibold text-right">Available</th>
              <th className="py-3 px-4 font-semibold text-right">Reserved</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-black/5 dark:divide-white/5">
            <tr className="hover:bg-black/[0.02] dark:hover:bg-white/[0.02] transition-colors">
              <td className="py-3 px-4 text-[13px] font-medium">Main Warehouse</td>
              <td className="py-3 px-4 text-right">
                <input type="number" defaultValue={1450} className="w-20 h-8 px-2 text-right rounded-md bg-white/50 dark:bg-black/50 border border-black/10 dark:border-white/10 text-[13px] focus:outline-none" />
              </td>
              <td className="py-3 px-4 text-[13px] text-amber-500 text-right font-medium">200</td>
            </tr>
            <tr className="hover:bg-black/[0.02] dark:hover:bg-white/[0.02] transition-colors">
              <td className="py-3 px-4 text-[13px] font-medium">NY Retail Store</td>
              <td className="py-3 px-4 text-right">
                <input type="number" defaultValue={45} className="w-20 h-8 px-2 text-right rounded-md bg-white/50 dark:bg-black/50 border border-black/10 dark:border-white/10 text-[13px] focus:outline-none" />
              </td>
              <td className="py-3 px-4 text-[13px] text-amber-500 text-right font-medium">5</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  );
}
