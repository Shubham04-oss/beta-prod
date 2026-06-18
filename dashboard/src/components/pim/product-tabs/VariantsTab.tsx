import { Plus, MoreHorizontal } from "lucide-react";

export function VariantsTab() {
  return (
    <div className="flex flex-col gap-6 pb-10 animate-in fade-in duration-200">
      <div className="flex items-center justify-between">
        <div>
          <h3 className="font-semibold text-[15px]">Product Variants</h3>
          <p className="text-[13px] text-muted-foreground mt-0.5">Manage SKUs, options, and variant-specific overrides.</p>
        </div>
        <button className="flex items-center gap-1.5 text-xs font-medium bg-black text-white dark:bg-white dark:text-black px-4 py-2 rounded-full hover:opacity-80 transition-opacity">
          <Plus className="w-3.5 h-3.5" />
          Add Options (Size, Color)
        </button>
      </div>
      
      <div className="bg-white/40 dark:bg-black/20 border border-black/5 dark:border-white/10 rounded-2xl overflow-hidden">
        <table className="w-full text-sm text-left">
          <thead>
            <tr className="text-[11px] text-muted-foreground uppercase tracking-wider border-b border-black/5 dark:border-white/5">
              <th className="py-3 px-4 font-semibold">Variant</th>
              <th className="py-3 px-4 font-semibold">SKU</th>
              <th className="py-3 px-4 font-semibold">Price</th>
              <th className="py-3 px-4 font-semibold">Stock</th>
              <th className="py-3 px-4 font-semibold"></th>
            </tr>
          </thead>
          <tbody className="divide-y divide-black/5 dark:divide-white/5">
            <tr className="hover:bg-black/[0.02] dark:hover:bg-white/[0.02] transition-colors">
              <td className="py-3 px-4">
                <div className="flex items-center gap-3">
                  <div className="w-8 h-8 bg-black/5 dark:bg-white/5 rounded-md border border-black/5 flex items-center justify-center text-[10px] text-muted-foreground">Img</div>
                  <span className="font-medium text-[13px]">Black / Medium</span>
                </div>
              </td>
              <td className="py-3 px-4">
                <span className="text-[13px] font-mono text-muted-foreground">PRD-BLK-M</span>
              </td>
              <td className="py-3 px-4">
                <span className="text-[13px]">$199.00</span>
              </td>
              <td className="py-3 px-4">
                <span className="text-[13px] font-medium text-green-500">45 in stock</span>
              </td>
              <td className="py-3 px-4 text-right">
                <button className="text-muted-foreground hover:text-foreground"><MoreHorizontal className="w-4 h-4" /></button>
              </td>
            </tr>
            <tr className="hover:bg-black/[0.02] dark:hover:bg-white/[0.02] transition-colors">
              <td className="py-3 px-4">
                <div className="flex items-center gap-3">
                  <div className="w-8 h-8 bg-black/5 dark:bg-white/5 rounded-md border border-black/5 flex items-center justify-center text-[10px] text-muted-foreground">Img</div>
                  <span className="font-medium text-[13px]">Black / Large</span>
                </div>
              </td>
              <td className="py-3 px-4">
                <span className="text-[13px] font-mono text-muted-foreground">PRD-BLK-L</span>
              </td>
              <td className="py-3 px-4">
                <span className="text-[13px]">$199.00</span>
              </td>
              <td className="py-3 px-4">
                <span className="text-[13px] font-medium text-amber-500">2 in stock</span>
              </td>
              <td className="py-3 px-4 text-right">
                <button className="text-muted-foreground hover:text-foreground"><MoreHorizontal className="w-4 h-4" /></button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  );
}
