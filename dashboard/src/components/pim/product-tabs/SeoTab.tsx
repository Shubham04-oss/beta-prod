export function SeoTab() {
  return (
    <div className="flex flex-col gap-6 pb-10 animate-in fade-in duration-200">
      <div>
        <h3 className="font-semibold text-[15px]">Search Engine Optimization</h3>
        <p className="text-[13px] text-muted-foreground mt-0.5">Control how this product appears in search engine results.</p>
      </div>

      <div className="flex flex-col gap-1.5 max-w-2xl">
        <label className="text-[11px] font-medium text-muted-foreground">Page Title</label>
        <input type="text" placeholder="e.g. Buy Premium Wireless Headphones" className="w-full h-10 px-3 rounded-xl bg-white/40 dark:bg-black/20 border border-black/5 dark:border-white/10 text-[13px] focus:outline-none focus:ring-2 focus:ring-primary/50" />
        <p className="text-[10px] text-muted-foreground text-right">0 of 70 characters used</p>
      </div>

      <div className="flex flex-col gap-1.5 max-w-2xl">
        <label className="text-[11px] font-medium text-muted-foreground">Meta Description</label>
        <textarea rows={4} placeholder="Write a compelling description to get people to click." className="w-full py-2 px-3 rounded-xl bg-white/40 dark:bg-black/20 border border-black/5 dark:border-white/10 text-[13px] focus:outline-none focus:ring-2 focus:ring-primary/50 resize-none"></textarea>
        <p className="text-[10px] text-muted-foreground text-right">0 of 320 characters used</p>
      </div>

      <div className="flex flex-col gap-1.5 max-w-2xl">
        <label className="text-[11px] font-medium text-muted-foreground">URL Handle</label>
        <div className="flex items-center">
          <span className="h-10 px-3 flex items-center bg-black/5 dark:bg-white/5 border border-r-0 border-black/5 dark:border-white/10 rounded-l-xl text-[13px] text-muted-foreground">https://store.com/products/</span>
          <input type="text" placeholder="premium-wireless-headphones" className="flex-1 h-10 px-3 rounded-r-xl bg-white/40 dark:bg-black/20 border border-black/5 dark:border-white/10 text-[13px] focus:outline-none focus:ring-2 focus:ring-primary/50" />
        </div>
      </div>

      <div className="flex flex-col gap-1.5 max-w-2xl">
        <label className="text-[11px] font-medium text-muted-foreground">Keywords</label>
        <div className="w-full min-h-[40px] p-1.5 rounded-xl bg-white/40 dark:bg-black/20 border border-black/5 dark:border-white/10 flex flex-wrap gap-1.5">
          <input type="text" placeholder="Add a keyword and press enter" className="flex-1 min-w-[150px] bg-transparent text-[13px] px-2 focus:outline-none" />
        </div>
      </div>
    </div>
  );
}
