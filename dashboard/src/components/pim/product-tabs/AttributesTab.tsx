import { Plus, Trash2 } from "lucide-react";

export function AttributesTab() {
  return (
    <div className="flex flex-col gap-6 pb-10 animate-in fade-in duration-200">
      <div className="flex items-center justify-between">
        <h3 className="font-semibold text-[15px]">Custom Attributes</h3>
        <button className="flex items-center gap-1.5 text-xs font-medium text-primary hover:text-primary/80 transition-colors">
          <Plus className="w-3.5 h-3.5" />
          Add Attribute
        </button>
      </div>
      
      <p className="text-[13px] text-muted-foreground -mt-4">Define product-specific properties like material, care instructions, or technical specs.</p>
      
      <div className="flex flex-col gap-3">
        {/* Attribute Row */}
        <div className="flex items-center gap-4">
          <div className="w-1/3">
            <input type="text" defaultValue="Material" className="w-full h-10 px-3 rounded-xl bg-white/40 dark:bg-black/20 border border-black/5 dark:border-white/10 text-[13px] focus:outline-none" />
          </div>
          <div className="flex-1">
            <input type="text" defaultValue="100% Recycled Aluminum" className="w-full h-10 px-3 rounded-xl bg-white/40 dark:bg-black/20 border border-black/5 dark:border-white/10 text-[13px] focus:outline-none" />
          </div>
          <button className="w-10 h-10 flex items-center justify-center text-muted-foreground hover:text-red-500 transition-colors">
            <Trash2 className="w-4 h-4" />
          </button>
        </div>

        {/* Attribute Row */}
        <div className="flex items-center gap-4">
          <div className="w-1/3">
            <input type="text" defaultValue="Care Instructions" className="w-full h-10 px-3 rounded-xl bg-white/40 dark:bg-black/20 border border-black/5 dark:border-white/10 text-[13px] focus:outline-none" />
          </div>
          <div className="flex-1">
            <input type="text" defaultValue="Wipe clean with a damp cloth" className="w-full h-10 px-3 rounded-xl bg-white/40 dark:bg-black/20 border border-black/5 dark:border-white/10 text-[13px] focus:outline-none" />
          </div>
          <button className="w-10 h-10 flex items-center justify-center text-muted-foreground hover:text-red-500 transition-colors">
            <Trash2 className="w-4 h-4" />
          </button>
        </div>

        {/* Empty Attribute Row */}
        <div className="flex items-center gap-4 opacity-50">
          <div className="w-1/3">
            <input type="text" placeholder="e.g. Dimensions" className="w-full h-10 px-3 rounded-xl bg-white/40 dark:bg-black/20 border border-black/5 dark:border-white/10 border-dashed text-[13px] focus:outline-none" />
          </div>
          <div className="flex-1">
            <input type="text" placeholder="Value" className="w-full h-10 px-3 rounded-xl bg-white/40 dark:bg-black/20 border border-black/5 dark:border-white/10 border-dashed text-[13px] focus:outline-none" />
          </div>
          <div className="w-10"></div>
        </div>
      </div>
    </div>
  );
}
