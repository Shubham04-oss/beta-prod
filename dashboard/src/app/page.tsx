"use client";

import { CalendarDays, Wallet, ShoppingCart, TrendingUp, Tag, ArrowUpRight, ArrowDownRight, PackageX, TrendingUp as TrendingUpIcon, Megaphone, CheckCircle2 } from "lucide-react";
import { Button, buttonVariants } from "@/components/ui/button";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import { Calendar } from "@/components/ui/calendar";
import { Area, AreaChart, CartesianGrid, XAxis, YAxis, ResponsiveContainer, Tooltip } from "recharts";
import { useState } from "react";
import { RightSidebar } from "@/components/RightSidebar";

const chartData = [
  { name: "May 1", revenue: 120000 },
  { name: "May 2", revenue: 180000 },
  { name: "May 3", revenue: 140000 },
  { name: "May 4", revenue: 220000 },
  { name: "May 5", revenue: 190000 },
  { name: "May 6", revenue: 280000 },
  { name: "May 7", revenue: 260000 },
];

export default function Home() {
  const [date, setDate] = useState<Date | undefined>(new Date());

  return (
    <>
      <main className="flex-1 overflow-y-auto custom-scrollbar pr-4">
        <div className="flex flex-col gap-6 pb-4">
      {/* Top Section Container */}
      <div className="bg-white/20 dark:bg-white/5 rounded-[2rem] p-6 lg:p-8 border border-white/30 dark:border-white/10 flex flex-col gap-8">
        {/* Header Row */}
        <div className="flex items-center justify-between">
          <div>
            <h2 className="text-3xl font-bold tracking-tight flex items-center gap-2">
              Good morning, Alex <span className="text-amber-500">☀️</span>
            </h2>
            <p className="text-muted-foreground mt-1">Here's what's happening with your business today.</p>
          </div>
          <div className="flex items-center gap-3">
            <Popover>
              <PopoverTrigger className={buttonVariants({ variant: "outline", className: "bg-white/40 dark:bg-black/20 backdrop-blur-md border-white/40 dark:border-white/10 rounded-full shadow-sm hover:bg-white/60 dark:hover:bg-white/10" })}>
                <CalendarDays className="mr-2 h-4 w-4" />
                May 1 – May 7, 2024
              </PopoverTrigger>
              <PopoverContent className="w-auto p-0 rounded-2xl" align="end">
                <Calendar mode="single" selected={date} onSelect={setDate} />
              </PopoverContent>
            </Popover>
          </div>
        </div>

        {/* Main KPI Cards (Naked) */}
        <div className="grid grid-cols-1 md:grid-cols-4 gap-6 relative">
          <KPICard title="Total Revenue" value="$482,430" trend="+12.6%" trendUp={true} icon={<Wallet className="text-amber-600" />} iconBg="bg-amber-100 dark:bg-amber-900/30" />
          <div className="hidden md:block absolute w-px h-16 bg-black/5 dark:bg-white/5 left-[25%] top-1/2 -translate-y-1/2"></div>
          
          <KPICard title="Orders" value="1,842" trend="+8.7%" trendUp={true} icon={<ShoppingCart className="text-blue-600" />} iconBg="bg-blue-100 dark:bg-blue-900/30" />
          <div className="hidden md:block absolute w-px h-16 bg-black/5 dark:bg-white/5 left-[50%] top-1/2 -translate-y-1/2"></div>
          
          <KPICard title="Conversion Rate" value="3.6%" trend="+4.3%" trendUp={true} icon={<TrendingUp className="text-green-600" />} iconBg="bg-green-100 dark:bg-green-900/30" />
          <div className="hidden md:block absolute w-px h-16 bg-black/5 dark:bg-white/5 left-[75%] top-1/2 -translate-y-1/2"></div>
          
          <KPICard title="Average Order Value" value="$78.6" trend="+6.1%" trendUp={true} icon={<Tag className="text-purple-600" />} iconBg="bg-purple-100 dark:bg-purple-900/30" />
        </div>
      </div>

      {/* Revenue Trend Chart Container */}
      <div className="bg-white/20 dark:bg-white/5 rounded-[2rem] p-6 lg:p-8 border border-white/30 dark:border-white/10">
        <div className="flex items-center justify-between mb-6">
          <h3 className="font-semibold text-lg">Revenue Trend</h3>
          <Button variant="outline" size="sm" className="rounded-full bg-transparent border-white/40 dark:border-white/10">
            This Week <ChevronDown className="ml-2 w-4 h-4" />
          </Button>
        </div>
        <div className="h-[250px] w-full">
          <ResponsiveContainer width="100%" height="100%">
            <AreaChart data={chartData} margin={{ top: 10, right: 0, left: -20, bottom: 0 }}>
              <defs>
                <linearGradient id="colorRev" x1="0" y1="0" x2="0" y2="1">
                  <stop offset="5%" stopColor="#f59e0b" stopOpacity={0.3} />
                  <stop offset="95%" stopColor="#f59e0b" stopOpacity={0} />
                </linearGradient>
              </defs>
              <CartesianGrid strokeDasharray="3 3" vertical={false} stroke="currentColor" className="opacity-10" />
              <XAxis dataKey="name" axisLine={false} tickLine={false} tick={{ fontSize: 12 }} className="text-muted-foreground opacity-70" />
              <YAxis axisLine={false} tickLine={false} tick={{ fontSize: 12 }} tickFormatter={(val) => `${val / 1000}K`} className="text-muted-foreground opacity-70" />
              <Tooltip 
                contentStyle={{ borderRadius: '12px', border: 'none', boxShadow: '0 4px 6px -1px rgb(0 0 0 / 0.1)' }}
                itemStyle={{ color: '#f59e0b', fontWeight: 600 }}
              />
              <Area type="monotone" dataKey="revenue" stroke="#f59e0b" strokeWidth={3} fillOpacity={1} fill="url(#colorRev)" />
            </AreaChart>
          </ResponsiveContainer>
        </div>
      </div>

      {/* Sales by Channel Container */}
      <div className="bg-white/20 dark:bg-white/5 rounded-[2rem] p-6 lg:p-8 border border-white/30 dark:border-white/10">
        <div className="flex items-center justify-between mb-6">
          <h3 className="font-semibold text-lg">Sales by Channel</h3>
          <button className="text-sm text-muted-foreground hover:text-foreground flex items-center transition-colors">
            View all channels <ChevronRight className="w-4 h-4 ml-1" />
          </button>
        </div>
        <div className="grid grid-cols-2 lg:grid-cols-5 gap-4 relative">
          <ChannelCard name="Amazon" rev="$156,230" trend="+12%" logo="https://cdn.jsdelivr.net/gh/devicons/devicon@latest/icons/amazonwebservices/amazonwebservices-original-wordmark.svg" />
          <div className="hidden lg:block absolute w-px h-12 bg-black/5 dark:bg-white/5 left-[20%] top-1/2 -translate-y-1/2"></div>
          
          <ChannelCard name="Shopify" rev="$124,760" trend="+8%" logo="https://cdn.simpleicons.org/shopify/95BF47" />
          <div className="hidden lg:block absolute w-px h-12 bg-black/5 dark:bg-white/5 left-[40%] top-1/2 -translate-y-1/2"></div>
          
          <ChannelCard name="Flipkart" rev="$98,540" trend="+17%" logo="https://cdn.simpleicons.org/flipkart/2874F0" />
          <div className="hidden lg:block absolute w-px h-12 bg-black/5 dark:bg-white/5 left-[60%] top-1/2 -translate-y-1/2"></div>
          
          <ChannelCard name="TikTok Shop" rev="$28,050" trend="+5%" logo="https://cdn.simpleicons.org/tiktok/black" />
          <div className="hidden lg:block absolute w-px h-12 bg-black/5 dark:bg-white/5 left-[80%] top-1/2 -translate-y-1/2"></div>
          
          <ChannelCard name="Etsy" rev="$32,850" trend="-2%" trendUp={false} logo="https://cdn.simpleicons.org/etsy/F16126" />
        </div>
      </div>

      {/* Bottom Small Cards Container */}
      <div className="bg-white/20 dark:bg-white/5 rounded-[2rem] p-6 lg:p-8 border border-white/30 dark:border-white/10">
        <div className="grid grid-cols-1 md:grid-cols-4 gap-4 relative">
          <BottomCard icon={<PackageX className="w-5 h-5 text-amber-600" />} iconBg="bg-amber-100" title="Low stock" subtitle="Wireless Headphones" alert="23 left in Mumbai" />
          <div className="hidden md:block absolute w-px h-12 bg-black/5 dark:bg-white/5 left-[25%] top-1/2 -translate-y-1/2"></div>
          
          <BottomCard icon={<TrendingUpIcon className="w-5 h-5 text-green-600" />} iconBg="bg-green-100" title="Top channel" subtitle="Shopify (+23%)" subtext="vs last week" />
          <div className="hidden md:block absolute w-px h-12 bg-black/5 dark:bg-white/5 left-[50%] top-1/2 -translate-y-1/2"></div>
          
          <BottomCard icon={<Megaphone className="w-5 h-5 text-blue-600" />} iconBg="bg-blue-100" title="2 campaigns ready" subtitle="Weekend Sale" action="Review now →" />
          <div className="hidden md:block absolute w-px h-12 bg-black/5 dark:bg-white/5 left-[75%] top-1/2 -translate-y-1/2"></div>
          
          <BottomCard icon={<Wallet className="w-5 h-5 text-emerald-600" />} iconBg="bg-emerald-100" title="Estimated profit today" subtitle="$18,240" subtext="↑ 14% vs yesterday" />
        </div>
      </div>
    </div>
      </main>

      {/* Right Matte Black Area */}
      <aside className="w-[360px] flex-shrink-0 bg-[#121212] text-white p-6 rounded-3xl relative z-20 flex flex-col h-full shadow-[0_8px_30px_rgb(0,0,0,0.4)] border border-white/5">
        <RightSidebar />
      </aside>
    </>
  );
}

// Subcomponents

function KPICard({ title, value, trend, trendUp, icon }: any) {
  return (
    <div className="p-2 flex flex-col justify-between">
      <div className="flex items-start gap-4">
        <div className="w-12 h-12 flex items-center justify-center">
          {icon}
        </div>
        <div>
          <p className="text-sm font-medium text-muted-foreground">{title}</p>
          <h3 className="text-2xl font-bold mt-1 text-foreground">{value}</h3>
          <p className={`text-xs mt-2 font-medium flex items-center gap-1 ${trendUp ? 'text-green-600 dark:text-green-500' : 'text-red-600 dark:text-red-500'}`}>
            {trendUp ? <ArrowUpRight className="w-3 h-3" /> : <ArrowDownRight className="w-3 h-3" />}
            {trend} <span className="text-muted-foreground font-normal ml-1">vs last week</span>
          </p>
        </div>
      </div>
    </div>
  );
}

function ChannelCard({ name, rev, trend, trendUp = true, logo }: any) {
  return (
    <div className="p-2 flex items-center gap-4">
      <div className="w-10 h-10 flex items-center justify-center p-1 flex-shrink-0">
        <img src={logo} alt={name} className="w-full h-full object-contain" />
      </div>
      <div>
        <p className="text-xs font-semibold text-muted-foreground">{name}</p>
        <p className="text-base font-bold text-foreground mt-0.5">{rev}</p>
        <p className={`text-[10px] mt-0.5 font-bold ${trendUp ? 'text-green-600 dark:text-green-500' : 'text-red-600 dark:text-red-500'}`}>
          {trendUp ? '↑' : '↓'} {trend}
        </p>
      </div>
    </div>
  );
}

function BottomCard({ icon, title, subtitle, alert, subtext, action }: any) {
  return (
    <div className="p-2 flex items-center gap-4">
      <div className="w-10 h-10 flex items-center justify-center flex-shrink-0">
        {icon}
      </div>
      <div className="min-w-0">
        <p className="text-[11px] font-semibold text-muted-foreground truncate uppercase tracking-wider">{title}</p>
        <p className="text-sm font-bold text-foreground mt-0.5 truncate">{subtitle}</p>
        {alert && <p className="text-[11px] font-semibold text-amber-600 mt-1">{alert}</p>}
        {subtext && <p className="text-[11px] text-muted-foreground mt-1">{subtext}</p>}
        {action && <p className="text-[11px] font-semibold text-blue-600 mt-1 cursor-pointer hover:underline">{action}</p>}
      </div>
    </div>
  );
}

function ChevronDown(props: any) {
  return <svg {...props} xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><path d="m6 9 6 6 6-6"/></svg>
}

function ChevronRight(props: any) {
  return <svg {...props} xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><path d="m9 18 6-6-6-6"/></svg>
}
