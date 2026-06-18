import { Layers, Clock, Loader2, Package, CheckCircle2, XCircle, Undo2, Settings, Truck, RotateCcw, Zap, ShieldQuestion } from "lucide-react";
import Link from "next/link";
import { usePathname, useSearchParams } from "next/navigation";

const orderNav = [
  { name: "All Orders", icon: Layers, count: "1,248", href: "/orders", status: null },
  { name: "Shipping & Pickup", icon: Truck, count: "128", href: "/orders/shipping", status: "shipping" },
  { name: "Pending", icon: Clock, count: "64", href: "/orders?status=pending_payment", status: "pending_payment" },
  { name: "Processing", icon: Loader2, count: "312", href: "/orders?status=processing", status: "processing" },
  { name: "Shipped", icon: Package, count: "682", href: "/orders?status=fulfilled", status: "fulfilled" },
  { name: "Delivered", icon: CheckCircle2, count: "158", href: "/orders?status=completed", status: "completed" },
  { name: "Cancelled", icon: XCircle, count: "32", href: "/orders?status=cancelled", status: "cancelled" },
  { name: "Returns", icon: Undo2, count: "45", href: "/orders?status=returned", status: "returned" },
];

const configNav = [
  { name: "Order Settings", icon: Settings, href: "/orders?tab=settings", tab: "settings" },
  { name: "Shipping Rules", icon: Truck, href: "/orders?tab=shipping-rules", tab: "shipping-rules" },
  { name: "Return Rules", icon: RotateCcw, href: "/orders?tab=return-rules", tab: "return-rules" },
  { name: "Automation Rules", icon: Zap, href: "/orders?tab=automation", tab: "automation" },
];

export function LeftSidebar() {
  const pathname = usePathname();
  const searchParams = useSearchParams();
  const activeStatus = searchParams.get("status");
  const activeTab = searchParams.get("tab");

  return (
    <aside className="w-[240px] h-full flex-shrink-0 flex flex-col pr-4 overflow-y-auto custom-scrollbar justify-center">
      
      {/* Order Management Section */}
      <div className="mb-8 font-sans">
        <h3 className="text-xs font-semibold text-muted-foreground uppercase tracking-wider mb-4 px-2">Order Management</h3>
        <nav className="flex flex-col gap-1">
          {orderNav.map((item) => {
            let isActive = false;
            if (item.status === "shipping") {
              isActive = pathname === "/orders/shipping";
            } else if (item.status === null) {
              isActive = pathname === "/orders" && !activeStatus && !activeTab;
            } else {
              isActive = pathname === "/orders" && activeStatus === item.status;
            }

            return (
              <Link 
                key={item.name} 
                href={item.href} 
                className={`flex items-center justify-between px-3 py-2 rounded-xl text-sm transition-colors ${
                  isActive 
                    ? "bg-primary/10 text-primary font-medium" 
                    : "text-muted-foreground hover:bg-black/5 dark:hover:bg-white/5 hover:text-foreground"
                }`}
              >
                <div className="flex items-center gap-3">
                  <item.icon className={`w-4 h-4 ${isActive ? "text-primary" : "text-muted-foreground"}`} />
                  {item.name}
                </div>
                {item.count && (
                  <span className={`text-xs ${isActive ? "text-primary font-semibold" : "text-muted-foreground"}`}>
                    {item.count}
                  </span>
                )}
              </Link>
            );
          })}
        </nav>
      </div>

      {/* Configuration Section */}
      <div className="mb-8">
        <h3 className="text-xs font-semibold text-muted-foreground uppercase tracking-wider mb-4 px-2">Configuration</h3>
        <nav className="flex flex-col gap-1">
          {configNav.map((item) => {
            const isActive = pathname === "/orders" && activeTab === item.tab;
            return (
              <Link 
                key={item.name} 
                href={item.href} 
                className={`flex items-center gap-3 px-3 py-2 rounded-xl text-sm transition-colors ${
                  isActive 
                    ? "bg-primary/10 text-primary font-medium" 
                    : "text-muted-foreground hover:bg-black/5 dark:hover:bg-white/5 hover:text-foreground"
                }`}
              >
                <item.icon className={`w-4 h-4 ${isActive ? "text-primary" : "text-muted-foreground"}`} />
                {item.name}
              </Link>
            );
          })}
        </nav>
      </div>

      {/* Spacer removed for vertical centering */}
      <div className="mt-8" />

      {/* Need Help Card */}
      <div className="bg-white/40 dark:bg-black/20 backdrop-blur-md rounded-2xl p-4 border border-white/40 dark:border-white/10 mt-6">
        <div className="flex items-start gap-3">
          <div className="w-8 h-8 rounded-full bg-purple-100 dark:bg-purple-900/30 flex items-center justify-center flex-shrink-0">
            <ShieldQuestion className="w-4 h-4 text-purple-600 dark:text-purple-400" />
          </div>
          <div>
            <h4 className="text-sm font-semibold text-foreground">Need help?</h4>
            <p className="text-xs text-muted-foreground mt-1">Visit our help center or contact support</p>
          </div>
        </div>
      </div>
      
    </aside>
  );
}
