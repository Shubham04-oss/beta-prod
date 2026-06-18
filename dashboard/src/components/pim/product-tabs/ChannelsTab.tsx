import { Globe, Store, MonitorSmartphone, Plus } from "lucide-react";

export function ChannelsTab() {
  return (
    <div className="flex flex-col gap-6 pb-10 animate-in fade-in duration-200">
      <div className="flex items-center justify-between">
        <div>
          <h3 className="font-semibold text-[15px]">Sales Channels & Markets</h3>
          <p className="text-[13px] text-muted-foreground mt-0.5">Manage where this product is available for sale and configure overrides.</p>
        </div>
      </div>

      <div className="flex flex-col gap-3">
        {/* Channel Row */}
        <div className="flex items-center gap-6 p-4 rounded-2xl bg-white/40 dark:bg-black/20 border border-black/5 dark:border-white/10">
          <div className="w-10 h-10 rounded-full bg-black/5 dark:bg-white/5 flex items-center justify-center">
            <Globe className="w-5 h-5 text-foreground" />
          </div>
          <div className="flex-1">
            <h4 className="text-[13px] font-semibold">Online Store (Global)</h4>
            <p className="text-[11px] text-muted-foreground">Available on main D2C website</p>
          </div>
          <div className="flex items-center gap-4">
            <button className="text-[11px] font-medium text-primary hover:underline">Edit Overrides</button>
            <div className="w-8 h-4 bg-green-500 rounded-full relative cursor-pointer">
              <div className="absolute right-0.5 top-0.5 w-3 h-3 bg-white rounded-full shadow-sm"></div>
            </div>
          </div>
        </div>

        {/* Channel Row */}
        <div className="flex items-center gap-6 p-4 rounded-2xl bg-white/40 dark:bg-black/20 border border-black/5 dark:border-white/10">
          <div className="w-10 h-10 rounded-full bg-black/5 dark:bg-white/5 flex items-center justify-center">
            <MonitorSmartphone className="w-5 h-5 text-foreground" />
          </div>
          <div className="flex-1">
            <h4 className="text-[13px] font-semibold">B2B Portal</h4>
            <p className="text-[11px] text-muted-foreground">Available for wholesale customers (Includes pricing overrides)</p>
          </div>
          <div className="flex items-center gap-4">
            <button className="text-[11px] font-medium text-primary hover:underline">Edit Overrides</button>
            <div className="w-8 h-4 bg-green-500 rounded-full relative cursor-pointer">
              <div className="absolute right-0.5 top-0.5 w-3 h-3 bg-white rounded-full shadow-sm"></div>
            </div>
          </div>
        </div>

        {/* Channel Row */}
        <div className="flex items-center gap-6 p-4 rounded-2xl bg-white/40 dark:bg-black/20 border border-black/5 dark:border-white/10">
          <div className="w-10 h-10 rounded-full bg-black/5 dark:bg-white/5 flex items-center justify-center">
            <Store className="w-5 h-5 text-foreground" />
          </div>
          <div className="flex-1">
            <h4 className="text-[13px] font-semibold">Retail POS (New York)</h4>
            <p className="text-[11px] text-muted-foreground">In-store availability</p>
          </div>
          <div className="flex items-center gap-4">
            <button className="text-[11px] font-medium text-primary hover:underline">Edit Overrides</button>
            <div className="w-8 h-4 bg-black/10 dark:bg-white/10 rounded-full relative cursor-pointer">
              <div className="absolute left-0.5 top-0.5 w-3 h-3 bg-white rounded-full shadow-sm"></div>
            </div>
          </div>
        </div>

      </div>
    </div>
  );
}
