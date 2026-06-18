import { DollarSign, Percent } from "lucide-react";

export function PricingTab() {
  return (
    <div className="flex gap-8 pb-10 animate-in fade-in duration-200">
      <div className="flex-1 flex flex-col gap-6">
        <h3 className="font-semibold text-[15px]">Pricing & Margins</h3>
        
        <div className="flex flex-col gap-1.5">
          <label className="text-[11px] font-medium text-muted-foreground">Base Price</label>
          <div className="relative">
            <input type="number" defaultValue={199} className="w-full h-10 pl-9 pr-3 rounded-xl bg-white/40 dark:bg-black/20 border border-black/5 dark:border-white/10 text-[13px] focus:outline-none focus:ring-2 focus:ring-primary/50" />
            <DollarSign className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground pointer-events-none" />
          </div>
        </div>

        <div className="flex flex-col gap-1.5">
          <label className="text-[11px] font-medium text-muted-foreground">Compare-at Price (MSRP)</label>
          <div className="relative">
            <input type="number" defaultValue={249} className="w-full h-10 pl-9 pr-3 rounded-xl bg-white/40 dark:bg-black/20 border border-black/5 dark:border-white/10 text-[13px] focus:outline-none focus:ring-2 focus:ring-primary/50" />
            <DollarSign className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground pointer-events-none" />
          </div>
          <p className="text-[10px] text-muted-foreground">To show a reduced price, move the original price into Compare-at price.</p>
        </div>

        <div className="flex flex-col gap-1.5">
          <label className="text-[11px] font-medium text-muted-foreground">Cost Price</label>
          <div className="relative">
            <input type="number" defaultValue={90} className="w-full h-10 pl-9 pr-3 rounded-xl bg-white/40 dark:bg-black/20 border border-black/5 dark:border-white/10 text-[13px] focus:outline-none focus:ring-2 focus:ring-primary/50" />
            <DollarSign className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground pointer-events-none" />
          </div>
          <p className="text-[10px] text-muted-foreground">Customers won't see this.</p>
        </div>
      </div>

      <div className="flex-1 flex flex-col gap-6">
        <h3 className="font-semibold text-[15px]">Taxes & Rules</h3>

        <div className="flex flex-col gap-4 bg-black/5 dark:bg-white/5 p-4 rounded-2xl border border-black/5 dark:border-white/5">
          <div className="flex items-center justify-between">
            <span className="text-[12px] font-medium">Profit Margin</span>
            <span className="text-[12px] font-bold text-green-500">54.7%</span>
          </div>
          <div className="flex items-center justify-between">
            <span className="text-[12px] font-medium">Gross Profit</span>
            <span className="text-[12px] font-bold text-green-500">$109.00</span>
          </div>
        </div>

        <div className="flex items-center justify-between mt-2">
          <label className="text-[12px] font-medium text-foreground">Charge tax on this product</label>
          <div className="w-8 h-4 bg-green-500 rounded-full relative cursor-pointer">
            <div className="absolute right-0.5 top-0.5 w-3 h-3 bg-white rounded-full shadow-sm"></div>
          </div>
        </div>

        <div className="flex flex-col gap-1.5 mt-2">
          <label className="text-[11px] font-medium text-muted-foreground">Tax Code</label>
          <input type="text" placeholder="e.g. TX-102" className="w-full h-10 px-3 rounded-xl bg-white/40 dark:bg-black/20 border border-black/5 dark:border-white/10 text-[13px] focus:outline-none focus:ring-2 focus:ring-primary/50" />
        </div>
      </div>
    </div>
  );
}
