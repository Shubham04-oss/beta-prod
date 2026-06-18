# Day 2: Frontend Identity & Wiring

## Status: Complete 🟢

## Objectives Completed
1. **Firebase Frontend SDK Initialization**: Connected Next.js successfully to the local Firebase Emulator running on `macmini`. Built a hostname proxy switch to ensure smooth routing when running from `localhost`.
2. **Global `<AuthProvider />`**: Created an overarching React Context that automatically intercepts the `onAuthStateChanged` hook.
   - Decodes the raw Firebase JWT on login.
   - Extracts the custom backend claims (`role`, `tenant_id`, `org_id`).
   - Syncs them natively into React State so the UI instantly knows the user's workspace boundaries.
3. **Login UI Implementation**: Matched the provided "Aurea" scenes perfectly. Built a frosted glass login form sitting inside the global `TimeSyncedBackground`. Hides the navigation header natively via `usePathname` interception.
4. **API Token Interceptor (`src/lib/api.ts`)**: Built a `fetchAPI` wrapper. Automatically pulls a fresh, unexpired `getIdToken()` from Firebase and injects it as `Bearer <token>` for all outgoing HTTP requests targeting the Go backend (`ops-api`). 

## Blockers Resolved
- **Next.js `NODE_ENV` Conflicts**: `npm run build` forced `production` bypass of the emulator connection logic, leading to `auth/user-not-found`. Refactored to use `window.location.hostname` checks instead.
- **Firebase Multi-Tenancy Mismatch**: The Go backend originally placed users inside strict Firebase Identity Platform Tenant Pools. Refactored to create users Globally in Firebase, relying entirely on Postgres + JWT Claims for our strict Multi-Tenant data isolation. This allows a seamless, single-screen login UI without requiring the user to type a "Workspace Slug" first.
- **Strict HTML Hierarchy Error**: React Hydration panicked when `<AuthProvider>` placed a loading spinner directly under `<html>`. Moved providers safely within `<body>`.

## Next Steps
We have successfully completed **Phase 1: The Agentic Skeleton**.
Both the Go backend (RBAC, SQL type-safety) and the Next.js frontend (Firebase Context, Interceptors, Shadcn UI) are perfectly synchronized. We are ready to begin **Phase 2: Product Information Management (PIM)**.
