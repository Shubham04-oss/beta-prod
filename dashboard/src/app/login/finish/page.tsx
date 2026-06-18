"use client";

import React, { useEffect, useState, Suspense } from "react";
import { isSignInWithEmailLink, signInWithEmailLink } from "firebase/auth";
import { auth } from "@/lib/firebase";
import { useRouter, useSearchParams } from "next/navigation";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";

function MagicLinkHandler() {
  const [email, setEmail] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [needsEmail, setNeedsEmail] = useState(false);
  const router = useRouter();

  useEffect(() => {
    // Only run if it's a valid link
    if (isSignInWithEmailLink(auth, window.location.href)) {
      // Try to get email from local storage
      let savedEmail = window.localStorage.getItem("emailForSignIn");
      
      if (!savedEmail) {
        // User opened the link on a different device. Ask for their email.
        setNeedsEmail(true);
      } else {
        // We have the email, proceed with sign in
        completeSignIn(savedEmail);
      }
    } else {
      setError("This magic link is invalid or has expired.");
    }
  }, []);

  const completeSignIn = async (signInEmail: string) => {
    try {
      await signInWithEmailLink(auth, signInEmail, window.location.href);
      // Clear email from storage
      window.localStorage.removeItem("emailForSignIn");
      // AuthProvider will automatically push to '/'
    } catch (err: any) {
      console.error("Magic link sign-in error", err);
      setError(err.message || "Failed to sign in with magic link.");
    }
  };

  const handleEmailSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (email) {
      completeSignIn(email);
    }
  };

  return (
    <div className="w-full h-full flex items-center justify-center">
      <div className="w-full max-w-md bg-white/80 dark:bg-black/60 backdrop-blur-xl p-8 rounded-3xl shadow-xl border border-white/20 dark:border-white/10 space-y-6">
        <h1 className="text-2xl font-bold tracking-tight text-slate-900 dark:text-white">
          Authenticating...
        </h1>

        {error ? (
          <div className="bg-red-50 text-red-600 dark:bg-red-950/50 dark:text-red-400 p-4 rounded-xl text-sm">
            {error}
            <Button variant="link" className="px-0 block mt-2 text-red-700" onClick={() => router.push('/login')}>
              Return to Login
            </Button>
          </div>
        ) : needsEmail ? (
          <form onSubmit={handleEmailSubmit} className="space-y-4">
            <p className="text-sm text-slate-500">
              You opened this link on a different device. Please confirm your email address to continue.
            </p>
            <div className="space-y-2">
              <Label htmlFor="confirm-email">Email address</Label>
              <Input
                id="confirm-email"
                type="email"
                placeholder="Enter your email"
                className="h-11"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                required
              />
            </div>
            <Button type="submit" className="w-full h-11 bg-purple-700 hover:bg-purple-800 text-white">
              Complete Sign In
            </Button>
          </form>
        ) : (
          <div className="flex flex-col items-center justify-center py-6">
            <div className="w-8 h-8 border-4 border-purple-600 border-t-transparent rounded-full animate-spin"></div>
            <p className="mt-4 text-sm text-slate-500">Securely logging you in...</p>
          </div>
        )}
      </div>
    </div>
  );
}

// Next.js requires useSearchParams to be wrapped in Suspense
export default function FinishLogin() {
  return (
    <Suspense fallback={<div className="text-white">Loading...</div>}>
      <MagicLinkHandler />
    </Suspense>
  );
}
