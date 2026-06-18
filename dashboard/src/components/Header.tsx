"use client";

import { Search, Bell } from "lucide-react";
import { Avatar, AvatarFallback, AvatarImage } from "./ui/avatar";
import { Popover, PopoverContent, PopoverTrigger } from "./ui/popover";
import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import { useAuth } from "@/providers/AuthProvider";
import { auth } from "@/lib/firebase";
import { signOut } from "firebase/auth";
import { LogOut, User } from "lucide-react";

export function Header() {
  const pathname = usePathname();
  const router = useRouter();
  const { user } = useAuth();

  if (pathname === "/login" || pathname === "/login/finish" || pathname === "/onboard") {
    return null;
  }

  const navItems = [
    { name: "Overview", href: "/" },
    { name: "PIM & Inventory", href: "/pim" },
    { name: "Orders", href: "/orders" },
    { name: "Channels", href: "/channels" },
    { name: "Analytics", href: "/analytics" },
    { name: "Teams", href: "/teams" },
    { name: "Settings", href: "/settings" },
  ];

  return (
    <header className="flex items-center justify-between px-10 py-8 w-full relative z-20">
      {/* Logo */}
      <Link href="/" className="flex items-center gap-3">
        <div className="w-10 h-10 bg-gradient-to-br from-amber-400 to-orange-600 rounded-full flex items-center justify-center shadow-lg">
          <div className="w-3.5 h-3.5 bg-white rounded-full"></div>
        </div>
        <div>
          <h1 className="text-xl font-bold tracking-tight leading-none text-foreground">Aurea</h1>
          <p className="text-[11px] font-medium text-muted-foreground mt-1 tracking-wider uppercase">AI Commerce Ops</p>
        </div>
      </Link>

      {/* Center Nav Pill */}
      <nav className="hidden lg:flex items-center bg-white/40 dark:bg-black/40 backdrop-blur-xl rounded-full p-1.5 border border-white/40 dark:border-white/10 shadow-sm ml-12">
        {navItems.map((item) => {
          // Highlight active if pathname exact matches '/' or starts with the specific path
          const isActive = item.href === "/" ? pathname === "/" : pathname.startsWith(item.href);
          
          return (
            <Link
              key={item.name}
              href={item.href}
              className={`px-6 py-2.5 text-sm font-semibold rounded-full transition-all ${
                isActive 
                  ? "bg-zinc-900 text-white dark:bg-white dark:text-zinc-900 shadow-md" 
                  : "text-foreground/70 hover:text-foreground hover:bg-black/5 dark:hover:bg-white/5"
              }`}
            >
              {item.name}
            </Link>
          );
        })}
      </nav>

      {/* Right Actions Pill */}
      <div className="flex items-center gap-2 bg-white/40 dark:bg-black/40 backdrop-blur-xl rounded-full p-2 px-4 border border-white/40 dark:border-white/10 shadow-sm">
        <button className="p-2 text-foreground/70 hover:text-foreground transition-colors rounded-full hover:bg-black/5 dark:hover:bg-white/5">
          <Search className="w-5 h-5" />
        </button>
        <button className="p-2 text-foreground/70 hover:text-foreground transition-colors rounded-full hover:bg-black/5 dark:hover:bg-white/5 relative">
          <Bell className="w-5 h-5" />
          <span className="absolute top-2.5 right-2.5 w-2 h-2 bg-amber-500 rounded-full border-2 border-background"></span>
        </button>
        <div className="w-[1px] h-6 bg-foreground/10 mx-2"></div>
        <Popover>
          <PopoverTrigger className="rounded-full cursor-pointer hover:opacity-80 transition-opacity focus:outline-none">
            <Avatar className="w-9 h-9 border border-white/20 shadow-sm pointer-events-none">
              <AvatarImage src="" />
              <AvatarFallback className="bg-purple-100 text-purple-700 dark:bg-purple-900 dark:text-purple-300 font-semibold uppercase">
                {user?.email?.charAt(0) || "U"}
              </AvatarFallback>
            </Avatar>
          </PopoverTrigger>
          <PopoverContent align="end" className="w-56 p-2 rounded-2xl bg-white/80 dark:bg-black/60 backdrop-blur-xl border border-black/5 dark:border-white/10 shadow-xl">
            <div className="px-2 py-2 mb-2 border-b border-black/5 dark:border-white/5">
              <p className="text-sm font-medium leading-none">{user?.email?.split('@')[0] || "User"}</p>
              <p className="text-xs text-muted-foreground mt-1 truncate">{user?.email || "No email"}</p>
            </div>
            <button 
              className="w-full flex items-center gap-2 px-2 py-2 text-sm text-foreground hover:bg-black/5 dark:hover:bg-white/5 rounded-xl transition-colors"
            >
              <User className="w-4 h-4" />
              Profile Settings
            </button>
            <button 
              onClick={() => {
                signOut(auth);
                router.push("/login");
              }}
              className="w-full flex items-center gap-2 px-2 py-2 text-sm text-red-600 dark:text-red-400 hover:bg-red-50 dark:hover:bg-red-950/30 rounded-xl transition-colors mt-1"
            >
              <LogOut className="w-4 h-4" />
              Sign out
            </button>
          </PopoverContent>
        </Popover>
      </div>
    </header>
  );
}
