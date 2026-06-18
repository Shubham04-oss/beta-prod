"use client";

import { useState } from "react";
import { PIMInsightsSidebar } from "@/components/PIMInsightsSidebar";
import { PIMLeftSidebar } from "@/components/PIMLeftSidebar";
import { Search, Plus, Image as ImageIcon, UploadCloud } from "lucide-react";
import { Button } from "@/components/ui/button";
import { useGetPIMStats } from "@/hooks/useProducts";
import useSWR from "swr";
import { fetchAPI, apiClient } from "@/lib/api";

export default function MediaLibraryPage() {
  const { data: stats } = useGetPIMStats();
  const [searchQuery, setSearchQuery] = useState("");
  const [isUploading, setIsUploading] = useState(false);
  const [newUrl, setNewUrl] = useState("");
  const [newAltText, setNewAltText] = useState("");

  const { data: media, mutate } = useSWR("/api/v1/pim/media", fetchAPI);

  const handleUpload = async () => {
    try {
      await apiClient.post("api/v1/pim/media", {
        json: { 
          url: newUrl, 
          alt_text: newAltText,
          sort_order: 0
        }
      });
      mutate();
      setIsUploading(false);
      setNewUrl("");
      setNewAltText("");
    } catch (e) {
      console.error(e);
    }
  };

  return (
    <>
      <PIMLeftSidebar />
      <main className="flex-1 flex flex-col h-full relative z-10 overflow-hidden pr-4">
        {/* Header */}
        <div className="flex items-center justify-between px-8 py-6 flex-shrink-0">
          <div>
            <h1 className="text-2xl font-bold text-foreground flex items-center gap-2">
              <ImageIcon className="w-6 h-6 text-primary" />
              Media Library
            </h1>
            <p className="text-sm text-muted-foreground mt-1">Manage global assets, images, and documents.</p>
          </div>
          <div className="flex items-center gap-3">
            <div className="relative">
              <Search className="w-4 h-4 absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground" />
              <input 
                type="text"
                placeholder="Search assets..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="w-64 h-10 pl-9 pr-4 rounded-xl bg-black/5 dark:bg-white/5 border-none text-sm focus:outline-none focus:ring-2 focus:ring-primary/20 transition-all"
              />
            </div>
            <Button onClick={() => setIsUploading(true)} className="rounded-full shadow-sm bg-black text-white hover:bg-zinc-800 dark:bg-white dark:text-black dark:hover:bg-zinc-200 gap-2">
              <UploadCloud className="w-4 h-4" />
              Upload Assets
            </Button>
          </div>
        </div>

        {/* Content Area */}
        <div className="flex-1 overflow-y-auto px-8 pb-8 custom-scrollbar">
          {(!media || (media as any[]).length === 0) ? (
            <div className="bg-white dark:bg-[#121212] rounded-3xl border border-black/5 dark:border-white/5 shadow-sm h-[400px] flex flex-col items-center justify-center text-center">
              <div className="w-16 h-16 rounded-full bg-black/5 dark:bg-white/5 flex items-center justify-center mb-4">
                <ImageIcon className="w-8 h-8 text-muted-foreground" />
              </div>
              <h3 className="text-lg font-semibold text-foreground">No media assets found</h3>
              <p className="text-sm text-muted-foreground max-w-sm mt-2 mb-6">Upload images, videos, or documents to associate them with your products and variants.</p>
              <Button onClick={() => setIsUploading(true)} className="rounded-full bg-black text-white hover:bg-zinc-800 dark:bg-white dark:text-black dark:hover:bg-zinc-200">
                <UploadCloud className="w-4 h-4 mr-2" />
                Upload Assets
              </Button>
            </div>
          ) : (
            <div className="grid grid-cols-2 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6 gap-4">
              {(media as any[]).map((asset: any) => (
                <div key={asset.id} className="group relative aspect-square bg-black/5 dark:bg-white/5 rounded-2xl overflow-hidden border border-black/5 dark:border-white/10 hover:border-black/20 dark:hover:border-white/30 transition-all">
                  <img src={asset.url} alt={asset.alt_text} className="w-full h-full object-cover" />
                  <div className="absolute inset-0 bg-black/60 opacity-0 group-hover:opacity-100 transition-opacity flex flex-col justify-end p-3">
                    <p className="text-xs font-medium text-white truncate">{asset.alt_text || 'Asset'}</p>
                    <p className="text-[10px] text-white/70">{new Date(asset.created_at).toLocaleDateString()}</p>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      </main>

      <aside className="w-[360px] flex-shrink-0 bg-[#121212] text-white p-6 rounded-3xl relative z-20 flex flex-col h-full shadow-[0_8px_30px_rgb(0,0,0,0.4)] border border-white/5">
        <PIMInsightsSidebar stats={stats} />
      </aside>

      {isUploading && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-md transition-all">
          <div className="w-[440px] bg-white/90 dark:bg-[#121212]/95 backdrop-blur-xl rounded-3xl p-8 shadow-[0_24px_48px_rgb(0,0,0,0.3)] border border-white/20 dark:border-white/10 flex flex-col gap-8 animate-in fade-in slide-in-from-bottom-4 zoom-in-95 duration-300">
            <div>
              <h2 className="text-2xl font-bold tracking-tight bg-gradient-to-br from-black to-black/70 dark:from-white dark:to-white/70 bg-clip-text text-transparent">Upload Asset</h2>
              <p className="text-sm text-muted-foreground mt-1.5">Add a new image to the media library.</p>
            </div>

            <div className="flex flex-col gap-5">
              <div className="flex flex-col gap-2">
                <label className="text-xs font-semibold text-foreground/80 tracking-wide">Image URL <span className="text-red-500">*</span></label>
                <input 
                  type="url" 
                  value={newUrl}
                  onChange={(e) => setNewUrl(e.target.value)}
                  placeholder="https://example.com/image.png"
                  className="w-full h-11 px-4 rounded-xl bg-black/5 dark:bg-white/5 border border-transparent dark:border-white/5 text-sm placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-black/10 dark:focus:ring-white/20 focus:bg-white dark:focus:bg-black transition-all" 
                />
              </div>
              <div className="flex flex-col gap-2">
                <label className="text-xs font-semibold text-foreground/80 tracking-wide">Alt Text</label>
                <input 
                  type="text" 
                  value={newAltText}
                  onChange={(e) => setNewAltText(e.target.value)}
                  placeholder="Describe the image for SEO..."
                  className="w-full h-11 px-4 rounded-xl bg-black/5 dark:bg-white/5 border border-transparent dark:border-white/5 text-sm placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-black/10 dark:focus:ring-white/20 focus:bg-white dark:focus:bg-black transition-all" 
                />
              </div>
            </div>

            <div className="flex items-center justify-end gap-3 mt-2">
              <Button onClick={() => setIsUploading(false)} variant="ghost" className="h-10 px-5 text-sm font-medium hover:bg-black/5 dark:hover:bg-white/5 rounded-xl transition-colors">Cancel</Button>
              <Button onClick={handleUpload} disabled={!newUrl} className="h-10 px-6 rounded-xl font-medium bg-gradient-to-b from-black to-zinc-800 text-white shadow-lg hover:shadow-xl hover:scale-[1.02] dark:from-white dark:to-zinc-200 dark:text-black transition-all active:scale-95 disabled:opacity-50 disabled:pointer-events-none">
                Upload Asset
              </Button>
            </div>
          </div>
        </div>
      )}
    </>
  );
}
