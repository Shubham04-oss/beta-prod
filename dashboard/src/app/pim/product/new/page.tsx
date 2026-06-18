"use client";

import { useRouter } from "next/navigation";
import { useState } from "react";
import { useCreateProduct } from "@/hooks/useProducts";
import { PIMLeftSidebar } from "@/components/PIMLeftSidebar";
import { ProductConfiguratorInsightsSidebar } from "@/components/ProductConfiguratorInsightsSidebar";
import { ArrowLeft, ChevronLeft, ChevronRight, ChevronDown, CheckCircle2, Copy, BarChart2, CalendarDays } from "lucide-react";
import { Button } from "@/components/ui/button";

import { GeneralTab } from "@/components/pim/product-tabs/GeneralTab";

export default function NewProductConfiguratorPage() {
  const router = useRouter();
  const createMutation = useCreateProduct();

  const [title, setTitle] = useState("");
  const [description, setDescription] = useState("");
  const [category, setCategory] = useState("");

  let score = 0;
  if (title.trim().length > 0) score += 35;
  if (category.trim().length > 0) score += 25;
  if (description.trim().length > 10) score += 40;
  else if (description.trim().length > 0) score += 20;

  const handleSave = async () => {
    if (!title) {
      alert("Please provide a title");
      return;
    }
    try {
      const newProduct = await createMutation.mutateAsync({ title, description, category });
      if (newProduct && newProduct.id) {
        router.push(`/pim/product/${newProduct.id}`);
      } else {
        router.push('/pim');
      }
    } catch (e) {
      alert("Failed to create product");
    }
  };

  return (
    <>
      <title>Create Product | SYNQ PIM</title>
      
      {/* Left Sidebar */}
      <PIMLeftSidebar />

      {/* Main Content Area */}
      <main className="flex-1 overflow-y-auto custom-scrollbar pr-4 flex flex-col gap-6">
        
        {/* Header */}
        <div className="flex items-center justify-between mt-2">
          <div>
            <h1 className="text-3xl font-bold tracking-tight">Create Product</h1>
            <p className="text-muted-foreground mt-1">Configure your new product before saving.</p>
          </div>
        </div>

        {/* Top Action Bar */}
        <div className="flex items-center justify-between mt-2">
          <button onClick={() => router.push('/pim')} className="flex items-center gap-2 text-xs font-medium text-muted-foreground hover:text-foreground transition-colors">
            <ArrowLeft className="w-3.5 h-3.5" /> Back to products
          </button>
          
          <div className="flex items-center gap-3">
            <div className="flex items-center gap-1 mr-2">
              <Button variant="outline" size="icon" className="w-8 h-8 rounded-full bg-white/40 dark:bg-black/20 border-white/40 dark:border-white/10 shadow-sm opacity-50 cursor-not-allowed">
                <ChevronLeft className="w-4 h-4" />
              </Button>
              <Button variant="outline" size="icon" className="w-8 h-8 rounded-full bg-white/40 dark:bg-black/20 border-white/40 dark:border-white/10 shadow-sm opacity-50 cursor-not-allowed">
                <ChevronRight className="w-4 h-4" />
              </Button>
            </div>
            <Button onClick={handleSave} disabled={createMutation.isPending} className="rounded-full shadow-sm bg-black text-white hover:bg-zinc-800 dark:bg-white dark:text-black dark:hover:bg-zinc-200">
              {createMutation.isPending ? "Creating..." : "Save Changes"}
            </Button>
            <Button disabled variant="outline" className="rounded-full bg-white/40 dark:bg-black/20 border-red-500/40 text-red-500 hover:bg-red-500/10 shadow-sm opacity-50">
              Delete Product
            </Button>
          </div>
        </div>

        {/* Product Highlight Card */}
        <div className="bg-white/20 dark:bg-white/5 rounded-[2rem] p-6 border border-white/30 dark:border-white/10 flex items-center gap-8">
          <div className="w-32 h-32 rounded-2xl bg-white p-2 border border-black/5 shadow-sm flex-shrink-0 flex items-center justify-center">
             <div className="text-muted-foreground text-[10px] text-center opacity-60">No Image</div>
          </div>
          
          <div className="flex flex-col flex-1 gap-4">
            <div>
              <div className="flex items-center gap-3">
                <h2 className="text-2xl font-bold">{title || "New Product Title"}</h2>
                <span className="px-2 py-0.5 bg-zinc-500/10 text-zinc-600 dark:text-zinc-400 text-[10px] font-bold rounded-md">DRAFT</span>
              </div>
              <p className="text-xs text-muted-foreground mt-1">ID: Pending</p>
              <p className="text-xs text-muted-foreground mt-0.5">{category || "No Category"}</p>
            </div>
            
            <div className="flex items-center gap-8 mt-2 opacity-50">
              <div className="flex items-center gap-3">
                <div className="w-8 h-8 rounded-full relative flex items-center justify-center">
                  <svg viewBox="0 0 36 36" className="w-full h-full absolute inset-0 transform -rotate-90">
                    <path className="text-black/5 dark:text-white/5" d="M18 2.0845 a 15.9155 15.9155 0 0 1 0 31.831 a 15.9155 15.9155 0 0 1 0 -31.831" fill="none" stroke="currentColor" strokeWidth="4" />
                    <path className={score > 80 ? "text-green-500" : score > 50 ? "text-amber-500" : "text-red-500"} strokeDasharray={`${score}, 100`} d="M18 2.0845 a 15.9155 15.9155 0 0 1 0 31.831 a 15.9155 15.9155 0 0 1 0 -31.831" fill="none" stroke="currentColor" strokeWidth="4" />
                  </svg>
                  <span className="text-[9px] font-bold">{score}%</span>
                </div>
                <div className="flex flex-col">
                  <span className="text-[10px] font-medium">Data Quality</span>
                </div>
              </div>
              
              <div className="flex items-center gap-3">
                <div className="w-8 h-8 rounded-full bg-zinc-500/10 flex items-center justify-center text-zinc-500">
                  <span className="text-sm font-bold">$</span>
                </div>
                <div className="flex flex-col">
                  <span className="text-[11px] font-bold">Draft</span>
                  <span className="text-[10px] text-muted-foreground">Status</span>
                </div>
              </div>
              
              <div className="flex items-center gap-3">
                <div className="w-8 h-8 rounded-full bg-black/5 dark:bg-white/5 flex items-center justify-center">
                  <Copy className="w-3.5 h-3.5" />
                </div>
                <div className="flex flex-col">
                  <span className="text-[11px] font-bold">0</span>
                  <span className="text-[10px] text-muted-foreground">Variants</span>
                </div>
              </div>

              <div className="flex items-center gap-3">
                <div className="w-8 h-8 rounded-full bg-black/5 dark:bg-white/5 flex items-center justify-center">
                  <BarChart2 className="w-3.5 h-3.5" />
                </div>
                <div className="flex flex-col">
                  <span className="text-[11px] font-bold">0</span>
                  <span className="text-[10px] text-muted-foreground">Channels</span>
                </div>
              </div>

              <div className="flex items-center gap-3">
                <div className="w-8 h-8 rounded-full bg-black/5 dark:bg-white/5 flex items-center justify-center">
                  <CalendarDays className="w-3.5 h-3.5" />
                </div>
                <div className="flex flex-col">
                  <span className="text-[11px] font-bold">Now</span>
                  <span className="text-[10px] text-muted-foreground">Last Updated</span>
                </div>
              </div>
            </div>
          </div>
        </div>

        {/* Tabs */}
        <div className="flex items-center gap-6 border-b border-black/5 dark:border-white/5 px-2">
          {[
            { id: 'general', label: 'General Information' },
            { id: 'attributes', label: 'Attributes' },
            { id: 'variants', label: 'Variants' },
            { id: 'media', label: 'Media' },
            { id: 'pricing', label: 'Pricing' },
            { id: 'inventory', label: 'Inventory' },
            { id: 'seo', label: 'SEO & Content' },
            { id: 'channels', label: 'Channels' },
            { id: 'history', label: 'History' }
          ].map(tab => (
            <button 
              key={tab.id}
              disabled={tab.id !== 'general'}
              className={`text-[13px] pb-3 transition-colors ${tab.id === 'general' ? 'font-semibold text-foreground border-b-2 border-primary' : 'font-medium text-muted-foreground opacity-50 cursor-not-allowed'}`}
              title={tab.id !== 'general' ? "Save to unlock" : ""}
            >
              {tab.label}
            </button>
          ))}
        </div>

        {/* Tab Content Container */}
        <div className="pt-2">
          <GeneralTab 
            title={title} setTitle={setTitle}
            description={description} setDescription={setDescription}
            category={category} setCategory={setCategory}
          />
        </div>

      </main>

      {/* Right Matte Black Area */}
      <aside className="w-[360px] flex-shrink-0 bg-[#121212] text-white p-6 rounded-3xl relative z-20 flex flex-col h-full shadow-[0_8px_30px_rgb(0,0,0,0.4)] border border-white/5">
        <ProductConfiguratorInsightsSidebar 
          title={title}
          description={description}
          category={category}
        />
      </aside>
    </>
  );
}
