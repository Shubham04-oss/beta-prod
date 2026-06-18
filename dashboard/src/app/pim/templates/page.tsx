"use client";

import { useState } from "react";
import { PIMInsightsSidebar } from "@/components/PIMInsightsSidebar";
import { PIMLeftSidebar } from "@/components/PIMLeftSidebar";
import { Search, Plus, FileText } from "lucide-react";
import { Button } from "@/components/ui/button";
import { useGetPIMStats } from "@/hooks/useProducts";
import useSWR from "swr";
import { fetchAPI, apiClient } from "@/lib/api";

export default function ProductTemplatesPage() {
  const { data: stats } = useGetPIMStats();
  const [searchQuery, setSearchQuery] = useState("");
  const [isCreating, setIsCreating] = useState(false);
  const [newName, setNewName] = useState("");
  const [newDescription, setNewDescription] = useState("");

  const { data: templates, mutate } = useSWR("/api/v1/pim/templates", fetchAPI);

  const handleCreate = async () => {
    try {
      await apiClient.post("api/v1/pim/templates", {
        json: { name: newName, description: newDescription }
      });
      mutate();
      setIsCreating(false);
      setNewName("");
      setNewDescription("");
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
              <FileText className="w-6 h-6 text-primary" />
              Product Templates
            </h1>
            <p className="text-sm text-muted-foreground mt-1">Manage reusable schema templates for your products.</p>
          </div>
          <div className="flex items-center gap-3">
            <div className="relative">
              <Search className="w-4 h-4 absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground" />
              <input 
                type="text"
                placeholder="Search templates..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="w-64 h-10 pl-9 pr-4 rounded-xl bg-black/5 dark:bg-white/5 border-none text-sm focus:outline-none focus:ring-2 focus:ring-primary/20 transition-all"
              />
            </div>
            <Button onClick={() => setIsCreating(true)} className="rounded-full shadow-sm bg-black text-white hover:bg-zinc-800 dark:bg-white dark:text-black dark:hover:bg-zinc-200 gap-2">
              <Plus className="w-4 h-4" />
              Add Template
            </Button>
          </div>
        </div>

        {/* Content Area */}
        <div className="flex-1 overflow-y-auto px-8 pb-8 custom-scrollbar">
          <div className="bg-white dark:bg-[#121212] rounded-3xl border border-black/5 dark:border-white/5 shadow-sm overflow-hidden min-h-[400px]">
            {(!templates || (templates as any[]).length === 0) ? (
              <div className="h-[400px] flex flex-col items-center justify-center text-center">
                <div className="w-16 h-16 rounded-full bg-black/5 dark:bg-white/5 flex items-center justify-center mb-4">
                  <FileText className="w-8 h-8 text-muted-foreground" />
                </div>
                <h3 className="text-lg font-semibold text-foreground">No templates found</h3>
                <p className="text-sm text-muted-foreground max-w-sm mt-2 mb-6">Create your first product template to enforce schema rules across your catalog.</p>
                <Button onClick={() => setIsCreating(true)} className="rounded-full bg-black text-white hover:bg-zinc-800 dark:bg-white dark:text-black dark:hover:bg-zinc-200">
                  <Plus className="w-4 h-4 mr-2" />
                  Add Template
                </Button>
              </div>
            ) : (
              <table className="w-full text-left border-collapse">
                <thead>
                  <tr className="border-b border-black/5 dark:border-white/5">
                    <th className="py-4 px-6 text-xs font-semibold text-muted-foreground uppercase tracking-wider">Template Name</th>
                    <th className="py-4 px-6 text-xs font-semibold text-muted-foreground uppercase tracking-wider">Description</th>
                    <th className="py-4 px-6 text-xs font-semibold text-muted-foreground uppercase tracking-wider">Created</th>
                    <th className="py-4 px-6 text-xs font-semibold text-muted-foreground uppercase tracking-wider text-right">Actions</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-black/5 dark:divide-white/5">
                  {(templates as any[]).map((tpl: any) => (
                    <tr key={tpl.id} className="hover:bg-black/5 dark:hover:bg-white/5 transition-colors group">
                      <td className="py-4 px-6">
                        <span className="font-medium text-sm text-foreground">{tpl.name}</span>
                      </td>
                      <td className="py-4 px-6">
                        <span className="text-sm text-muted-foreground">{tpl.description || '-'}</span>
                      </td>
                      <td className="py-4 px-6">
                        <span className="text-sm text-muted-foreground">{new Date(tpl.created_at).toLocaleDateString()}</span>
                      </td>
                      <td className="py-4 px-6 text-right">
                        <Button size="sm" variant="ghost" className="h-8 text-xs rounded-lg opacity-0 group-hover:opacity-100 transition-opacity">Edit</Button>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            )}
          </div>
        </div>
      </main>

      <aside className="w-[360px] flex-shrink-0 bg-[#121212] text-white p-6 rounded-3xl relative z-20 flex flex-col h-full shadow-[0_8px_30px_rgb(0,0,0,0.4)] border border-white/5">
        <PIMInsightsSidebar stats={stats} />
      </aside>

      {isCreating && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-md transition-all">
          <div className="w-[440px] bg-white/90 dark:bg-[#121212]/95 backdrop-blur-xl rounded-3xl p-8 shadow-[0_24px_48px_rgb(0,0,0,0.3)] border border-white/20 dark:border-white/10 flex flex-col gap-8 animate-in fade-in slide-in-from-bottom-4 zoom-in-95 duration-300">
            <div>
              <h2 className="text-2xl font-bold tracking-tight bg-gradient-to-br from-black to-black/70 dark:from-white dark:to-white/70 bg-clip-text text-transparent">New Template</h2>
              <p className="text-sm text-muted-foreground mt-1.5">Create a schema template for products.</p>
            </div>

            <div className="flex flex-col gap-5">
              <div className="flex flex-col gap-2">
                <label className="text-xs font-semibold text-foreground/80 tracking-wide">Template Name <span className="text-red-500">*</span></label>
                <input 
                  type="text" 
                  value={newName}
                  onChange={(e) => setNewName(e.target.value)}
                  placeholder="e.g. Apparel, Electronics"
                  className="w-full h-11 px-4 rounded-xl bg-black/5 dark:bg-white/5 border border-transparent dark:border-white/5 text-sm placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-black/10 dark:focus:ring-white/20 focus:bg-white dark:focus:bg-black transition-all" 
                />
              </div>
              <div className="flex flex-col gap-2">
                <label className="text-xs font-semibold text-foreground/80 tracking-wide">Description</label>
                <textarea 
                  value={newDescription}
                  onChange={(e) => setNewDescription(e.target.value)}
                  placeholder="What products is this template for?"
                  className="w-full min-h-[100px] p-4 rounded-xl bg-black/5 dark:bg-white/5 border border-transparent dark:border-white/5 text-sm placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-black/10 dark:focus:ring-white/20 focus:bg-white dark:focus:bg-black transition-all resize-y" 
                />
              </div>
            </div>

            <div className="flex items-center justify-end gap-3 mt-2">
              <Button onClick={() => setIsCreating(false)} variant="ghost" className="h-10 px-5 text-sm font-medium hover:bg-black/5 dark:hover:bg-white/5 rounded-xl transition-colors">Cancel</Button>
              <Button onClick={handleCreate} disabled={!newName} className="h-10 px-6 rounded-xl font-medium bg-gradient-to-b from-black to-zinc-800 text-white shadow-lg hover:shadow-xl hover:scale-[1.02] dark:from-white dark:to-zinc-200 dark:text-black transition-all active:scale-95 disabled:opacity-50 disabled:pointer-events-none">
                Create Template
              </Button>
            </div>
          </div>
        </div>
      )}
    </>
  );
}
