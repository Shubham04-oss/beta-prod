"use client";

import { 
  ArrowLeft,
  Layers, 
  Plus,
  Settings2,
  FolderTree,
  Box,
  Tags,
  RefreshCw,
  Upload,
  FileText,
  FileBadge,
  Image as ImageIcon,
  ShieldQuestion
} from "lucide-react";
import Link from "next/link";
import { usePathname } from "next/navigation";

const productNav = [
  { name: "All Products", href: "/pim", icon: Layers },
  { name: "Inventory", href: "/pim/inventory", icon: Box },
  { name: "Product Configurator", href: "/pim/product", icon: Settings2 },
  { name: "Categories", href: "/pim/categories", icon: FolderTree },
  { name: "Brands", href: "/pim/brands", icon: Box },
  { name: "Attributes", href: "/pim/attributes", icon: Tags },
  { name: "Bulk Update", href: "/pim/bulk", icon: RefreshCw },
  { name: "Import / Export", href: "/pim/import-export", icon: Upload },
];

const toolsNav = [
  { name: "Product Templates", href: "/pim/templates", icon: FileText },
  { name: "Attribute Groups", href: "/pim/attribute-groups", icon: FileBadge },
  { name: "Media Library", href: "/pim/media", icon: ImageIcon },
];

export function PIMLeftSidebar() {
  const pathname = usePathname();

  return (
    <aside className="w-[240px] h-full flex-shrink-0 flex flex-col pr-4 overflow-y-auto custom-scrollbar justify-center">
      
      {pathname === "/pim/product" && (
        <div className="mb-6 px-2">
          <Link href="/pim" className="flex items-center gap-2 text-xs font-medium text-muted-foreground hover:text-foreground transition-colors">
            <ArrowLeft className="w-3.5 h-3.5" />
            Back to PIM & Inventory
          </Link>
        </div>
      )}

      {/* Product Management Section */}
      <div className="mb-8 mt-2">
        <h3 className="text-[10px] font-bold text-muted-foreground uppercase tracking-wider mb-4 px-2">Product Management</h3>
        <nav className="flex flex-col gap-1">
          {productNav.map((item) => {
            const isActive = pathname === item.href;
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
              </Link>
            );
          })}
        </nav>
      </div>

      {/* Data Quality Section */}
      <div className="mb-8">
        <h3 className="text-[10px] font-bold text-muted-foreground uppercase tracking-wider mb-4 px-2">Data Quality</h3>
        <div className="flex flex-col gap-4 px-3">
          
          <div className="flex flex-col gap-2 cursor-pointer group">
            <div className="flex items-center justify-between">
              <span className="text-sm text-muted-foreground group-hover:text-foreground transition-colors">Completeness Score</span>
              <div className="flex items-center gap-2">
                <div className="w-1.5 h-1.5 rounded-full bg-green-500"></div>
                <span className="text-sm font-bold text-foreground">92%</span>
              </div>
            </div>
            <div className="w-full h-1 bg-black/5 dark:bg-white/5 rounded-full overflow-hidden">
              <div className="h-full bg-green-500 rounded-full" style={{ width: '92%' }}></div>
            </div>
          </div>

          <Link href="/pim/validation" className="flex items-center justify-between cursor-pointer group">
            <span className="text-sm text-muted-foreground group-hover:text-foreground transition-colors">Validation Issues</span>
            <span className="text-sm font-bold text-red-500">23</span>
          </Link>

          <Link href="/pim/audit" className="flex items-center justify-between cursor-pointer group">
            <span className="text-sm text-muted-foreground group-hover:text-foreground transition-colors">Data Audit Log</span>
          </Link>

        </div>
      </div>

      {/* Tools Section */}
      <div className="mb-8">
        <h3 className="text-[10px] font-bold text-muted-foreground uppercase tracking-wider mb-4 px-2">Tools</h3>
        <nav className="flex flex-col gap-1">
          {toolsNav.map((item) => (
            <Link 
              key={item.name} 
              href={item.href} 
              className="flex items-center gap-3 px-3 py-2 rounded-xl text-sm text-muted-foreground hover:bg-black/5 dark:hover:bg-white/5 hover:text-foreground transition-colors"
            >
              <item.icon className="w-4 h-4" />
              {item.name}
            </Link>
          ))}
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
          <div className="flex flex-col">
            <h4 className="text-sm font-semibold text-foreground">Need help?</h4>
            <p className="text-[10px] text-muted-foreground mt-0.5">Visit our help center or contact support</p>
          </div>
        </div>
      </div>
      
    </aside>
  );
}
