# Settings (Organization/Tenant) Domain - Development Plan

## 1. Executive Summary & User POV

This document outlines the architecture and development plan for the Settings (Organization/Tenant) domain. From a User's Point of View (POV), the Settings domain is the command center for their organization. It must feel empowering, secure, and deeply customizable. Users expect a seamless self-serve experience for complex tasks like SAML/SSO setup, fine-grained control over API keys, and comprehensive visibility into their organization's activity through customizable audit logs.

## 2. Core Technologies

Our architecture is strictly built on top of our existing robust tech stack:
*   **Frontend**: Next.js (App Router) with Firebase Auth for a highly responsive, modern UI/UX.
*   **Backend**: Go (chi router) for performant, type-safe API endpoints.
*   **Database**: PostgreSQL (with pgvector and `sqlc`) for relational data integrity and tenant configuration storage.
*   **Orchestration**: Temporal for reliable, fault-tolerant background processes, state machines, and complex sagas.
*   **Integrations**: Unified.to for seamless third-party connections.
*   **Cloud Infrastructure**: GCP Native (Secret Manager for sensitive data, Pub/Sub, Eventarc, Memorystore, Cloud Run/GKE).

## 3. Architecture & Data Model

### 3.1. Tenant Configuration (PostgreSQL & Go)
We will leverage **PostgreSQL** to store tenant/organization settings. Using `sqlc`, we ensure type-safe queries within our Go backend.
*   **`organizations` table**: Core tenant info (ID, name, billing tier).
*   **`organization_settings` table**: Will utilize a `JSONB` column for flexible, deeply customizable user preferences (e.g., UI themes, default notification channels, timezone). A JSONB approach allows us to iterate rapidly on new user-facing settings without requiring frequent schema migrations.
*   **`organization_members` tables**: Roles and RBAC (Role-Based Access Control) mapping.

### 3.2. User-Facing API Key Management (GCP Secret Manager)
Users need the ability to generate, rotate, and revoke API keys with granular, scope-based permissions.
*   **Storage Mechanism**: We will use **GCP Secret Manager** as the system of record for the actual cryptographic secret material.
*   **Database Mapping**: An `api_keys` table in PostgreSQL will store the key metadata (Key ID, Name, Scopes, Expiration Date, Last Used Timestamp, Created By). The actual key secret will *never* be stored in plaintext in the DB; instead, it will reference the GCP Secret Manager version payload.
*   **Backend Validation**: The Go backend will validate API requests by checking the provided key against the PostgreSQL database (for fast scope and revocation checks) and securely retrieving the secret material from GCP Secret Manager when necessary.

### 3.3. Self-Serve SAML/SSO Setup (Firebase Auth & Go)
Enterprise customers demand self-serve SAML/SSO to manage their users.
*   **UI/UX**: The Next.js frontend will provide a step-by-step wizard for administrators to input their IdP (Identity Provider) details (Entity ID, SSO URL, X.509 Certificate).
*   **Backend Flow**: The Go backend will receive these details and programmatically configure the respective SAML provider within **Firebase Auth** via the Firebase Admin SDK.
*   **Validation**: Provide a "Test Connection" UI flow that temporarily stashes configurations in Memorystore before finalizing the Firebase Auth setup to ensure a smooth user experience.

### 3.4. Customizable Audit Log Exports (Temporal & GCP Pub/Sub)
Administrators require deep visibility into who did what and when within their tenant.
*   **Event Generation**: The Go backend will emit audit events (e.g., "API Key Created", "SSO Configured", "User Role Changed") to **GCP Pub/Sub**.
*   **Storage**: Events will be ingested into PostgreSQL for fast, paginated UI viewing.
*   **Export Workflows**: Users can configure automated or one-off exports (e.g., "Export the last 30 days of logs as a CSV", "Send a weekly audit summary"). We will use **Temporal** to orchestrate these export jobs. Temporal will manage the retries, state, data fetching, formatting, and reliable delivery of these potentially large data payloads to user-specified destinations.

## 4. Feature Breakdown & User Experience

### 4.1. Workspace & Robust UI/UX Settings
*   **User POV**: "I want the application to map perfectly to my team's workflow and branding."
*   **Implementation**: The Next.js frontend fetches `organization_settings` (the JSONB payload) from the Go API. This powers deep UI customizability, including default landing pages, data table view preferences, dark/light mode overrides, and custom branding options.

### 4.2. API Keys & Granular Scopes
*   **User POV**: "I need to give a specific script access to read data, but I want to guarantee it absolutely cannot write or delete data."
*   **Implementation**: A dedicated UI in Next.js for managing keys. Users select granular scopes (e.g., `data:read`, `billing:none`). The Go backend generates a secure token, registers it in GCP Secret Manager, stores the metadata in Postgres, and returns the key *only once* to the user in the UI.

### 4.3. Integrations Hub
*   **User POV**: "I need to connect my external tools seamlessly without reading API documentation."
*   **Implementation**: Leverage **Unified.to** for handling third-party OAuth flows. The Next.js frontend redirects users through the Unified.to flow. The resulting connection IDs and configurations are securely stored and managed via our Go API.

## 5. Development Phases

1.  **Phase 1: Foundation & Data Layer**
    *   Define the `sqlc` schema and queries for `organizations` and `organization_settings`.
    *   Set up the Go (chi) API CRUD endpoints for retrieving and updating settings.
    *   Build the Next.js Settings layout, navigation, and basic UI components.
2.  **Phase 2: API Keys & Security**
    *   Implement the GCP Secret Manager integration within the Go backend.
    *   Build the UI for API key generation, granular scope selection, and revocation.
    *   Implement API middleware in Go to validate incoming keys against allowed scopes.
3.  **Phase 3: Identity & Access (SSO)**
    *   Build the SAML/SSO configuration UI wizard in Next.js.
    *   Implement Go backend logic to interface with the Firebase Auth API for SAML setup.
4.  **Phase 4: Audit Logs & Temporal Workflows**
    *   Instrument the Go API with Pub/Sub audit event emitters.
    *   Develop Temporal worker code for processing, formatting, and delivering customizable data exports.
    *   Build the Audit Log viewer and export configuration UI in Next.js.


## Frontend & UI Constraints (STRICT)
**CRITICAL RULE:** The overarching frontend design layout is **FIXED**. We will absolutely **NOT** alter the core layout structure.
*   **Strict Design Language:** We will only polish custom components using existing libraries to respect the current design language.
*   **Right-Sidebar Standard:** Every domain must utilize the standard right-sidebar drawer for contextual information.
*   **Visual Polish:** We will utilize approved visual effects (like background blur and glassmorphism on modals/sidebars) to ensure the application feels premium without breaking the underlying grid or routing structure.
