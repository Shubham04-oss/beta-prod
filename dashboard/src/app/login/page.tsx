"use client";

import React, { useState } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Checkbox } from "@/components/ui/checkbox";
import { Mail, Lock, EyeOff, Sparkles, Bot, Package, LineChart, ShieldCheck, ArrowLeft, Wand2 } from "lucide-react";
import { auth } from "@/lib/firebase";
import { signInWithEmailAndPassword, GoogleAuthProvider, signInWithPopup, sendPasswordResetEmail, sendSignInLinkToEmail } from "firebase/auth";

type AuthMode = "signin" | "forgot" | "magic";

export default function LoginPage() {
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [loading, setLoading] = useState(false);
  const [mode, setMode] = useState<AuthMode>("signin");
  const [message, setMessage] = useState<{ text: string; type: "success" | "error" } | null>(null);

  const handleAction = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setMessage(null);
    try {
      if (mode === "signin") {
        await signInWithEmailAndPassword(auth, email, password);
        // AuthProvider handles redirect
      } else if (mode === "forgot") {
        await sendPasswordResetEmail(auth, email);
        setMessage({ text: "Password reset email sent! Check your terminal if using the emulator.", type: "success" });
      } else if (mode === "magic") {
        const actionCodeSettings = {
          url: window.location.origin + '/login/finish',
          handleCodeInApp: true,
        };
        await sendSignInLinkToEmail(auth, email, actionCodeSettings);
        window.localStorage.setItem('emailForSignIn', email);
        setMessage({ text: "Magic link sent! Check your terminal if using the emulator.", type: "success" });
      }
    } catch (error: any) {
      console.error("Auth action failed:", error);
      setMessage({ text: error.message || "An error occurred. Check emulator logs.", type: "error" });
    } finally {
      setLoading(false);
    }
  };

  const handleGoogleSignIn = async () => {
    const provider = new GoogleAuthProvider();
    try {
      await signInWithPopup(auth, provider);
    } catch (error) {
      console.error("Google sign-in failed:", error);
    }
  };

  return (
    <div className="w-full h-full flex items-center justify-between">
      {/* Left Column: Form */}
      <div className="w-full max-w-md flex flex-col justify-center h-full space-y-8 animate-in fade-in slide-in-from-left-4 duration-1000">
        <div>
          <h1 className="text-3xl font-bold tracking-tight text-slate-900 dark:text-white mb-2">
            {mode === "signin" && "Welcome back 👋"}
            {mode === "forgot" && "Reset Password"}
            {mode === "magic" && "Magic Link Login"}
          </h1>
          <p className="text-sm text-slate-500 dark:text-slate-400">
            {mode === "signin" && "Sign in to access your AI Commerce Operations platform."}
            {mode === "forgot" && "Enter your email and we'll send you a link to reset your password."}
            {mode === "magic" && "Enter your email to receive a passwordless sign-in link."}
          </p>
        </div>

        {message && (
          <div className={`p-3 rounded-md text-sm ${message.type === "success" ? "bg-emerald-50 text-emerald-600 dark:bg-emerald-950/50 dark:text-emerald-400" : "bg-red-50 text-red-600 dark:bg-red-950/50 dark:text-red-400"}`}>
            {message.text}
          </div>
        )}

        <form onSubmit={handleAction} className="space-y-5">
          <div className="space-y-2">
            <Label htmlFor="email">Email address</Label>
            <div className="relative">
              <Mail className="absolute left-3 top-3 h-4 w-4 text-slate-400" />
              <Input
                id="email"
                type="email"
                placeholder="Enter your email"
                className="pl-9 h-11"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                required
              />
            </div>
          </div>

          {mode === "signin" && (
            <div className="space-y-2">
              <div className="flex items-center justify-between">
                <Label htmlFor="password">Password</Label>
                <button type="button" onClick={() => { setMode("forgot"); setMessage(null); }} className="text-sm font-medium text-purple-600 hover:text-purple-500">
                  Forgot password?
                </button>
              </div>
              <div className="relative">
                <Lock className="absolute left-3 top-3 h-4 w-4 text-slate-400" />
                <Input
                  id="password"
                  type="password"
                  placeholder="Enter your password"
                  className="pl-9 pr-9 h-11"
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  required
                />
                <button type="button" className="absolute right-3 top-3">
                  <EyeOff className="h-4 w-4 text-slate-400" />
                </button>
              </div>
            </div>
          )}

          {mode === "signin" && (
            <div className="flex items-center space-x-2">
              <Checkbox id="remember" />
              <Label htmlFor="remember" className="text-sm font-normal text-slate-600 cursor-pointer">
                Remember me
              </Label>
            </div>
          )}

          <Button type="submit" className="w-full h-11 bg-purple-700 hover:bg-purple-800 text-white" disabled={loading}>
            {loading ? "Processing..." : mode === "signin" ? "Sign in" : mode === "forgot" ? "Send Reset Link" : "Send Magic Link"}
          </Button>

          {mode !== "signin" && (
            <Button type="button" variant="ghost" className="w-full h-11" onClick={() => { setMode("signin"); setMessage(null); }}>
              <ArrowLeft className="w-4 h-4 mr-2" /> Back to Login
            </Button>
          )}
        </form>

        <div className="relative">
          <div className="absolute inset-0 flex items-center">
            <div className="w-full border-t border-slate-200 dark:border-slate-800" />
          </div>
          <div className="relative flex justify-center text-sm">
            <span className="bg-white/50 dark:bg-black/50 px-2 text-slate-500">or</span>
          </div>
        </div>

        {mode === "signin" && (
          <div className="space-y-3">
            <Button variant="outline" className="w-full h-11 bg-white/50 dark:bg-black/50" onClick={() => { setMode("magic"); setMessage(null); }} type="button">
              <Wand2 className="w-5 h-5 mr-2 text-purple-600" />
              Sign in with Magic Link
            </Button>
            <Button variant="outline" className="w-full h-11 bg-white/50 dark:bg-black/50" onClick={handleGoogleSignIn} type="button">
              <img src="https://www.svgrepo.com/show/475656/google-color.svg" alt="Google" className="w-5 h-5 mr-2" />
              Continue with Google
            </Button>
            <Button variant="outline" className="w-full h-11 bg-white/50 dark:bg-black/50" type="button">
              <img src="https://www.svgrepo.com/show/452062/microsoft.svg" alt="Microsoft" className="w-5 h-5 mr-2" />
              Continue with Microsoft
            </Button>
          </div>
        )}

        {mode === "signin" && (
          <p className="text-center text-sm text-slate-500 mt-8">
            Don't have an account?{" "}
            <a href="/onboard" className="font-medium text-purple-600 hover:text-purple-500">
              Sign up your organization
            </a>
          </p>
        )}
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
