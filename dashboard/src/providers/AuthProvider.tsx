"use client";

import React, { createContext, useContext, useEffect, useState } from "react";
import { onAuthStateChanged, User } from "firebase/auth";
import { auth } from "@/lib/firebase";
import { useRouter, usePathname } from "next/navigation";

interface AuthContextType {
  user: User | null;
  loading: boolean;
  role: string | null;
  tenantId: string | null;
  orgId: string | null;
  dbUserId: string | null;
}

const AuthContext = createContext<AuthContextType>({
  user: null,
  loading: true,
  role: null,
  tenantId: null,
  orgId: null,
  dbUserId: null,
});

export const useAuth = () => useContext(AuthContext);

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);
  const [role, setRole] = useState<string | null>(null);
  const [tenantId, setTenantId] = useState<string | null>(null);
  const [orgId, setOrgId] = useState<string | null>(null);
  const [dbUserId, setDbUserId] = useState<string | null>(null);
  
  const router = useRouter();
  const pathname = usePathname();

  useEffect(() => {
    const unsubscribe = onAuthStateChanged(auth, async (firebaseUser) => {
      if (firebaseUser) {
        setUser(firebaseUser);
        
        try {
          // Force refresh to get the latest custom claims
          const idTokenResult = await firebaseUser.getIdTokenResult(true);
          const claims = idTokenResult.claims;
          
          setRole(claims.role as string || null);
          setTenantId(claims.tenant_id as string || claims.tenantid as string || null);
          setOrgId(claims.org_id as string || claims.orgid as string || null);
          setDbUserId(claims.db_uid as string || null);
          
          // Store token for client-side API requests
          const token = await firebaseUser.getIdToken();
          // TODO: Move token to HttpOnly secure cookie to prevent XSS exfiltration
          if (typeof window !== 'undefined') {
             window.sessionStorage.setItem('synqAuthToken', token);
          }

          if (pathname === '/login') {
            router.push('/');
          }
        } catch (err) {
          console.error("Failed to parse custom claims", err);
        }
      } else {
        setUser(null);
        setRole(null);
        setTenantId(null);
        setOrgId(null);
        setDbUserId(null);
        
        if (typeof window !== 'undefined') {
          window.sessionStorage.removeItem('synqAuthToken');
        }
        
        if (pathname !== '/login') {
          router.push('/login');
        }
      }
      setLoading(false);
    });

    return () => unsubscribe();
  }, [pathname, router]);

  // If we are loading and not on the login page, we can show a spinner. 
  // For the login page itself, we shouldn't block the UI rendering if loading.
  if (loading && pathname !== '/login') {
    return <div className="h-screen w-screen flex items-center justify-center text-white">Loading Aurea...</div>;
  }

  return (
    <AuthContext.Provider value={{ user, loading, role, tenantId, orgId, dbUserId }}>
      {children}
    </AuthContext.Provider>
  );
}
