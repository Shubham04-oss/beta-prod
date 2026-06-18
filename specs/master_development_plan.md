# Master Development Plan: AI Operating System Blueprint (2026 Edition)

This master plan reflects the **Core Architecture Pivot**, deeply validated and cross-checked against the absolute bleeding-edge 2026 industry standards by four specialized research sub-agents. 

We have shifted from a monolithic e-commerce application to an **AI Operating System**. Commerce, PIM, and Orders are logically separated packages operating within a high-performance **Modular Monolith** API, working asynchronously with a dedicated standalone **Agent Microservice**.

## 🚨 MUST READ: The Core Security & Data Constraints 🚨

To ensure absolute security, multi-tenant isolation, and type-safety throughout the AI Operating System, the following architectural invariants are strictly enforced:


1. **The Four Pillars of Context (JWT Claims):** Every authenticated request (Human or Agent) *must* carry and extract exactly four contextual identifiers into the Go middleware: `tenantid`, `orgid`, `userid`, and `role`. If any of these are missing, the request is permanently rejected (403 Forbidden).
2. **Compile-Time SQL Type Safety (`sqlc`):** Raw `db.Exec()` SQL strings are forbidden in the Go backend. Every single database query must be written in pure `.sql` files and compiled via **`sqlc`** to auto-generate type-safe Go interfaces, guaranteeing that we never pass an incorrect type (like a string instead of a `pgtype.UUID`) to Postgres.

For deep-dives into these exact mechanisms, the following specification documents are mandatory reading:
- 🛡️ **[Security Practices (RLS & sqlc)](file:///Users/shubham/Projects/synq/specs/security_practices.md)**
- ⚡ **[Event-Driven Architecture (Temporal & Pub/Sub)](file:///Users/shubham/Projects/synq/specs/event_driven_architecture.md)**

## Pre-Launch Checklists (Validated 2026 Standards)

### 1. Functional & Product Checklist
- [ ] **Agent & Human Identity Separation:** **Firebase Auth** for Human authentication (now viable due to simplified custom claims), and **GCP Native Agent Identity** for Agents. The Go backend maps JWT `user_id`s to internal Agent IDs.
- [ ] **Authorization (Standard RBAC):** Simplified Role-Based Access Control instead of ReBAC. Basic roles (Admin, Editor, Viewer) and `tenant_id` fit safely within Firebase Auth's 1000-byte Custom Claims limit.
- [ ] **Headless Project Management (Plane.so):** Kanban boards powered via Plane's headless REST APIs. Webhooks notify the OS of task state changes.
- [ ] **Intelligent Notifications (Novu):** The definitive OSS choice for digesting agent alerts to prevent spam, utilizing the Next.js `<Inbox />` component.
- [ ] **Search (Typesense):** Global, typo-tolerant Command-K searching. (Chosen due to native Algolia-API drop-in component compatibility).
- [ ] **Billing (Stripe):** Metered usage tracking and webhooks via strict signature verification.

### 2. Technical & Operations Checklist
- [ ] **Infrastructure Edge:** ALB + Cloud Armor (Adaptive Protection). Implement **Model Armor** at the edge to sanitize prompts (jailbreaks/PII) *before* they hit the Go backend.
- [ ] **A2UI Hybrid Streaming & Secure WebSockets:** 
  - **Backend:** **Go WebSockets** for stateful, persistent bi-directional communication.
  - **Auth (The BFF Fix):** Next.js BFF generates short-lived **WebSocket Tickets** to securely bridge the HTTP-only cookie session into the Go WebSocket connection without exposing raw JWTs.
  - **Frontend:** **Native React WebSocket Hook** (`useAgentStream`) consuming the ticket. Eliminates the Vercel AI SDK entirely, natively parsing ADK JSON intents and mapping them directly to frontend UI components.
  - **Protocol:** ADK 1.4 **AG-UI protocol**, streaming structured declarative A2UI (Agent-to-User Interface) intents (JSON) instead of raw text, rendering native React components progressively.
- [ ] **AI Orchestration (Go ADK):** 3-Layer **Progressive Disclosure** (Discovery -> Activation -> Execution) to minimize token overhead when loading Personas and Skills.
- [ ] **Native MCP Server Architecture:** Standalone Cloud Run containers exposing the **Model Context Protocol (MCP)**. The Go ADK 1.4 connects directly to external MCP tools via WebSockets, natively passing the context and dynamic tenant API keys.
- [ ] **Native AI (Vertex AI):** Gemini 1.5 Flash/Pro, Veo, and Imagen 3 directly via private GCP VPC. Implement **Multimodal RAG** using Vertex AI Search.
- [ ] **AI Gateway & Observability:** **Vertex AI Model Router (Go ADK Native) + GCP Trace** (Pivot from LiteLLM/Langfuse to maximize Go ADK native compatibility and reduce proxy overhead).
- [ ] **Microservices Boundary (Modular Monolith + Agent Server):** We strictly utilize a two-service Cloud Run deployment model. `ops-api` operates as a high-performance **Modular Monolith** containing all business logic (Commerce, PIM, Identity) compiled into a single binary, preventing network gRPC overhead and simplifying shared `sqlc` database transaction pooling. The `agent-server` operates as a completely independent, asynchronous **Agent Microservice** dedicated to heavy Temporal and Go ADK reasoning loops. 
- [ ] **Event Routing & Orchestration Pattern:**
  - **Temporal:** The centralized "Brain" orchestrating the internal agent loop, pacing, rate-limiting, and HITL pauses (via Temporal `Signals`).
  - **Eventarc Advanced:** The "Nervous System" used purely to emit lifecycle events to decoupled external domains (like billing or analytics).
- [ ] **The Monitoring Grid:** **OpenTelemetry (OTel)** collector hub piping `gen_ai.*` metrics. **Sentry** natively ingests OTLP for debugging, and syncs directly into **PostHog** Session Replays.

### 3. Frontend Domain Data Integration (Dynamic Sidebar Binding)
The modular plugins must bind real-time API data to their respective domain sidebars to eliminate static placeholders:
- [ ] **Orders Domain (`LeftSidebar.tsx`)**: Replace static order counts (Pending, Shipped, etc.) with real-time aggregates from the Orders Plugin.
- [ ] **AI Teams Domain (`TeamsLeftSidebar.tsx`)**: Replace static AI Teams with live Agent registries. Bind the "Team Utilization" chart to live OTel metrics to reflect real Agent efficiency.
- [ ] **PIM Domain (`PIMLeftSidebar.tsx`)**: Bind the "Data Quality" completeness score and validation issues to the PIM Plugin's live data auditing engine.
- [ ] **Channels Domain (`ChannelsLeftSidebar.tsx`)**: Replace the static "Channel Health" SVG chart with live status aggregations from external integrations (e.g., Unified.to connection states).
- [ ] **Analytics & Settings Domains**: Dynamically filter available reports and configuration panels based on standard RBAC permissions for the current user.

---

## The "Velocity Mandate" Tracker

We execute under the strict rule of **One Full Domain Per Day**. Every day represents a complete vertical slice shipped across the Database, Go API, and Next.js UI.

### Phase 1: Identity, Security, & Architecture Core (✅ COMPLETE)
*   **Day 1 & Day 2 (Identity & Front/Back Wiring):** Implement Firebase Auth (Humans) and GCP Native Agent Identity. Establish standard RBAC middleware parsing simplified Custom Claims directly from the JWT. Connect Next.js to Firebase and implement API Interceptors. (✅ UI + API + DB)
*   **Day 2 (Edge & Guardrails):** Configure ALB + Cloud Armor. Implement Model Armor to sanitize AI inputs/outputs at the edge. (✅ UI + API + DB) *(Note: Cloud Edge infrastructure is bypassed for local development)*
*   **Day 3 (A2UI Hybrid Streaming & Secure Auth):** Implement the Go WebSockets server and the native React `useAgentStream` hook. Build the Ticket-Based Auth exchange in Next.js to secure the socket. Configure ADK 1.4 AG-UI declarative JSON intent mapping. (✅ UI + API + DB)
*   **Day 4 (The Monitoring Grid):** Set up the OTel Collector Hub. Route `gen_ai.*` spans to Sentry (via OTLP) and PostHog for AI analytics and Session Replay integration. (✅ UI + API + DB)

### Phase 2: AI Orchestration & MCP Architecture
*   **Day 5 (AI Gateway & Tracking):** Initialize the Go ADK native Vertex AI Model Router. Connect standard OTel tracing directly to ADK tools and LLM steps, eliminating Python proxy dependencies (Gemini 3.5 Flash primary). (✅ UI + API + DB)
*   **Day 6 (Go ADK Personas):** Implement the 3-Layer Progressive Disclosure pattern within the Go ADK. Build the Persona Router. (UI + API + DB)
*   **Day 7 (Native Multi-Tenant MCP Servers):** Deploy stateless Cloud Run MCP servers. Implement Go ADK 1.4 native WebSocket connections to external Model Context Protocol (MCP) servers with dynamic tenant key injection. (UI + API + DB)
*   **Day 8 (Native Multimodal):** Connect Vertex AI Gemini 1.5 Flash/Pro, Veo, and Imagen 3. Implement Interleaved Prompting (native byte streams) and Multimodal RAG. (UI + API + DB)

### Phase 3: Modular Monolith Ecosystem & OSS Integrations
*   **Day 9 (Commerce & Inventory Core):** Implement the Commerce and Inventory modules directly within the `ops-api` **Modular Monolith**. Utilize standard Go interfaces for strict package boundaries while sharing the same high-speed `pgx/v5` connection pool and RLS context. (UI + API + DB)
*   **Day 10 (External Sync):** Integrate Unified.to via their Go SDK. Use the pass-through architecture and pre-built embedded Authorization UI. (UI + API + DB)
*   **Day 11 (Headless Kanban):** Deploy self-hosted Plane.so. Query the REST endpoints and map `state` IDs to Next.js draggable components. (UI + API + DB)
*   **Day 12 (Notifications):** Integrate Novu for intelligent alert digesting. Embed the `@novu/nextjs` `<Inbox />` UI component into the frontend. (UI + API + DB)
*   **Day 13 (Global Search):** Deploy Typesense. Implement asynchronous DB syncing via background workers and expose Search-Only API Keys to the Next.js `react-instantsearch` frontend. (UI + API + DB)
*   **Day 14 (Billing Plugin):** Integrate Stripe metered usage tracking and webhook processing (`stripe-go` with strict signature verification). (UI + API + DB)

### Phase 4: Eventing, Safety, & Launch
*   **Day 15 (Eventarc Choreography):** Wire Eventarc Advanced to act as the decoupling nervous system, triggering domain events asynchronously when Temporal completes workflows. (UI + API + DB)
*   **Day 16 (Temporal HITL):** Implement Temporal `Signals` for indefinite Human-in-the-Loop "durable wait" states without consuming idle compute. (UI + API + DB)
*   **Day 17 (Chaos & Performance Testing):** Integrate ChaosKit (code-level) and Toxiproxy (network-level) to simulate latency and panics. Verify Cloud Armor limits and LiteLLM fallbacks. (UI + API + DB)
*   **Day 18 (Public Beta Launch):** Final deployment via CI/CD. Target <200ms API latency and sub-3s page loads. Activate the PostHog + Sentry live monitoring grid. (UI + API + DB)
