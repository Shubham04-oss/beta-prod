"use client";

import React, { useState } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Building2, Layers, Mail, Lock, EyeOff, Sparkles, Bot, Package, LineChart, ShieldCheck } from "lucide-react";
import { auth } from "@/lib/firebase";
import { signInWithEmailAndPassword } from "firebase/auth";
import { fetchAPI } from "@/lib/api";

export default function OnboardPage() {
  const [orgName, setOrgName] = useState("");
  const [tenantName, setTenantName] = useState("");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleOnboard = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError(null);
    try {
      // 1. Send the provisioning request to the Go backend
      await fetchAPI("/api/v1/onboard", {
        method: "POST",
        body: JSON.stringify({
          org_name: orgName,
          tenant_name: tenantName,
          admin_email: email,
          admin_password: password,
        })
      });

      // 2. The Go backend successfully created the Global Firebase User and injected Custom Claims.
      // Now, securely sign into the Firebase SDK to trigger the AuthProvider redirect.
      await signInWithEmailAndPassword(auth, email, password);
      
    } catch (err: any) {
      console.error("Onboarding failed:", err);
      setError(err.message || "Failed to provision workspace. Check backend logs.");
      setLoading(false); // Only set loading to false if it failed. If success, AuthProvider redirects.
    }
  };

  return (
    <div className="w-full h-full flex items-center justify-between">
      {/* Left Column: Form */}
      <div className="w-full max-w-md flex flex-col justify-center h-full space-y-8 animate-in fade-in slide-in-from-left-4 duration-1000">
        <div>
          <h1 className="text-3xl font-bold tracking-tight text-slate-900 dark:text-white mb-2">
            Create your workspace 🚀
          </h1>
          <p className="text-sm text-slate-500 dark:text-slate-400">
            Set up your organization to start automating commerce operations.
          </p>
        </div>

        {error && (
          <div className="p-3 rounded-md text-sm bg-red-50 text-red-600 dark:bg-red-950/50 dark:text-red-400">
            {error}
          </div>
        )}

        <form onSubmit={handleOnboard} className="space-y-4">
          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <Label htmlFor="orgName">Organization Name</Label>
              <div className="relative">
                <Building2 className="absolute left-3 top-3 h-4 w-4 text-slate-400" />
                <Input
                  id="orgName"
                  placeholder="e.g. Acme Corp"
                  className="pl-9 h-11"
                  value={orgName}
                  onChange={(e) => setOrgName(e.target.value)}
                  required
                />
              </div>
            </div>
            
            <div className="space-y-2">
              <Label htmlFor="tenantName">Workspace Name</Label>
              <div className="relative">
                <Layers className="absolute left-3 top-3 h-4 w-4 text-slate-400" />
                <Input
                  id="tenantName"
                  placeholder="e.g. Acme Production"
                  className="pl-9 h-11"
                  value={tenantName}
                  onChange={(e) => setTenantName(e.target.value)}
                  required
                />
              </div>
            </div>
          </div>

          <div className="space-y-2">
            <Label htmlFor="email">Admin Email address</Label>
            <div className="relative">
              <Mail className="absolute left-3 top-3 h-4 w-4 text-slate-400" />
              <Input
                id="email"
                type="email"
                placeholder="admin@yourcompany.com"
                className="pl-9 h-11"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                required
              />
            </div>
          </div>

          <div className="space-y-2">
            <Label htmlFor="password">Admin Password</Label>
            <div className="relative">
              <Lock className="absolute left-3 top-3 h-4 w-4 text-slate-400" />
              <Input
                id="password"
                type="password"
                placeholder="Choose a strong password"
                className="pl-9 pr-9 h-11"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                required
                minLength={8}
              />
              <button type="button" className="absolute right-3 top-3">
                <EyeOff className="h-4 w-4 text-slate-400" />
              </button>
            </div>
          </div>

          <Button type="submit" className="w-full h-11 bg-purple-700 hover:bg-purple-800 text-white mt-4" disabled={loading}>
            {loading ? "Provisioning Workspace..." : "Create Workspace"}
          </Button>
        </form>

        <p className="text-center text-sm text-slate-500 mt-8">
          Already have an account?{" "}
          <a href="/login" className="font-medium text-purple-600 hover:text-purple-500">
            Sign in
          </a>
        </p>
      </div>

      {/* Right Column: Features Card overlaying the background */}
      <div className="flex-1 h-full hidden lg:flex items-center justify-end pl-10">
        <div className="w-[480px] bg-black/60 backdrop-blur-2xl border border-white/10 rounded-3xl p-8 text-white space-y-8 animate-in fade-in slide-in-from-right-8 duration-1000 shadow-2xl">
          
          <div className="flex items-start gap-4">
            <div className="p-2 bg-white/10 rounded-lg shrink-0">
              <Sparkles className="w-5 h-5 text-purple-300" />
            </div>
            <div>
              <h3 className="font-semibold text-lg mb-1">Intelligent. Integrated. Impactful.</h3>
              <p className="text-sm text-white/60 leading-relaxed">Aurea AI Commerce Ops unifies your commerce operations with the power of AI.</p>
            </div>
          </div>

          <div className="w-full border-t border-white/10" />

          <div className="flex items-start gap-4">
            <div className="p-2 bg-white/10 rounded-lg shrink-0">
              <Bot className="w-5 h-5 text-purple-300" />
            </div>
            <div>
              <h3 className="font-medium mb-1">AI-Powered Automation</h3>
              <p className="text-sm text-white/60 leading-relaxed">Automate complex tasks and reduce manual effort with intelligent AI agents.</p>
            </div>
          </div>

          <div className="w-full border-t border-white/10" />

          <div className="flex items-start gap-4">
            <div className="p-2 bg-white/10 rounded-lg shrink-0">
              <Package className="w-5 h-5 text-purple-300" />
            </div>
            <div>
              <h3 className="font-medium mb-1">Unified Commerce Operations</h3>
              <p className="text-sm text-white/60 leading-relaxed">Manage PIM, orders, shipments, inventory, and more from a single platform.</p>
            </div>
          </div>

          <div className="w-full border-t border-white/10" />

          <div className="flex items-start gap-4">
            <div className="p-2 bg-white/10 rounded-lg shrink-0">
              <LineChart className="w-5 h-5 text-purple-300" />
            </div>
            <div>
              <h3 className="font-medium mb-1">Real-time Insights</h3>
              <p className="text-sm text-white/60 leading-relaxed">Get real-time visibility and actionable insights to drive better decisions.</p>
            </div>
          </div>

          <div className="w-full border-t border-white/10" />

          <div className="flex items-start gap-4">
            <div className="p-2 bg-white/10 rounded-lg shrink-0">
              <ShieldCheck className="w-5 h-5 text-purple-300" />
            </div>
            <div>
              <h3 className="font-medium mb-1">Secure & Scalable</h3>
              <p className="text-sm text-white/60 leading-relaxed">Enterprise-grade security and scalability built for your growing business.</p>
            </div>
          </div>

        </div>
      </div>
    </div>
  );
}
