"use client";

import { useEffect, Suspense } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import { Loader2 } from "lucide-react";
import { fetchAPI } from "@/lib/api";

function CallbackHandler() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const connectionId = searchParams.get("connectionId");
  const provider = searchParams.get("provider") || searchParams.get("integration_type") || "";
  const category = searchParams.get("category") || "commerce";

  useEffect(() => {
    if (!connectionId) {
      router.replace("/settings/integrations");
      return;
    }

    fetchAPI("/api/v1/integrations/callback", {
      method: "POST",
      body: JSON.stringify({ 
        connection_id: connectionId,
        category,
        provider,
      }),
    }).then(() => {
      console.log("Successfully connected ID:", connectionId);
      router.replace("/settings/integrations");
    }).catch(err => {
      console.error("Failed to save connection:", err);
      router.replace("/settings/integrations");
    });
  }, [category, connectionId, provider, router]);

  return (
    <div className="min-h-screen flex flex-col items-center justify-center bg-[#f7f7f8] dark:bg-zinc-950">
      <div className="flex flex-col items-center gap-4">
        <Loader2 className="w-10 h-10 animate-spin text-primary" />
        <h2 className="text-xl font-bold">Securing Connection...</h2>
        <p className="text-sm text-muted-foreground">Please wait while we verify your integration.</p>
      </div>
    </div>
  );
}

export default function CallbackPage() {
  return (
    <Suspense fallback={
      <div className="min-h-screen flex flex-col items-center justify-center bg-[#f7f7f8] dark:bg-zinc-950">
        <Loader2 className="w-10 h-10 animate-spin text-primary" />
      </div>
    }>
      <CallbackHandler />
    </Suspense>
  );
}
