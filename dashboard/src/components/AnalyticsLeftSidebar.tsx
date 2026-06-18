import { 
  LayoutDashboard, 
  TrendingUp, 
  ShoppingCart, 
  Truck, 
  Box, 
  Tag, 
  Users, 
  Share2, 
  Activity, 
  FileText, 
  Save, 
  Clock, 
  ShieldQuestion 
} from "lucide-react";
import Link from "next/link";

const reportsNav = [
  { name: "Overview", icon: LayoutDashboard, active: true },
  { name: "Sales Performance", icon: TrendingUp },
  { name: "Order Analytics", icon: ShoppingCart },
  { name: "Shipment Analytics", icon: Truck },
  { name: "Inventory Reports", icon: Box },
  { name: "Product Performance", icon: Tag },
  { name: "Customer Reports", icon: Users },
  { name: "Channel Performance", icon: Share2 },
  { name: "Operational Reports", icon: Activity },
  { name: "Custom Reports", icon: FileText },
];

const configNav = [
  { name: "Saved Reports", icon: Save },
  { name: "Schedules", icon: Clock },
];

export function AnalyticsLeftSidebar() {
  return (
    <aside className="w-[240px] h-full flex-shrink-0 flex flex-col pr-4 overflow-y-auto custom-scrollbar justify-center">
      
      {/* REPORTS Section */}
      <div className="mb-8 mt-2">
        <h3 className="text-[10px] font-bold text-muted-foreground uppercase tracking-wider mb-4 px-2">Reports</h3>
        <nav className="flex flex-col gap-1">
          {reportsNav.map((item) => (
            <Link 
              key={item.name} 
              href="#" 
              className={`flex items-center gap-3 px-3 py-2 rounded-xl text-sm transition-colors ${
                item.active 
                  ? "bg-primary/10 text-primary font-semibold" 
                  : "text-muted-foreground hover:bg-black/5 dark:hover:bg-white/5 hover:text-foreground font-medium"
              }`}
            >
              <item.icon className={`w-4 h-4 ${item.active ? "text-primary" : "text-muted-foreground"}`} />
              {item.name}
            </Link>
          ))}
        </nav>
      </div>

      {/* REPORT CONFIGURATION Section */}
      <div className="mb-8">
        <h3 className="text-[10px] font-bold text-muted-foreground uppercase tracking-wider mb-4 px-2">Report Configuration</h3>
        <nav className="flex flex-col gap-1">
          {configNav.map((item) => (
            <Link 
              key={item.name} 
              href="#" 
              className="flex items-center gap-3 px-3 py-2 rounded-xl text-sm font-medium text-muted-foreground hover:bg-black/5 dark:hover:bg-white/5 hover:text-foreground transition-colors"
            >
              <item.icon className="w-4 h-4" />
              {item.name}
            </Link>
          ))}
        </nav>
      </div>

      <div className="mt-8" />

      {/* Need Help Card */}
      <div className="bg-white/40 dark:bg-black/20 backdrop-blur-md rounded-2xl p-4 border border-white/40 dark:border-white/10 mt-6">
        <div className="flex items-start gap-3">
          <div className="w-8 h-8 rounded-full bg-blue-100 dark:bg-blue-900/30 flex items-center justify-center flex-shrink-0">
            <ShieldQuestion className="w-4 h-4 text-blue-600 dark:text-blue-400" />
          </div>
          <div className="flex flex-col">
            <h4 className="text-sm font-semibold text-foreground">Need help?</h4>
            <p className="text-[10px] text-muted-foreground mt-0.5">Visit our help center or contact support</p>
          </div>
        </div>
      </div>
      
    </aside>
  );
}
