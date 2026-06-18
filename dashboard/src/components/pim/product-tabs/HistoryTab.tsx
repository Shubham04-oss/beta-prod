import { History, User, Pencil, PackagePlus } from "lucide-react";

export function HistoryTab() {
  return (
    <div className="flex flex-col gap-6 pb-10 animate-in fade-in duration-200">
      <div>
        <h3 className="font-semibold text-[15px]">Audit Log & History</h3>
        <p className="text-[13px] text-muted-foreground mt-0.5">Track every change made to this product over time.</p>
      </div>

      <div className="flex flex-col relative before:absolute before:inset-0 before:ml-5 before:-translate-x-px md:before:mx-auto md:before:translate-x-0 before:h-full before:w-0.5 before:bg-gradient-to-b before:from-transparent before:via-black/10 dark:before:via-white/10 before:to-transparent">
        
        {/* Timeline Item */}
        <div className="relative flex items-center justify-between md:justify-normal md:odd:flex-row-reverse group is-active mb-8">
          <div className="flex items-center justify-center w-10 h-10 rounded-full border-4 border-white dark:border-[#0a0a0a] bg-black/5 dark:bg-white/10 text-foreground shrink-0 md:order-1 md:group-odd:-translate-x-1/2 md:group-even:translate-x-1/2 shadow-sm z-10">
            <Pencil className="w-4 h-4" />
          </div>
          <div className="w-[calc(100%-4rem)] md:w-[calc(50%-2.5rem)] p-4 rounded-2xl bg-white/40 dark:bg-black/20 border border-black/5 dark:border-white/10 shadow-sm">
            <div className="flex items-center justify-between mb-1">
              <div className="font-bold text-[13px]">Updated Price</div>
              <time className="text-[10px] font-medium text-muted-foreground">2 hours ago</time>
            </div>
            <div className="text-[12px] text-muted-foreground">
              Changed base price from $149.00 to $199.00.
            </div>
            <div className="flex items-center gap-1.5 mt-3 text-[10px] font-medium text-muted-foreground">
              <User className="w-3 h-3" /> Sophia
            </div>
          </div>
        </div>

        {/* Timeline Item */}
        <div className="relative flex items-center justify-between md:justify-normal md:odd:flex-row-reverse group is-active mb-8">
          <div className="flex items-center justify-center w-10 h-10 rounded-full border-4 border-white dark:border-[#0a0a0a] bg-black/5 dark:bg-white/10 text-foreground shrink-0 md:order-1 md:group-odd:-translate-x-1/2 md:group-even:translate-x-1/2 shadow-sm z-10">
            <PackagePlus className="w-4 h-4" />
          </div>
          <div className="w-[calc(100%-4rem)] md:w-[calc(50%-2.5rem)] p-4 rounded-2xl bg-white/40 dark:bg-black/20 border border-black/5 dark:border-white/10 shadow-sm">
            <div className="flex items-center justify-between mb-1">
              <div className="font-bold text-[13px]">Added Variant</div>
              <time className="text-[10px] font-medium text-muted-foreground">3 days ago</time>
            </div>
            <div className="text-[12px] text-muted-foreground">
              Added "Black / Large" SKU: PRD-BLK-L.
            </div>
            <div className="flex items-center gap-1.5 mt-3 text-[10px] font-medium text-muted-foreground">
              <User className="w-3 h-3" /> System
            </div>
          </div>
        </div>

        {/* Timeline Item */}
        <div className="relative flex items-center justify-between md:justify-normal md:odd:flex-row-reverse group is-active">
          <div className="flex items-center justify-center w-10 h-10 rounded-full border-4 border-white dark:border-[#0a0a0a] bg-black text-white shrink-0 md:order-1 md:group-odd:-translate-x-1/2 md:group-even:translate-x-1/2 shadow-sm z-10">
            <History className="w-4 h-4" />
          </div>
          <div className="w-[calc(100%-4rem)] md:w-[calc(50%-2.5rem)] p-4 rounded-2xl bg-white/40 dark:bg-black/20 border border-black/5 dark:border-white/10 shadow-sm">
            <div className="flex items-center justify-between mb-1">
              <div className="font-bold text-[13px]">Product Created</div>
              <time className="text-[10px] font-medium text-muted-foreground">Apr 30, 2024</time>
            </div>
            <div className="text-[12px] text-muted-foreground">
              Initial product creation via Product Configurator.
            </div>
            <div className="flex items-center gap-1.5 mt-3 text-[10px] font-medium text-muted-foreground">
              <User className="w-3 h-3" /> Sophia
            </div>
          </div>
        </div>

      </div>
    </div>
  );
}
