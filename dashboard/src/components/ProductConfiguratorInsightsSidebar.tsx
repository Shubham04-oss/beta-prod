import { ChevronRight, Eye, Copy, FileText, Printer, Trash2, Plus } from "lucide-react";
import { PieChart, Pie, Cell, ResponsiveContainer } from "recharts";

interface ProductConfiguratorInsightsSidebarProps {
  title?: string;
  description?: string;
  category?: string;
  product?: any;
}

export function ProductConfiguratorInsightsSidebar({ title = "", description = "", category = "", product = null }: ProductConfiguratorInsightsSidebarProps) {
  
  // Calculate Data Quality dynamically
  let score = 0;
  if (title.trim().length > 0) score += 35;
  if (category.trim().length > 0) score += 25;
  if (description.trim().length > 10) score += 40; // Full score if desc > 10 chars, else proportional
  else if (description.trim().length > 0) score += 20;

  const dataQuality = [
    { name: "Complete", value: score, color: score > 80 ? "#22c55e" : score > 50 ? "#f59e0b" : "#ef4444" },
    { name: "Incomplete", value: 100 - score, color: "#27272a" },
  ];

  const qualityText = score > 80 ? "Excellent" : score > 50 ? "Good" : "Needs Work";
  const qualityColor = score > 80 ? "text-green-500" : score > 50 ? "text-amber-500" : "text-red-500";
  const dotColor = score > 80 ? "bg-green-500" : score > 50 ? "bg-amber-500" : "bg-red-500";

  const isDraft = !product;
  const statusText = isDraft ? "Draft" : product.status || "Active";
  const statusColor = isDraft ? "text-zinc-400" : "text-green-500";
  const statusDot = isDraft ? "bg-zinc-500" : "bg-green-500";

  const formatDate = (dateString?: string) => {
    if (!dateString) return "Just now";
    return new Date(dateString).toLocaleDateString("en-US", { year: 'numeric', month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' });
  };

  return (
    <div className="flex flex-col h-full w-full">
      
      <div className="flex-1 overflow-y-auto pr-2 custom-scrollbar flex flex-col gap-8 pb-10">
        
        {/* Product Status */}
        <div>
          <h4 className="font-semibold text-[13px] text-zinc-100 mb-4">Product Status</h4>
          <div className="flex flex-col gap-4">
            
            <div className="flex items-center justify-between">
              <span className="text-[11px] text-zinc-400">Overall Status</span>
              <div className="flex items-center gap-1.5">
                <div className={`w-1.5 h-1.5 rounded-full ${statusDot}`}></div>
                <span className={`text-[11px] font-semibold ${statusColor}`}>{statusText}</span>
              </div>
            </div>

            <div className="flex items-center justify-between">
              <span className="text-[11px] text-zinc-400">Data Quality Score</span>
              <div className="flex items-center gap-3">
                <div className="w-10 h-10 relative flex-shrink-0">
                  <ResponsiveContainer width="100%" height="100%">
                    <PieChart>
                      <Pie
                        data={dataQuality}
                        innerRadius={16}
                        outerRadius={20}
                        paddingAngle={0}
                        dataKey="value"
                        stroke="none"
                        startAngle={90}
                        endAngle={-270}
                        isAnimationActive={true}
                      >
                        {dataQuality.map((entry, index) => (
                          <Cell key={`cell-${index}`} fill={entry.color} />
                        ))}
                      </Pie>
                    </PieChart>
                  </ResponsiveContainer>
                </div>
                <div className="flex flex-col text-right">
                  <span className="text-sm font-bold text-white">{score}%</span>
                  <span className={`text-[9px] font-medium ${qualityColor}`}>{qualityText}</span>
                </div>
              </div>
            </div>

            <div className="flex items-center justify-between">
              <span className="text-[11px] text-zinc-400">Last Updated</span>
              <span className="text-[11px] text-zinc-300">{formatDate(product?.updatedAt)}</span>
            </div>

            <div className="flex items-center justify-between">
              <span className="text-[11px] text-zinc-400">Created At</span>
              <span className="text-[11px] text-zinc-300">{formatDate(product?.createdAt)}</span>
            </div>

            <button className="w-full py-2 mt-2 text-[11px] font-medium text-white bg-zinc-900 border border-zinc-800 rounded-xl hover:bg-zinc-800 transition-colors">
              View Change History
            </button>
          </div>
        </div>

        {/* Media Gallery */}
        <div>
          <div className="flex items-center justify-between mb-4">
            <h4 className="font-semibold text-[13px] text-zinc-100">Media Gallery</h4>
            <button className="text-[10px] text-zinc-400 hover:text-white transition-colors">Manage</button>
          </div>
          <div className="grid grid-cols-3 gap-2">
            {/* Images */}
            <div className="aspect-square bg-zinc-900 rounded-xl border border-zinc-800 flex items-center justify-center p-2">
               {isDraft ? <span className="text-[9px] text-zinc-600">None</span> : <img src="https://cdn.jsdelivr.net/gh/devicons/devicon@latest/icons/apple/apple-original.svg" className="w-full h-full object-contain opacity-40 grayscale" />}
            </div>
            <div className="aspect-square bg-zinc-900 rounded-xl border border-zinc-800 flex items-center justify-center p-2">
               {isDraft ? <span className="text-[9px] text-zinc-600">None</span> : <img src="https://cdn.jsdelivr.net/gh/devicons/devicon@latest/icons/apple/apple-original.svg" className="w-full h-full object-contain opacity-40 grayscale" />}
            </div>
            <div className="aspect-square bg-zinc-900 rounded-xl border border-zinc-800 flex items-center justify-center p-2">
               {isDraft ? <span className="text-[9px] text-zinc-600">None</span> : <img src="https://cdn.jsdelivr.net/gh/devicons/devicon@latest/icons/apple/apple-original.svg" className="w-full h-full object-contain opacity-40 grayscale" />}
            </div>
            <div className="aspect-square bg-zinc-900 rounded-xl border border-zinc-800 flex items-center justify-center overflow-hidden">
               {/* Simulating lifestyle image with gradient */}
               <div className="w-full h-full bg-gradient-to-br from-zinc-700 to-zinc-900 opacity-50"></div>
            </div>
            <div className="aspect-square bg-zinc-900 rounded-xl border border-zinc-800 flex items-center justify-center overflow-hidden">
               {/* Simulating box image */}
               <div className="w-full h-full bg-zinc-800/50"></div>
            </div>
            
            {/* Add Media */}
            <button className="aspect-square bg-zinc-900/50 rounded-xl border border-zinc-800 border-dashed flex flex-col items-center justify-center gap-1 hover:bg-zinc-800 transition-colors text-zinc-400 hover:text-white">
               <Plus className="w-4 h-4" />
               <span className="text-[9px]">Add</span>
            </button>
          </div>
        </div>

        {/* Quick Actions */}
        <div className={isDraft ? "opacity-50 pointer-events-none" : ""}>
          <h4 className="font-semibold text-[13px] text-zinc-100 mb-3">Quick Actions</h4>
          <div className="flex flex-col gap-1">
            <button className="flex items-center justify-between w-full py-2.5 group">
              <div className="flex items-center gap-3 text-zinc-400 group-hover:text-white transition-colors">
                <Eye className="w-3.5 h-3.5" />
                <span className="text-[11px] font-medium">Preview Product</span>
              </div>
              <ChevronRight className="w-3 h-3 text-zinc-600 group-hover:text-zinc-400" />
            </button>
            <button className="flex items-center justify-between w-full py-2.5 group">
              <div className="flex items-center gap-3 text-zinc-400 group-hover:text-white transition-colors">
                <Copy className="w-3.5 h-3.5" />
                <span className="text-[11px] font-medium">Duplicate Product</span>
              </div>
              <ChevronRight className="w-3 h-3 text-zinc-600 group-hover:text-zinc-400" />
            </button>
            <button className="flex items-center justify-between w-full py-2.5 group">
              <div className="flex items-center gap-3 text-zinc-400 group-hover:text-white transition-colors">
                <FileText className="w-3.5 h-3.5" />
                <span className="text-[11px] font-medium">Generate Product Sheet</span>
              </div>
              <ChevronRight className="w-3 h-3 text-zinc-600 group-hover:text-zinc-400" />
            </button>
            <button className="flex items-center justify-between w-full py-2.5 group">
              <div className="flex items-center gap-3 text-zinc-400 group-hover:text-white transition-colors">
                <Printer className="w-3.5 h-3.5" />
                <span className="text-[11px] font-medium">Print Labels</span>
              </div>
              <ChevronRight className="w-3 h-3 text-zinc-600 group-hover:text-zinc-400" />
            </button>
            <div className="h-px w-full bg-zinc-800/50 my-1"></div>
            <button className="flex items-center justify-between w-full py-2.5 group">
              <div className="flex items-center gap-3 text-red-500/80 group-hover:text-red-500 transition-colors">
                <Trash2 className="w-3.5 h-3.5" />
                <span className="text-[11px] font-medium">Archive Product</span>
              </div>
            </button>
          </div>
        </div>

      </div>
    </div>
  );
}
