"use client";

import { useState } from "react";
import { PIMInsightsSidebar } from "@/components/PIMInsightsSidebar";
import { PIMLeftSidebar } from "@/components/PIMLeftSidebar";
import { Upload, Download, FileSpreadsheet, Search, Clock, CheckCircle2, XCircle, Settings } from "lucide-react";
import { Button } from "@/components/ui/button";
import { useGetPIMStats } from "@/hooks/useProducts";
import useSWR from "swr";
import { fetchAPI, apiClient } from "@/lib/api";

export default function ImportExportPage() {
  const { data: stats } = useGetPIMStats();
  const [searchQuery, setSearchQuery] = useState("");
  const [isImporting, setIsImporting] = useState(false);
  const [mockFileName, setMockFileName] = useState("");

  // Reusing the bulk-jobs endpoint but passing IMPORT/EXPORT types. 
  // SWR automatically handles query params. Let's fetch IMPORT jobs here.
  const { data: jobs, mutate } = useSWR("/api/v1/pim/bulk-jobs?type=IMPORT", fetchAPI);

  const handleImport = async () => {
    try {
      await apiClient.post("api/v1/pim/bulk-jobs", {
        json: { 
          job_type: "IMPORT",
          payload: Buffer.from(JSON.stringify({ file_name: mockFileName, source: "CSV_UPLOAD" })).toString('base64')
        }
      });
      mutate();
      setIsImporting(false);
      setMockFileName("");
    } catch (e) {
      console.error(e);
    }
  };

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'COMPLETED': return <CheckCircle2 className="w-4 h-4 text-green-500" />;
      case 'FAILED': return <XCircle className="w-4 h-4 text-red-500" />;
      case 'RUNNING': return <Settings className="w-4 h-4 text-blue-500 animate-spin" />;
      default: return <Clock className="w-4 h-4 text-amber-500" />;
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
              <FileSpreadsheet className="w-6 h-6 text-primary" />
              Import & Export
            </h1>
            <p className="text-sm text-muted-foreground mt-1">Move large volumes of catalog data in and out of SYNQ.</p>
          </div>
          <div className="flex items-center gap-3">
            <div className="relative">
              <Search className="w-4 h-4 absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground" />
              <input 
                type="text"
                placeholder="Search history..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="w-64 h-10 pl-9 pr-4 rounded-xl bg-black/5 dark:bg-white/5 border-none text-sm focus:outline-none focus:ring-2 focus:ring-primary/20 transition-all"
              />
            </div>
            <Button variant="outline" className="rounded-full shadow-sm bg-transparent border-black/10 dark:border-white/10 hover:bg-black/5 dark:hover:bg-white/5 gap-2 h-10 px-5">
              <Download className="w-4 h-4" />
              Export CSV
            </Button>
            <Button onClick={() => setIsImporting(true)} className="rounded-full shadow-sm bg-black text-white hover:bg-zinc-800 dark:bg-white dark:text-black dark:hover:bg-zinc-200 gap-2 h-10 px-5">
              <Upload className="w-4 h-4" />
              Import CSV
            </Button>
          </div>
        </div>

        {/* Content Area */}
        <div className="flex-1 overflow-y-auto px-8 pb-8 custom-scrollbar">
          <div className="bg-white dark:bg-[#121212] rounded-3xl border border-black/5 dark:border-white/5 shadow-sm overflow-hidden min-h-[400px]">
            {(!jobs || (jobs as any[]).length === 0) ? (
              <div className="h-[400px] flex flex-col items-center justify-center text-center">
                <div className="w-16 h-16 rounded-full bg-black/5 dark:bg-white/5 flex items-center justify-center mb-4">
                  <FileSpreadsheet className="w-8 h-8 text-muted-foreground" />
                </div>
                <h3 className="text-lg font-semibold text-foreground">No import history</h3>
                <p className="text-sm text-muted-foreground max-w-sm mt-2 mb-6">Upload a CSV or Excel file to instantly map and import thousands of products into your catalog.</p>
                <div className="flex gap-3">
                  <Button onClick={() => setIsImporting(true)} className="rounded-full bg-black text-white hover:bg-zinc-800 dark:bg-white dark:text-black dark:hover:bg-zinc-200">
                    <Upload className="w-4 h-4 mr-2" />
                    Upload File
                  </Button>
                </div>
              </div>
            ) : (
              <table className="w-full text-left border-collapse">
                <thead>
                  <tr className="border-b border-black/5 dark:border-white/5">
                    <th className="py-4 px-6 text-xs font-semibold text-muted-foreground uppercase tracking-wider">Type / Status</th>
                    <th className="py-4 px-6 text-xs font-semibold text-muted-foreground uppercase tracking-wider">File Name</th>
                    <th className="py-4 px-6 text-xs font-semibold text-muted-foreground uppercase tracking-wider">Progress</th>
                    <th className="py-4 px-6 text-xs font-semibold text-muted-foreground uppercase tracking-wider">Created</th>
                    <th className="py-4 px-6 text-xs font-semibold text-muted-foreground uppercase tracking-wider text-right">Actions</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-black/5 dark:divide-white/5">
                  {(jobs as any[]).map((job: any) => {
                    let payloadStr = "{}";
                    let fileName = "unknown.csv";
                    try {
                      if (job.payload_json) {
                         const decoded = Buffer.from(job.payload_json, 'base64').toString('utf8');
                         const parsed = JSON.parse(decoded);
                         payloadStr = JSON.stringify(parsed);
                         if (parsed.file_name) fileName = parsed.file_name;
                      }
                    } catch(e) {}

                    return (
                      <tr key={job.id} className="hover:bg-black/5 dark:hover:bg-white/5 transition-colors group">
                        <td className="py-4 px-6">
                          <div className="flex items-center gap-2">
                            {getStatusIcon(job.status)}
                            <span className="font-medium text-xs tracking-wider uppercase text-foreground">{job.status}</span>
                          </div>
                        </td>
                        <td className="py-4 px-6">
                          <span className="text-sm font-medium text-foreground">
                            {fileName}
                          </span>
                        </td>
                        <td className="py-4 px-6">
                          <div className="flex items-center gap-2">
                            <div className="w-24 h-1.5 bg-black/10 dark:bg-white/10 rounded-full overflow-hidden">
                              <div 
                                className="h-full bg-primary rounded-full transition-all duration-500" 
                                style={{ width: job.total_items > 0 ? `${(job.processed_items / job.total_items) * 100}%` : '0%' }}
                              />
                            </div>
                            <span className="text-xs text-muted-foreground">{job.processed_items} / {job.total_items}</span>
                          </div>
                        </td>
                        <td className="py-4 px-6">
                          <span className="text-sm text-muted-foreground">{new Date(job.created_at).toLocaleString()}</span>
                        </td>
                        <td className="py-4 px-6 text-right">
                          <Button size="sm" variant="ghost" className="h-8 text-xs rounded-lg opacity-0 group-hover:opacity-100 transition-opacity">Details</Button>
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

      <aside className="w-[360px] flex-shrink-0 bg-[#121212] text-white p-6 rounded-3xl relative z-20 flex flex-col h-full shadow-[0_8px_30px_rgb(0,0,0,0.4)] border border-white/5">
        <PIMInsightsSidebar stats={stats} />
      </aside>

      {isImporting && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-md transition-all">
          <div className="w-[440px] bg-white/90 dark:bg-[#121212]/95 backdrop-blur-xl rounded-3xl p-8 shadow-[0_24px_48px_rgb(0,0,0,0.3)] border border-white/20 dark:border-white/10 flex flex-col gap-8 animate-in fade-in slide-in-from-bottom-4 zoom-in-95 duration-300">
            <div>
              <h2 className="text-2xl font-bold tracking-tight bg-gradient-to-br from-black to-black/70 dark:from-white dark:to-white/70 bg-clip-text text-transparent">Import CSV</h2>
              <p className="text-sm text-muted-foreground mt-1.5">Upload a product catalog spreadsheet.</p>
            </div>

            <div className="flex flex-col gap-5">
              <div className="flex flex-col gap-2">
                <label className="text-xs font-semibold text-foreground/80 tracking-wide">File Name (Mock)</label>
                <input 
                  type="text"
                  value={mockFileName}
                  onChange={(e) => setMockFileName(e.target.value)}
                  placeholder="e.g. q2_catalog_update.csv"
                  className="w-full h-11 px-4 rounded-xl bg-black/5 dark:bg-white/5 border border-transparent dark:border-white/5 text-sm placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-black/10 dark:focus:ring-white/20 focus:bg-white dark:focus:bg-black transition-all" 
                />
                <p className="text-xs text-muted-foreground mt-2">Note: This is a mockup modal. In production, this would be a dropzone to securely upload to S3.</p>
              </div>
            </div>

            <div className="flex items-center justify-end gap-3 mt-2">
              <Button onClick={() => setIsImporting(false)} variant="ghost" className="h-10 px-5 text-sm font-medium hover:bg-black/5 dark:hover:bg-white/5 rounded-xl transition-colors">Cancel</Button>
              <Button onClick={handleImport} disabled={!mockFileName} className="h-10 px-6 rounded-xl font-medium bg-gradient-to-b from-black to-zinc-800 text-white shadow-lg hover:shadow-xl hover:scale-[1.02] dark:from-white dark:to-zinc-200 dark:text-black transition-all active:scale-95 disabled:opacity-50 disabled:pointer-events-none">
                Start Import
              </Button>
            </div>
          </div>
        </div>
      )}
    </>
  );
}
