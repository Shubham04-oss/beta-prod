"use client";

import { useState } from "react";
import { PIMLeftSidebar } from "@/components/PIMLeftSidebar";
import { PIMInsightsSidebar } from "@/components/PIMInsightsSidebar";
import { Search, Plus, FolderTree } from "lucide-react";
import { Button } from "@/components/ui/button";
import { useCategories } from "@/hooks/useCategories";
import { useGetPIMStats } from "@/hooks/useProducts";

export default function CategoriesPage() {
  const { categories, isLoading, createCategory } = useCategories();
  const { data: stats } = useGetPIMStats(); // For the right sidebar stats
  const [searchQuery, setSearchQuery] = useState("");
  const [isCreating, setIsCreating] = useState(false);
  
  const [newName, setNewName] = useState("");
  const [newSlug, setNewSlug] = useState("");
  const [newDescription, setNewDescription] = useState("");
  const [newParentId, setNewParentId] = useState("");

  const filteredCategories = categories?.filter(cat => 
    cat.name.toLowerCase().includes(searchQuery.toLowerCase())
  );

  const handleCreate = async () => {
    if (!newName || !newSlug) return;
    try {
      await createCategory({ 
        name: newName, 
        slug: newSlug, 
        description: newDescription, 
        parent_id: newParentId || null 
      });
      setIsCreating(false);
      setNewName("");
      setNewSlug("");
      setNewDescription("");
      setNewParentId("");
    } catch (e) {
      alert("Failed to create category");
    }
  };

  return (
    <>
      <PIMLeftSidebar />

      <main className="flex-1 overflow-y-auto custom-scrollbar pr-4 flex flex-col gap-6">
        
        {/* Header */}
        <div className="flex items-center justify-between mt-2">
          <div>
            <h1 className="text-3xl font-bold tracking-tight">Categories Management</h1>
            <p className="text-muted-foreground mt-1">Organize products into hierarchical taxonomy.</p>
          </div>
          <Button onClick={() => setIsCreating(true)} className="rounded-full shadow-sm bg-black text-white hover:bg-zinc-800 dark:bg-white dark:text-black dark:hover:bg-zinc-200">
            <Plus className="w-4 h-4 mr-2" /> Add Category
          </Button>
        </div>

        {/* Categories Table Container */}
        <div className="bg-white/20 dark:bg-white/5 rounded-[2rem] p-6 border border-white/30 dark:border-white/10 flex flex-col flex-1">
          
          <div className="flex items-center gap-6 mb-6 px-2">
            <div className="flex items-center gap-3 w-full">
              <div className="relative w-64">
                <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
                <input 
                  type="text" 
                  placeholder="Search categories..." 
                  value={searchQuery}
                  onChange={e => setSearchQuery(e.target.value)}
                  className="w-full h-9 pl-9 pr-4 rounded-full bg-white/40 dark:bg-black/20 border border-black/5 dark:border-white/10 text-xs focus:outline-none focus:ring-2 focus:ring-primary/50"
                />
              </div>
            </div>
          </div>

          <div className="w-full overflow-x-auto">
            {isLoading ? (
              <p className="text-sm text-muted-foreground p-4">Loading categories...</p>
            ) : filteredCategories?.length === 0 ? (
              <p className="text-sm text-muted-foreground p-4">No categories found.</p>
            ) : (
              <table className="w-full text-sm text-left">
                <thead>
                  <tr className="text-[11px] text-muted-foreground uppercase tracking-wider border-b border-black/5 dark:border-white/5">
                    <th className="pb-3 font-semibold px-2">Category Name</th>
                    <th className="pb-3 font-semibold px-2">Slug</th>
                    <th className="pb-3 font-semibold px-2">Parent</th>
                    <th className="pb-3 font-semibold px-2">Products</th>
                    <th className="pb-3 font-semibold px-2 text-right">Actions</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-black/5 dark:divide-white/5">
                  {filteredCategories?.map((cat) => {
                    const parent = categories?.find(c => c.id === cat.parent_id);
                    return (
                      <tr key={cat.id} className="group hover:bg-black/[0.02] dark:hover:bg-white/[0.02] transition-colors">
                        <td className="py-3 px-2">
                          <div className="flex items-center gap-3">
                            <div className="w-8 h-8 rounded-md bg-black/5 dark:bg-white/5 flex items-center justify-center border border-black/5 dark:border-white/5">
                              <FolderTree className="w-4 h-4 text-muted-foreground" />
                            </div>
                            <span className="font-medium text-sm text-foreground">{cat.name}</span>
                          </div>
                        </td>
                        <td className="py-3 px-2">
                          <span className="text-xs px-2 py-1 bg-black/5 dark:bg-white/5 rounded-md font-medium text-muted-foreground">{cat.slug}</span>
                        </td>
                        <td className="py-3 px-2">
                          {parent ? (
                            <span className="text-xs text-muted-foreground">{parent.name}</span>
                          ) : (
                            <span className="text-xs italic text-muted-foreground/50">Root</span>
                          )}
                        </td>
                        <td className="py-3 px-2">
                          <span className="text-xs px-2 py-1 bg-black/5 dark:bg-white/5 rounded-md font-medium">0</span>
                        </td>
                        <td className="py-3 px-2 text-right">
                          <Button size="sm" variant="ghost" className="h-7 text-[10px] rounded-md">Edit</Button>
                        </td>
                      </tr>
                    );
                  })}
                </tbody>
              </table>
            )}
          </div>
        </div>
      </main>

      {/* Right Matte Black Area */}
      <aside className="w-[360px] flex-shrink-0 bg-[#121212] text-white p-6 rounded-3xl relative z-20 flex flex-col h-full shadow-[0_8px_30px_rgb(0,0,0,0.4)] border border-white/5">
        <PIMInsightsSidebar stats={stats} />
      </aside>

      {/* Create Modal Overlay */}
      {isCreating && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-md transition-all">
          <div className="w-[440px] bg-white/90 dark:bg-[#121212]/95 backdrop-blur-xl rounded-3xl p-8 shadow-[0_24px_48px_rgb(0,0,0,0.3)] border border-white/20 dark:border-white/10 flex flex-col gap-8 animate-in fade-in slide-in-from-bottom-4 zoom-in-95 duration-300">
            <div>
              <h2 className="text-2xl font-bold tracking-tight bg-gradient-to-br from-black to-black/70 dark:from-white dark:to-white/70 bg-clip-text text-transparent">New Category</h2>
              <p className="text-sm text-muted-foreground mt-1.5">Add a new category to your taxonomy tree.</p>
            </div>

            <div className="flex flex-col gap-5">
              <div className="flex flex-col gap-2">
                <label className="text-xs font-semibold text-foreground/80 tracking-wide">Category Name <span className="text-red-500">*</span></label>
                <input 
                  type="text" 
                  value={newName}
                  onChange={(e) => {
                    setNewName(e.target.value);
                    if (!newSlug || newSlug === newName.toLowerCase().replace(/\s+/g, '-')) {
                      setNewSlug(e.target.value.toLowerCase().replace(/\s+/g, '-'));
                    }
                  }}
                  placeholder="e.g. Electronics, Clothing"
                  className="w-full h-11 px-4 rounded-xl bg-black/5 dark:bg-white/5 border border-transparent dark:border-white/5 text-sm placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-black/10 dark:focus:ring-white/20 focus:bg-white dark:focus:bg-black transition-all" 
                />
              </div>
              <div className="flex flex-col gap-2">
                <label className="text-xs font-semibold text-foreground/80 tracking-wide">Slug <span className="text-red-500">*</span></label>
                <input 
                  type="text" 
                  value={newSlug}
                  onChange={(e) => setNewSlug(e.target.value)}
                  placeholder="e.g. electronics"
                  className="w-full h-11 px-4 rounded-xl bg-black/5 dark:bg-white/5 border border-transparent dark:border-white/5 text-sm placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-black/10 dark:focus:ring-white/20 focus:bg-white dark:focus:bg-black transition-all" 
                />
              </div>
              <div className="flex flex-col gap-2">
                <label className="text-xs font-semibold text-foreground/80 tracking-wide">Parent Category</label>
                <select
                  value={newParentId}
                  onChange={(e) => setNewParentId(e.target.value)}
                  className="w-full h-11 px-4 rounded-xl bg-black/5 dark:bg-white/5 border border-transparent dark:border-white/5 text-sm focus:outline-none focus:ring-2 focus:ring-black/10 dark:focus:ring-white/20 focus:bg-white dark:focus:bg-black transition-all appearance-none"
                >
                  <option value="">None (Root Category)</option>
                  {categories?.map(cat => (
                    <option key={cat.id} value={cat.id}>{cat.name}</option>
                  ))}
                </select>
              </div>
              <div className="flex flex-col gap-2">
                <label className="text-xs font-semibold text-foreground/80 tracking-wide">Description</label>
                <textarea 
                  value={newDescription}
                  onChange={(e) => setNewDescription(e.target.value)}
                  placeholder="Optional details about this category..."
                  className="w-full min-h-[80px] p-4 rounded-xl bg-black/5 dark:bg-white/5 border border-transparent dark:border-white/5 text-sm placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-black/10 dark:focus:ring-white/20 focus:bg-white dark:focus:bg-black transition-all resize-y" 
                />
              </div>
            </div>

            <div className="flex items-center justify-end gap-3 mt-2">
              <Button onClick={() => setIsCreating(false)} variant="ghost" className="h-10 px-5 text-sm font-medium hover:bg-black/5 dark:hover:bg-white/5 rounded-xl transition-colors">Cancel</Button>
              <Button onClick={handleCreate} disabled={!newName || !newSlug} className="h-10 px-6 rounded-xl font-medium bg-gradient-to-b from-black to-zinc-800 text-white shadow-lg hover:shadow-xl hover:scale-[1.02] dark:from-white dark:to-zinc-200 dark:text-black transition-all active:scale-95 disabled:opacity-50 disabled:pointer-events-none">
                Create Category
              </Button>
            </div>
          </div>
        </div>
      )}
    </>
  );
}
