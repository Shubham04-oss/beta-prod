"use client";

import { Sparkles, MoreVertical, ChevronRight, Tag, Package, Megaphone, Headphones } from "lucide-react";

const agents = [
  { name: "Pricing Team", role: "Pricing Optimization", task: "Reviewing competitor prices", eta: "Next update in 10 mins", icon: Tag, color: "text-amber-500", bg: "bg-amber-500/10", border: "border-amber-500/20" },
  { name: "Inventory Team", role: "Stock Management", task: "Checking warehouse stock", eta: "2 alerts found", icon: Package, color: "text-blue-500", bg: "bg-blue-500/10", border: "border-blue-500/20", alert: true },
  { name: "Ads Team", role: "Marketing & Campaigns", task: "Preparing weekend campaign", eta: "Scheduled at 2:00 PM", icon: Megaphone, color: "text-rose-500", bg: "bg-rose-500/10", border: "border-rose-500/20" },
  { name: "Support Team", role: "Customer Experience", task: "Responding to customer queries", eta: "12 conversations today", icon: Headphones, color: "text-emerald-500", bg: "bg-emerald-500/10", border: "border-emerald-500/20" },
];

export function RightSidebar() {
  return (
    <div className="flex flex-col h-full w-full">
      {/* Sidebar Header */}
      <div className="flex items-center gap-3 mb-10 mt-2">
        <Sparkles className="w-5 h-5 text-amber-500" />
        <div>
          <h2 className="font-semibold text-lg leading-tight tracking-tight">Your Digital Team</h2>
          <p className="text-xs text-zinc-400 mt-1">AI teammates working behind the scenes</p>
        </div>
      </div>

      {/* Agents List */}
      <div className="flex-1 flex flex-col gap-6 overflow-y-auto pr-2 custom-scrollbar">
        {agents.map((agent) => (
          <div key={agent.name} className="flex gap-4 group">
            {/* Team Icon Badge */}
            <div className="relative flex-shrink-0">
              <div className={`w-12 h-12 rounded-full flex items-center justify-center border ${agent.bg} ${agent.border}`}>
                <agent.icon className={`w-5 h-5 ${agent.color}`} />
              </div>
              <div className={`absolute -bottom-0.5 -right-0.5 w-3.5 h-3.5 rounded-full border-[2.5px] border-[#121212] ${agent.alert ? 'bg-amber-500' : 'bg-green-500'}`}></div>
            </div>
            
            {/* Details */}
            <div className="flex-1 min-w-0">
              <div className="flex justify-between items-start">
                <div className="flex items-center gap-1.5">
                  <h3 className="font-medium text-sm text-zinc-100">{agent.name}</h3>
                  <Sparkles className="w-3 h-3 text-amber-500/80" />
                </div>
                <button className="text-zinc-600 hover:text-zinc-300 opacity-0 group-hover:opacity-100 transition-opacity p-1 -mr-1">
                  <MoreVertical className="w-4 h-4" />
                </button>
              </div>
              <p className="text-[11px] text-zinc-400 mt-0.5">{agent.role}</p>
              
              {/* Task Bubble */}
              <div className="mt-3 bg-zinc-900/60 rounded-xl p-3 border border-zinc-800/60 shadow-inner">
                <div className="flex items-center gap-2">
                  <div className={`w-1.5 h-1.5 rounded-full flex-shrink-0 ${agent.alert ? 'bg-amber-500' : 'bg-green-500'}`}></div>
                  <p className="text-xs text-zinc-300 truncate font-medium">{agent.task}</p>
                </div>
                <p className="text-[10px] text-amber-500/80 mt-1.5 font-semibold ml-3.5 flex items-center gap-1">
                  <Sparkles className="w-2.5 h-2.5" />
                  {agent.eta}
                </p>
              </div>
            </div>
          </div>
        ))}
      </div>

      {/* Footer Action */}
      <button className="mt-6 flex items-center justify-between w-full p-4 rounded-2xl bg-zinc-900 border border-zinc-800 text-sm font-medium text-zinc-300 hover:text-white hover:bg-zinc-800 transition-all hover:scale-[1.02] shadow-sm">
        <div className="flex items-center gap-2.5">
          <svg className="w-4 h-4 text-zinc-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z" />
          </svg>
          View all team activity
        </div>
        <ChevronRight className="w-4 h-4 text-zinc-500" />
      </button>
    </div>
  );
}
