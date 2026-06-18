import { 
  Settings as SettingsIcon,
  Building2,
  Users,
  Bell,
  Puzzle,
  CreditCard,
  ShieldCheck,
  Lock,
  Zap,
  FileClock,
  Terminal,
  ShieldQuestion
} from "lucide-react";
import Link from "next/link";

const navItems = [
  { name: "General", desc: "Workspace, profile & preferences", icon: SettingsIcon, href: "/settings" },
  { name: "Organization", desc: "Company, members & roles", icon: Building2, href: "/settings/organization" },
  { name: "AI Team Assignments", desc: "Assign teams to owners", icon: Users, href: "#" },
  { name: "Notifications", desc: "Alerts, email & in-app", icon: Bell, href: "#" },
  { name: "Integrations", desc: "Platforms & third-party apps", icon: Puzzle, href: "/settings/integrations" },
  { name: "Billing & Plans", desc: "Subscription & payment", icon: CreditCard, href: "#" },
  { name: "Security", desc: "Access, SSO & authentication", icon: ShieldCheck, href: "#" },
  { name: "Data & Privacy", desc: "Data policies & exports", icon: Lock, href: "#" },
  { name: "Automation", desc: "Schedules, rules & triggers", icon: Zap, href: "#" },
  { name: "Audit Logs", desc: "Activity & change history", icon: FileClock, href: "/settings/audit" },
  { name: "Developer", desc: "API keys & webhooks", icon: Terminal, href: "/settings/developer" },
];
import { usePathname } from "next/navigation";

export function SettingsLeftSidebar() {
  const pathname = usePathname();
  return (
    <aside className="w-[280px] h-full flex-shrink-0 flex flex-col pr-4 overflow-y-auto custom-scrollbar justify-center">
      
      <div className="mb-6 px-2 mt-2">
        <h2 className="text-xl font-bold tracking-tight">Settings</h2>
      </div>

      <nav className="flex flex-col gap-1 mb-8">
        {navItems.map((item) => {
          const isActive = pathname === item.href;
          return (
            <Link 
              key={item.name} 
              href={item.href} 
              className={`flex flex-col justify-center px-3 py-2.5 rounded-xl transition-colors ${
                isActive 
                  ? "bg-primary/10 text-primary" 
                  : "text-muted-foreground hover:bg-black/5 dark:hover:bg-white/5 hover:text-foreground"
              }`}
            >
              <div className="flex items-center gap-3">
                <item.icon className={`w-4 h-4 flex-shrink-0 ${isActive ? "text-primary" : "text-muted-foreground"}`} />
                <span className={`text-sm font-semibold ${isActive ? "text-primary" : "text-foreground"}`}>
                  {item.name}
                </span>
              </div>
              <span className={`text-[10px] mt-0.5 ml-7 ${isActive ? "text-primary/80" : "text-muted-foreground"}`}>
                {item.desc}
              </span>
            </Link>
          );
        })}
      </nav>

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
