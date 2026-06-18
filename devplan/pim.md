# Product Information Management (PIM) Architecture & Development Plan

## 1. Executive Summary
This document outlines the architectural blueprint for the Product Information Management (PIM) domain. Designed strictly from a User Point of View (POV), the system prioritizes extreme customizability, allowing merchants and brands to define their own complex product taxonomies, schemas, and UI layouts. The architecture is explicitly built **on top of our existing, highly scalable tech stack**: Go (chi), PostgreSQL (pgvector, sqlc), Temporal, Next.js (App Router), Firebase Auth, Unified.to, and GCP Native services.

## 2. The User POV: Designing for Deep Customizability
A successful PIM must bend to the user's business model. Users should never feel constrained by a rigid schema; the PIM must adapt to them.
- **User-Defined Schemas & Attributes:** Users can define custom attributes (text, numbers, booleans, rich text, references, assets) globally, per category, or per product family. No hardcoded limitations.
- **Flexible Taxonomies:** Multi-dimensional categorization. Products can exist in hierarchical categories (e.g., Clothing > Men > Shirts) and non-hierarchical user-defined tags/collections (e.g., "Summer Collection 2026").
- **Dynamic UI Configurations:** The Next.js frontend will render dynamic forms based entirely on user-defined schemas. Users can customize list views, save complex filters as "Smart Views", and configure which attributes are prominent.
- **Robust UI/UX Settings:** Deep personalization of the workspace. Users can define their own dashboard layouts, localized data entry preferences, and personalized workflows. The interface must feel like a customizable workbench.

## 3. Technology Stack Alignment
The architecture leverages our core stack to its maximum potential without introducing redundant technologies:
- **Backend:** Go (chi router) for a high-performance, strongly-typed REST API.
- **Database:** PostgreSQL for ACID guarantees. We utilize `JSONB` for schema-less attribute flexibility, and `pgvector` for AI-powered semantic product search.
- **Orchestration:** Temporal is our eternal orchestrator, serving as the central nervous system for all stateful workflows, sagas, and long-running background tasks.
- **Frontend:** Next.js (App Router) for an SEO-friendly, highly responsive SPA, secured by Firebase Auth.
- **Integrations:** Unified.to for standardized, unified connectivity to external e-commerce platforms, ERPs, and POS systems.
- **Cloud Infrastructure:** GCP Native (Pub/Sub, Eventarc, Secret Manager, Memorystore, Cloud Run) for a scalable, event-driven backbone.

## 4. Data Architecture: Postgres JSONB vs. Traditional EAV
We will strictly avoid the performance and maintenance pitfalls of traditional Entity-Attribute-Value (EAV) tables by fully leveraging **Postgres JSONB**.

### Core Data Models
1.  **`attribute_schemas` & `product_templates`:** Defines the expected structure, validation rules, and UI hints for different product types, serving as the blueprint for the dynamic frontend.
2.  **`products` & `product_variants`:**
    *   Standard columns for essential routing/indexing data: `id`, `sku`, `created_at`, `tenant_id`, `status`.
    *   `attributes` (`JSONB`): Stores all user-defined fields.
    *   **Performance:** We will utilize GIN (Generalized Inverted Index) on the `attributes` `JSONB` column to enable sub-millisecond querying across millions of dynamically defined fields (e.g., `SELECT * FROM products WHERE attributes @> '{"color": "red"}'`).
3.  **`taxonomies` & `categories`:** Using a materialized path (`ltree` extension) approach for lightning-fast querying of deep, user-defined category trees.
4.  **`pgvector` Integration:** An `embeddings` table linked to products will store vector embeddings of product descriptions and key attributes, enabling "fuzzy" semantic search, AI-driven duplicate detection, and automated categorization.

## 5. Workflow Orchestration with Temporal
Temporal is the backbone of the PIM's asynchronous and distributed capabilities. We will use Temporal extensively for:

- **Bulk Imports & Exports:** Processing massive CSV/Excel uploads or ERP syncs. Temporal workflows will handle file parsing, data validation, chunking, rate-limiting against the DB, and reporting progress streams back to the UI.
- **Syndication Sagas (Unified.to Sync):** When a product is updated, a Temporal Saga will orchestrate the propagation of these changes to downstream channels (Shopify, Amazon, BigCommerce) via Unified.to. Temporal guarantees eventual consistency, handling API rate limits, backoffs, and compensation logic on failures.
- **Bulk Mutations:** Executing complex, long-running rules like "Apply a 10% discount and add a 'Sale' tag to all products matching a specific dynamic filter."
- **Data Enrichment Pipelines:** Coordinating background tasks like auto-generating product descriptions via LLMs, optimizing media assets, or translating attributes into multiple languages.

## 6. Backend API (Go + chi + sqlc)
- **Domain-Driven Design (DDD):** The Go backend will be structured around domains (Products, Taxonomies, Assets). Go provides the concurrency and performance needed for heavy payload processing.
- **sqlc Integration:** We will use `sqlc` for type-safe query generation. For dynamic `JSONB` updates, we will leverage carefully constructed `sqlc` functions to safely mutate JSONB documents without race conditions (utilizing `jsonb_set` and concurrent-safe patches).
- **Idempotency:** All mutation endpoints, especially those triggering Temporal workflows, will be strictly idempotent to safely interact with UI retries and distributed workflow triggers.

## 7. Frontend Architecture (Next.js)
- **Schema-Driven UI:** The application will dynamically generate forms based on the `attribute_schemas` fetched from the backend. We will use robust state management (like React Hook Form combined with Zod) for dynamic validation.
- **State Management:** React Query for server state, caching, and optimistic UI updates, combined with Zustand for local UI preferences (e.g., column visibility, sidebar state, saved filters).
- **Customizable Data Grid:** A highly performant data grid (e.g., AG Grid or TanStack Table) allowing users to drag-and-drop columns, pin fields, perform inline edits, and filter deeply into JSONB attribute structures.

## 8. Integration Strategy (Unified.to & GCP)
- **Unified.to:** Acts as our universal abstraction layer. We will map our core JSONB product structure to Unified.to's standard E-commerce/Inventory models, enabling seamless omnichannel synchronization without building bespoke integrations.
- **Event-Driven Architecture:** Core entities will emit events (e.g., `ProductUpdated`, `TaxonomyCreated`) to GCP Pub/Sub via Eventarc. This decouples the core PIM from downstream services like search indexing, cache invalidation, or analytics pipelines.

## 9. Phased Implementation Plan
- **Phase 1: Foundation.** Postgres schema design (JSONB, ltree), core Go CRUD APIs, and Next.js scaffolding with Firebase Auth.
- **Phase 2: Deep Customizability.** Implementation of the Schema Builder UI, dynamic forms, and user-defined taxonomies.
- **Phase 3: Orchestration.** Temporal integration for bulk CSV import/export, asset processing, and robust job tracking UI.
- **Phase 4: Omnichannel & AI.** Unified.to integration via Temporal Sagas for downstream syndication and `pgvector` integration for semantic search.


## Frontend & UI Constraints (STRICT)
**CRITICAL RULE:** The overarching frontend design layout is **FIXED**. We will absolutely **NOT** alter the core layout structure.
*   **Strict Design Language:** We will only polish custom components using existing libraries to respect the current design language.
*   **Right-Sidebar Standard:** Every domain must utilize the standard right-sidebar drawer for contextual information.
*   **Visual Polish:** We will utilize approved visual effects (like background blur and glassmorphism on modals/sidebars) to ensure the application feels premium without breaking the underlying grid or routing structure.
