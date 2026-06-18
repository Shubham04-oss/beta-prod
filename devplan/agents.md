# Enterprise AI Agents Domain (ADK & A2UI) Development Plan

## 1. Executive Summary & User POV Focus
This document outlines the architecture and development plan for the AI Agents domain, comprising the Agent Development Kit (ADK) and the Agent-to-User Interface (A2UI). We are building our agents directly onto our robust, existing infrastructure to ensure maximum reliability and seamless product integration.

The core philosophy centers on empowering the end-user with deep customizability, flexible configurations, and granular control over agent behavior and interactions. We are moving away from black-box AI toward transparent, steerable, and highly personalized intelligent partners.

### User Empowerment Core Principles
*   **Customizable Personas:** Users can intuitively define and tweak agent personas, tones, and domain expertise through intuitive UI settings in the Next.js frontend, moving beyond static system prompts.
*   **Granular Guardrails:** User-configurable safety, ethical, and compliance guardrails. Users define the exact boundaries and constraints of their agents.
*   **Human-in-the-Loop (HITL) by Design:** Configurable approval workflows via Temporal sagas allowing users to define exactly which destructive, high-stakes, or third-party integration actions require explicit human consent.
*   **Flexible A2UI:** Deeply customizable workspace, chat layouts, streaming preferences, and interaction modalities, giving users a robust UI/UX experience tailored to their workflow.

## 2. Core Architecture & Tech Stack

Our approach mandates extending our current stable tech stack rather than introducing overlapping systems. We lean into our existing ecosystem:

*   **Backend Services (ADK):** **Go (with chi router)**. Go provides the high-concurrency, low-latency foundation required for the Agent Development Kit (ADK), handling HTTP/WebSocket connections, prompt assembly, and API brokering.
*   **Orchestration & State Machine:** **Temporal**. Temporal is our eternal orchestrator. It manages all long-running agent workflows, state machines, tool executions, and sagas. 
*   **Database & Vector Store (RAG):** **PostgreSQL (with `pgvector` and `sqlc`)**. PostgreSQL acts as the single source of truth. We use `pgvector` to store and query embeddings natively alongside relational data, eliminating the need for a separate vector database. Data access is strictly typed via `sqlc`.
*   **Frontend (A2UI):** **Next.js (App Router)** with **Firebase Auth**. The React-based frontend delivers the Agent-to-User Interface (A2UI) with server-side rendering for speed and Firebase for identity management.
*   **Integrations:** **Unified.to** for standardized, seamless integrations with third-party APIs (CRMs, ATS, Ticketing), governed by agent logic.
*   **Foundation Models:** **Google Gemini API** (Pro for reasoning, Flash for speed) accessed natively from Go.
*   **Cloud Infrastructure:** **GCP Native**. We leverage Pub/Sub and Eventarc for asynchronous messaging, Secret Manager for credentials, Memorystore for fast caching, and Cloud Run/GKE for compute.

## 3. Agent Development Kit (ADK) in Go

The ADK is written in Go to maximize concurrency and performance.

### 3.1. Go-Powered Core
*   **Prompt Engine:** A Go module that constructs dynamic system prompts by fetching user persona configurations from Postgres (via `sqlc`) and dynamically injecting relevant context.
*   **Tool Registry:** Go interfaces define tools. Before an agent executes a tool via Unified.to or internal APIs, the Go backend validates user permissions and configured guardrails.

### 3.2. Temporal for Agent Memory and Orchestration
Temporal is the backbone of our agent architecture. We do not use ephemeral memory hacks; we use Temporal's durable state.
*   **Eternal Workflows as Agents:** An agent's lifecycle is modeled as a long-running Temporal workflow. This ensures that agent state, conversation history, and pending actions survive service restarts.
*   **Human-in-the-Loop (HITL) via Signals:** When a user sets a guardrail requiring manual approval, the Temporal workflow pauses and waits for a specific Temporal Signal (approved/rejected) triggered from the Next.js frontend.
*   **Sagas for Tool Execution:** Multi-step agent actions (e.g., fetching data from Unified.to, formatting it, and sending an email) are executed as Temporal Sagas, ensuring atomic rollbacks and robust error handling.

### 3.3. RAG with PostgreSQL (`pgvector` + `sqlc`)
*   **Episodic Memory & Context:** All user documents, past conversations, and structured data are embedded via Gemini and stored in PostgreSQL using `pgvector`.
*   **Type-Safe Retrieval:** We use `sqlc` to generate type-safe Go code for querying `pgvector`, performing rapid cosine similarity searches (`ORDER BY embedding <-> $1`) to inject highly relevant context into the Gemini API.

## 4. Agent-to-User Interface (A2UI) Deep Dive

The Next.js (App Router) frontend focuses heavily on giving the user an adaptable viewport into the agent's mind.

### 4.1. Robust UI/UX Customization
*   **Persona Builder:** A rich Next.js UI where users can define agent traits using sliders, multi-select tags, and explicit instruction blocks.
*   **Modular Workspace:** Users can arrange chat widgets, tool execution logs, and active context viewers in a customizable grid layout.
*   **Real-Time Streaming:** The Go backend streams Gemini responses via WebSockets or SSE directly to Next.js UI components.

### 4.2. Granular Guardrails & HITL Configuration
*   **Action Scopes Interface:** A settings panel where users map specific Unified.to integrations or internal APIs to Allow/Deny lists or "Require Approval" states.
*   **Approval Inbox:** A dedicated UI section showing paused Temporal workflows awaiting user consent. The UI presents exactly what the agent intends to do (e.g., "Draft an email via Unified.to") and allows the user to approve, modify, or reject the payload.

### 4.3. Transparency & Trust
*   **"Show Agent Reasoning":** A toggleable UI pane that exposes the agent's chain-of-thought, current Temporal state, and raw tool invocations, demystifying the black box.

## 5. Implementation Roadmap

1.  **Phase 1: Foundation & Temporal Integration.** Scaffold the Go (chi) ADK and connect it to Temporal. Establish basic long-running workflows representing simple agents. Set up Next.js frontend with Firebase Auth.
2.  **Phase 2: Database & RAG.** Implement `pgvector` schemas in PostgreSQL. Generate `sqlc` queries for semantic search. Integrate Gemini API for embeddings and text generation in Go.
3.  **Phase 3: The User POV UI/UX.** Build the Persona Builder and Action Scopes configuration screens in Next.js. Tie these settings into the Go prompt builder.
4.  **Phase 4: Integrations & HITL.** Connect Unified.to for external tools. Implement Temporal Signals to handle the "Review & Approve" HITL workflows directly from the A2UI.
5.  **Phase 5: Streaming & Polish.** Implement low-latency WebSocket streaming from Go to Next.js. Refine the modular workspace and "Show Agent Reasoning" panels. Deploy on Cloud Run/GKE with Pub/Sub eventing.


## Frontend & UI Constraints (STRICT)
**CRITICAL RULE:** The overarching frontend design layout is **FIXED**. We will absolutely **NOT** alter the core layout structure.
*   **Strict Design Language:** We will only polish custom components using existing libraries to respect the current design language.
*   **Right-Sidebar Standard:** Every domain must utilize the standard right-sidebar drawer for contextual information.
*   **Visual Polish:** We will utilize approved visual effects (like background blur and glassmorphism on modals/sidebars) to ensure the application feels premium without breaking the underlying grid or routing structure.
