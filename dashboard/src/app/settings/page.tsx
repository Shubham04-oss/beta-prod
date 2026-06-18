"use client";

import { useEffect, useState } from "react";
import { SettingsLeftSidebar } from "@/components/SettingsLeftSidebar";
import { SettingsInsightsSidebar } from "@/components/SettingsInsightsSidebar";
import { ChevronDown, Save, ShieldCheck, SlidersHorizontal, RotateCcw } from "lucide-react";
import { Button } from "@/components/ui/button";
import { fetchAPI } from "@/lib/api";

type TenantSettings = {
  id: string;
  org_id: string;
  tenant_id: string;
  inventory_allocation_model: "HARD" | "SOFT";
  auto_po_enabled: boolean;
  default_low_stock_threshold: number;
  costing_method: "WAC" | "FIFO";
  updated_by?: string;
  created_at: string;
  updated_at: string;
};

const defaultSettings: TenantSettings = {
  id: "",
  org_id: "",
  tenant_id: "",
  inventory_allocation_model: "HARD",
  auto_po_enabled: false,
  default_low_stock_threshold: 10,
  costing_method: "WAC",
  created_at: "",
  updated_at: "",
};

export default function SettingsPage() {
  const [settings, setSettings] = useState<TenantSettings>(defaultSettings);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [message, setMessage] = useState<{ type: "success" | "error"; text: string } | null>(null);

  useEffect(() => {
    let mounted = true;
    fetchAPI("/api/v1/settings/tenant")
      .then((data) => {
        if (!mounted) return;
        setSettings(data as TenantSettings);
        setMessage(null);
      })
      .catch(() => {
        if (!mounted) return;
        setSettings(defaultSettings);
        setMessage({ type: "error", text: "Tenant settings have not been initialized yet. Saving will create them." });
      })
      .finally(() => {
        if (mounted) setLoading(false);
      });
    return () => {
      mounted = false;
    };
  }, []);

  const saveSettings = async () => {
    setSaving(true);
    setMessage(null);
    try {
      const saved = await fetchAPI("/api/v1/settings/tenant", {
        method: "PUT",
        body: JSON.stringify({
          inventory_allocation_model: settings.inventory_allocation_model,
          auto_po_enabled: settings.auto_po_enabled,
          default_low_stock_threshold: settings.default_low_stock_threshold,
          costing_method: settings.costing_method,
        }),
      });
      setSettings(saved as TenantSettings);
      setMessage({ type: "success", text: "Tenant settings saved and audited." });
    } catch (err) {
      setMessage({ type: "error", text: err instanceof Error ? err.message : "Failed to save settings." });
    } finally {
      setSaving(false);
    }
  };

  return (
    <>
      <SettingsLeftSidebar />

      <main className="flex-1 overflow-y-auto custom-scrollbar pr-4 flex flex-col gap-8 pb-10">
        <div className="mt-2 flex items-end justify-between gap-4">
          <div>
            <h1 className="text-3xl font-bold tracking-tight">General Settings</h1>
            <p className="text-muted-foreground mt-1">Manage tenant-level operational configuration with audited changes.</p>
          </div>
          <Button onClick={saveSettings} disabled={saving || loading} className="rounded-full shadow-sm">
            <Save className="mr-2 h-4 w-4" />
            {saving ? "Saving..." : "Save Settings"}
          </Button>
        </div>

        {message && (
          <div className={`p-3 rounded-xl text-sm border ${message.type === "success" ? "bg-emerald-500/10 text-emerald-700 dark:text-emerald-400 border-emerald-500/20" : "bg-amber-500/10 text-amber-700 dark:text-amber-400 border-amber-500/20"}`}>
            {message.text}
          </div>
        )}

        <div className="grid grid-cols-2 gap-6">
          <section className="bg-white/20 dark:bg-white/5 rounded-[2rem] p-6 border border-white/30 dark:border-white/10 flex flex-col gap-6">
            <div className="flex items-center gap-2">
              <ShieldCheck className="w-5 h-5 text-foreground" />
              <div>
                <h3 className="font-semibold text-[15px]">Tenant Identity</h3>
                <p className="text-[11px] text-muted-foreground mt-0.5">Read-only identifiers from the authenticated tenant context.</p>
              </div>
            </div>

            <ReadOnlyField label="Tenant ID" value={settings.tenant_id || "Pending initialization"} />
            <ReadOnlyField label="Organization ID" value={settings.org_id || "Pending initialization"} />
            <ReadOnlyField label="Last Updated" value={settings.updated_at ? new Date(settings.updated_at).toLocaleString() : "Not saved yet"} />
          </section>

          <section className="bg-white/20 dark:bg-white/5 rounded-[2rem] p-6 border border-white/30 dark:border-white/10 flex flex-col gap-6">
            <div className="flex items-center gap-2">
              <SlidersHorizontal className="w-5 h-5 text-foreground" />
              <div>
                <h3 className="font-semibold text-[15px]">Inventory Allocation</h3>
                <p className="text-[11px] text-muted-foreground mt-0.5">Control reservation strictness and replenishment behavior.</p>
              </div>
            </div>

            <SelectField
              label="Allocation Model"
              value={settings.inventory_allocation_model}
              onChange={(value) => setSettings((current) => ({ ...current, inventory_allocation_model: value as "HARD" | "SOFT" }))}
              options={[
                { value: "HARD", label: "Hard reservation" },
                { value: "SOFT", label: "Soft reservation" },
              ]}
            />

            <div className="flex items-center justify-between rounded-xl bg-white/40 dark:bg-black/20 border border-black/5 dark:border-white/10 p-4">
              <div>
                <p className="text-[13px] font-semibold">Auto purchase orders</p>
                <p className="text-[11px] text-muted-foreground mt-0.5">Allow the system to create replenishment tasks when thresholds are crossed.</p>
              </div>
              <button
                type="button"
                onClick={() => setSettings((current) => ({ ...current, auto_po_enabled: !current.auto_po_enabled }))}
                className={`w-10 h-5 rounded-full relative transition-colors ${settings.auto_po_enabled ? "bg-green-500" : "bg-zinc-400/40"}`}
                aria-pressed={settings.auto_po_enabled}
              >
                <span className={`absolute top-0.5 w-4 h-4 bg-white rounded-full transition-all ${settings.auto_po_enabled ? "left-5" : "left-0.5"}`} />
              </button>
            </div>
          </section>
        </div>

        <section className="bg-white/20 dark:bg-white/5 rounded-[2rem] p-6 border border-white/30 dark:border-white/10">
          <div className="flex items-center gap-2 mb-6">
            <RotateCcw className="w-5 h-5 text-foreground" />
            <div>
              <h3 className="font-semibold text-[15px]">Costing & Thresholds</h3>
              <p className="text-[11px] text-muted-foreground mt-0.5">Persisted defaults used by backend workflows and operational screens.</p>
            </div>
          </div>

          <div className="grid grid-cols-3 gap-6">
            <SelectField
              label="Costing Method"
              value={settings.costing_method}
              onChange={(value) => setSettings((current) => ({ ...current, costing_method: value as "WAC" | "FIFO" }))}
              options={[
                { value: "WAC", label: "Weighted average cost" },
                { value: "FIFO", label: "First in, first out" },
              ]}
            />

            <div className="flex flex-col gap-1.5">
              <label className="text-[11px] font-semibold text-foreground">Default Low Stock Threshold</label>
              <span className="text-[10px] text-muted-foreground mb-1">Minimum on-hand quantity before replenishment workflows react</span>
              <input
                type="number"
                min={0}
                value={settings.default_low_stock_threshold}
                onChange={(event) => setSettings((current) => ({ ...current, default_low_stock_threshold: Number(event.target.value) }))}
                className="w-full h-10 px-3 rounded-xl bg-white/40 dark:bg-black/20 border border-black/5 dark:border-white/10 text-[13px] focus:outline-none focus:ring-2 focus:ring-primary/50"
              />
            </div>

            <div className="flex flex-col gap-1.5">
              <label className="text-[11px] font-semibold text-foreground">Audit Status</label>
              <span className="text-[10px] text-muted-foreground mb-1">Every successful save writes an audit event</span>
              <div className="h-10 px-3 rounded-xl bg-white/40 dark:bg-black/20 border border-black/5 dark:border-white/10 text-[13px] flex items-center">
                Enabled
              </div>
            </div>
          </div>
        </section>
      </main>

      <aside className="w-[360px] flex-shrink-0 bg-[#121212] text-white p-6 rounded-3xl relative z-20 flex flex-col h-full shadow-[0_8px_30px_rgb(0,0,0,0.4)] border border-white/5">
        <SettingsInsightsSidebar />
      </aside>
    </>
  );
}

function ReadOnlyField({ label, value }: { label: string; value: string }) {
  return (
    <div className="flex flex-col gap-1.5">
      <label className="text-[11px] font-medium text-muted-foreground">{label}</label>
      <input readOnly value={value} className="w-full h-10 px-3 rounded-xl bg-black/5 dark:bg-white/5 border border-transparent text-[13px] text-muted-foreground focus:outline-none" />
    </div>
  );
}

function SelectField({ label, value, onChange, options }: { label: string; value: string; onChange: (value: string) => void; options: { value: string; label: string }[] }) {
  return (
    <div className="flex flex-col gap-1.5">
      <label className="text-[11px] font-semibold text-foreground">{label}</label>
      <div className="relative">
        <select value={value} onChange={(event) => onChange(event.target.value)} className="w-full h-10 px-3 rounded-xl bg-white/40 dark:bg-black/20 border border-black/5 dark:border-white/10 text-[13px] focus:outline-none appearance-none pr-8">
          {options.map((option) => (
            <option key={option.value} value={option.value}>{option.label}</option>
          ))}
        </select>
        <ChevronDown className="absolute right-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground pointer-events-none" />
      </div>
    </div>
  );
}
