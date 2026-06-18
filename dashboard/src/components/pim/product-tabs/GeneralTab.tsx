import { ChevronDown, CalendarDays } from "lucide-react";

interface GeneralTabProps {
  product?: any;
  title: string;
  setTitle: (val: string) => void;
  description: string;
  setDescription: (val: string) => void;
  category: string;
  setCategory: (val: string) => void;
}

export function GeneralTab({ product, title, setTitle, description, setDescription, category, setCategory }: GeneralTabProps) {
  const isDraft = !product;

  return (
    <div className="flex gap-8 pb-10 animate-in fade-in duration-200">
      {/* Left Column: Basic Information */}
      <div className="flex-1 flex flex-col gap-6">
        <h3 className="font-semibold text-[15px]">Basic Information</h3>
        
        <div className="flex flex-col gap-1.5">
          <label className="text-[11px] font-medium text-muted-foreground">Product Name <span className="text-red-500">*</span></label>
          <input type="text" placeholder="e.g. Wireless Noise-Cancelling Headphones" value={title} onChange={(e) => setTitle(e.target.value)} className="w-full h-10 px-3 rounded-xl bg-white/40 dark:bg-black/20 border border-black/5 dark:border-white/10 text-[13px] focus:outline-none focus:ring-2 focus:ring-primary/50" />
        </div>

        <div className="flex flex-col gap-1.5">
          <label className="text-[11px] font-medium text-muted-foreground">ID</label>
          <input type="text" disabled value={isDraft ? "" : product.id} placeholder={isDraft ? "Pending" : ""} className="w-full h-10 px-3 rounded-xl bg-white/40 dark:bg-black/20 border border-black/5 dark:border-white/10 text-[13px] focus:outline-none opacity-50" />
        </div>

        <div className="flex flex-col gap-1.5">
          <label className="text-[11px] font-medium text-muted-foreground">Short Description</label>
          <textarea rows={2} placeholder="A short description of the product" value={description} onChange={(e) => setDescription(e.target.value)} className="w-full py-2 px-3 rounded-xl bg-white/40 dark:bg-black/20 border border-black/5 dark:border-white/10 text-[13px] focus:outline-none focus:ring-2 focus:ring-primary/50 resize-none"></textarea>
        </div>

        <div className="flex flex-col gap-1.5">
          <label className="text-[11px] font-medium text-muted-foreground">Category</label>
          <input type="text" placeholder="e.g. Electronics" value={category} onChange={(e) => setCategory(e.target.value)} className="w-full h-10 px-3 rounded-xl bg-white/40 dark:bg-black/20 border border-black/5 dark:border-white/10 text-[13px] focus:outline-none focus:ring-2 focus:ring-primary/50" />
        </div>

        <div className="flex flex-col gap-1.5">
          <label className="text-[11px] font-medium text-muted-foreground">Brand <span className="text-red-500">*</span></label>
          <div className="relative">
            <select className="w-full h-10 px-3 rounded-xl bg-white/40 dark:bg-black/20 border border-black/5 dark:border-white/10 text-[13px] focus:outline-none appearance-none pr-8">
              <option>Select a brand...</option>
              <option>Sony</option>
            </select>
            <ChevronDown className="absolute right-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground pointer-events-none" />
          </div>
        </div>

        <div className="flex flex-col gap-1.5">
          <label className="text-[11px] font-medium text-muted-foreground">Tags</label>
          <div className="w-full min-h-[40px] p-1.5 rounded-xl bg-white/40 dark:bg-black/20 border border-black/5 dark:border-white/10 flex flex-wrap gap-1.5">
            {!isDraft ? (
              <>
                <span className="px-2 py-1 bg-black/5 dark:bg-white/10 rounded-md text-[11px] flex items-center gap-1">wireless <button className="text-muted-foreground hover:text-foreground">×</button></span>
                <span className="px-2 py-1 bg-black/5 dark:bg-white/10 rounded-md text-[11px] flex items-center gap-1">premium <button className="text-muted-foreground hover:text-foreground">×</button></span>
              </>
            ) : (
              <span className="px-2 py-1 bg-black/5 dark:bg-white/10 rounded-md text-[11px] flex items-center gap-1 text-muted-foreground">No tags added</span>
            )}
          </div>
        </div>

      </div>

      {/* Right Column: Key Details */}
      <div className="flex-1 flex flex-col gap-6">
        <h3 className="font-semibold text-[15px]">Key Details</h3>
        
        <div className="flex flex-col gap-1.5">
          <label className="text-[11px] font-medium text-muted-foreground">Product Type</label>
          <div className="relative">
            <select className="w-full h-10 px-3 rounded-xl bg-white/40 dark:bg-black/20 border border-black/5 dark:border-white/10 text-[13px] focus:outline-none appearance-none pr-8">
              <option>Simple Product</option>
              <option>Configurable Product</option>
              <option>Virtual Product</option>
            </select>
            <ChevronDown className="absolute right-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground pointer-events-none" />
          </div>
        </div>

        <div className="flex flex-col gap-1.5">
          <label className="text-[11px] font-medium text-muted-foreground">Status</label>
          <div className="relative">
            <select className="w-full h-10 px-3 rounded-xl bg-white/40 dark:bg-black/20 border border-black/5 dark:border-white/10 text-[13px] focus:outline-none appearance-none pr-8">
              <option>Draft</option>
              <option>Active</option>
            </select>
            <ChevronDown className="absolute right-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground pointer-events-none" />
          </div>
        </div>

        <div className="flex flex-col gap-1.5">
          <label className="text-[11px] font-medium text-muted-foreground">Launch Date</label>
          <div className="relative">
            <input type="text" placeholder="Select date" defaultValue={!isDraft ? "May 1, 2024" : ""} className="w-full h-10 pl-9 pr-3 rounded-xl bg-white/40 dark:bg-black/20 border border-black/5 dark:border-white/10 text-[13px] focus:outline-none focus:ring-2 focus:ring-primary/50" />
            <CalendarDays className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground pointer-events-none" />
          </div>
        </div>

        <div className="flex flex-col gap-1.5">
          <label className="text-[11px] font-medium text-muted-foreground">Discontinue Date</label>
          <div className="relative">
            <input type="text" placeholder="Select date" className="w-full h-10 pl-9 pr-3 rounded-xl bg-white/40 dark:bg-black/20 border border-black/5 dark:border-white/10 text-[13px] focus:outline-none focus:ring-2 focus:ring-primary/50" />
            <CalendarDays className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground pointer-events-none" />
          </div>
        </div>

        <div className="flex flex-col gap-1.5">
          <label className="text-[11px] font-medium text-muted-foreground">Warranty Period</label>
          <input type="text" placeholder="e.g. 12 Months" defaultValue={!isDraft ? "12 Months" : ""} className="w-full h-10 px-3 rounded-xl bg-white/40 dark:bg-black/20 border border-black/5 dark:border-white/10 text-[13px] focus:outline-none focus:ring-2 focus:ring-primary/50" />
        </div>

        <div className="flex items-center justify-between mt-2">
          <label className="text-[11px] font-medium text-muted-foreground">Taxable</label>
          <div className="w-8 h-4 bg-green-500 rounded-full relative cursor-pointer">
            <div className="absolute right-0.5 top-0.5 w-3 h-3 bg-white rounded-full shadow-sm"></div>
          </div>
        </div>

        <div className="flex items-center justify-between mt-2">
          <label className="text-[11px] font-medium text-muted-foreground">Returnable</label>
          <div className="w-8 h-4 bg-green-500 rounded-full relative cursor-pointer">
            <div className="absolute right-0.5 top-0.5 w-3 h-3 bg-white rounded-full shadow-sm"></div>
          </div>
        </div>

        <div className="flex items-center justify-between mt-2">
          <label className="text-[11px] font-medium text-muted-foreground">Requires Serial Number</label>
          <div className="w-8 h-4 bg-black/10 dark:bg-white/10 rounded-full relative cursor-pointer">
            <div className="absolute left-0.5 top-0.5 w-3 h-3 bg-white rounded-full shadow-sm"></div>
          </div>
        </div>

      </div>
    </div>
  );
}
