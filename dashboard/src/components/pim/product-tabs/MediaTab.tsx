import { UploadCloud, Image as ImageIcon } from "lucide-react";

export function MediaTab() {
  return (
    <div className="flex flex-col gap-6 pb-10 animate-in fade-in duration-200">
      <div className="flex items-center justify-between">
        <h3 className="font-semibold text-[15px]">Media Gallery</h3>
        <span className="text-[11px] text-muted-foreground">3 images uploaded</span>
      </div>
      
      {/* Dropzone */}
      <div className="w-full h-40 border-2 border-dashed border-black/10 dark:border-white/10 rounded-2xl bg-black/5 dark:bg-white/5 flex flex-col items-center justify-center gap-3 hover:bg-black/10 dark:hover:bg-white/10 transition-colors cursor-pointer">
        <div className="w-10 h-10 rounded-full bg-white dark:bg-black flex items-center justify-center shadow-sm">
          <UploadCloud className="w-5 h-5 text-primary" />
        </div>
        <div className="text-center">
          <p className="text-[13px] font-medium">Click to upload or drag and drop</p>
          <p className="text-[11px] text-muted-foreground mt-0.5">SVG, PNG, JPG or GIF (max. 10MB)</p>
        </div>
      </div>

      {/* Grid */}
      <div className="grid grid-cols-4 gap-4 mt-2">
        <div className="aspect-square rounded-xl bg-white/40 dark:bg-black/20 border border-black/5 dark:border-white/10 flex flex-col items-center justify-center relative overflow-hidden group">
          <img src="https://cdn.jsdelivr.net/gh/devicons/devicon@latest/icons/apple/apple-original.svg" className="w-16 h-16 object-contain opacity-40 grayscale" />
          <div className="absolute top-2 left-2 px-1.5 py-0.5 bg-black/50 backdrop-blur-md text-white text-[9px] font-bold rounded-sm">MAIN</div>
        </div>
        <div className="aspect-square rounded-xl bg-white/40 dark:bg-black/20 border border-black/5 dark:border-white/10 flex items-center justify-center overflow-hidden">
          <img src="https://cdn.jsdelivr.net/gh/devicons/devicon@latest/icons/apple/apple-original.svg" className="w-16 h-16 object-contain opacity-20 grayscale" />
        </div>
        <div className="aspect-square rounded-xl bg-white/40 dark:bg-black/20 border border-black/5 dark:border-white/10 flex items-center justify-center overflow-hidden">
          <img src="https://cdn.jsdelivr.net/gh/devicons/devicon@latest/icons/apple/apple-original.svg" className="w-16 h-16 object-contain opacity-20 grayscale" />
        </div>
        <div className="aspect-square rounded-xl border border-dashed border-black/10 dark:border-white/10 flex flex-col items-center justify-center gap-2 hover:bg-black/5 dark:hover:bg-white/5 transition-colors cursor-pointer text-muted-foreground">
          <ImageIcon className="w-6 h-6" />
          <span className="text-[10px] font-medium">Add via URL</span>
        </div>
      </div>
    </div>
  );
}
