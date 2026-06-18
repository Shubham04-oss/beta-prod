"use client";

import { SettingsLeftSidebar } from "@/components/SettingsLeftSidebar";
import { SettingsInsightsSidebar } from "@/components/SettingsInsightsSidebar";
import { KeyRound, Webhook, ShieldAlert } from "lucide-react";

export default function SettingsDeveloperPage() {
  return (
    <>
      <SettingsLeftSidebar />

      <main className="flex-1 overflow-y-auto custom-scrollbar pr-4 flex flex-col gap-8 pb-10">
        <div className="mt-2">
          <h1 className="text-3xl font-bold tracking-tight">Developer</h1>
          <p className="text-muted-foreground mt-1">API access and webhook delivery controls.</p>
        </div>

        <section className="bg-white/20 dark:bg-white/5 rounded-[2rem] p-6 border border-white/30 dark:border-white/10 grid grid-cols-2 gap-6">
          <div className="rounded-2xl bg-white/40 dark:bg-black/20 border border-black/5 dark:border-white/10 p-5">
            <div className="flex items-center gap-2 mb-3">
              <KeyRound className="w-5 h-5 text-foreground" />
              <h3 className="font-semibold text-[15px]">API Keys</h3>
            </div>
            <p className="text-[12px] text-muted-foreground leading-relaxed">
              Key generation is disabled until the `api_keys` Postgres schema and GCP Secret Manager storage path are installed.
            </p>
          </div>

          <div className="rounded-2xl bg-white/40 dark:bg-black/20 border border-black/5 dark:border-white/10 p-5">
            <div className="flex items-center gap-2 mb-3">
              <Webhook className="w-5 h-5 text-foreground" />
              <h3 className="font-semibold text-[15px]">Webhooks</h3>
            </div>
            <p className="text-[12px] text-muted-foreground leading-relaxed">
              Inbound Unified.to webhooks are HMAC-verified and persisted to the OMS outbox. User-configurable outbound webhook delivery is pending its durable delivery schema.
            </p>
          </div>
        </section>

        <section className="bg-amber-500/10 rounded-[2rem] p-6 border border-amber-500/20 flex items-start gap-4">
          <ShieldAlert className="w-5 h-5 text-amber-600 dark:text-amber-400 mt-0.5" />
          <div>
            <h3 className="font-semibold text-[15px] text-amber-700 dark:text-amber-300">Security gate active</h3>
            <p className="text-[12px] text-amber-700/80 dark:text-amber-300/80 mt-1">
              This page intentionally avoids creating or displaying secrets until key metadata, scope validation, rotation, revocation, and Secret Manager references are fully available.
            </p>
          </div>
        </section>
      </main>

      <aside className="w-[360px] flex-shrink-0 bg-[#121212] text-white p-6 rounded-3xl relative z-20 flex flex-col h-full shadow-[0_8px_30px_rgb(0,0,0,0.4)] border border-white/5">
        <SettingsInsightsSidebar />
      </aside>
    </>
  );
}
