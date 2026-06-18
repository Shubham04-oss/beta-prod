# Master Execution Plan: Core Infra & Backend Event-First

## Scope

In scope:
- OMS and fulfillment from `devplan/order.md`
- Sales Channel and Unified.to integrations from `devplan/sales_channel.md`
- Team Workspace and RBAC from `devplan/team_workspace.md`
- Settings, configs, and audit logs from `devplan/settings.md`

Out of scope:
- `inventory.md`
- `pim.md`
- `agents.md`
- AI/ADK/catalog work beyond the Team Workspace tenant/RBAC shell requested by the in-scope plan

## Current Execution Order

1. Read the four in-scope plans and the master execution prompt.
2. Inspect existing Go/chi, Postgres/sqlc, Temporal, Pub/Sub, and Next.js surfaces.
3. Delegate independent exploration slices and log every subagent outcome.
4. Stabilize backend compile blockers and remove stubs/hardcoding in the in-scope paths.
5. Implement OMS tenant-scoped API and workflow correctness.
6. Implement Sales Channel API and Unified.to connection persistence/listing.
7. Wire Next.js Orders and Channels pages to real APIs while preserving the fixed layout and right sidebar pattern.
8. Implement Settings tenant config API and integrate real settings/integrations flows.
9. Implement Team Workspace tenant/RBAC API and replace static Teams UI with real tenant-scoped data.
10. Execute targeted Go tests, dashboard lint/build checks, and remote Mac Mini Docker verification when reachable.
11. Record residual risks and blocked infra validations in `status/codex/`.

## Completed In This Run

- Created `status/codex/`.
- Logged all completed or errored subagent outcomes, including backend/schema, dashboard UI, remote infra, Team Workspace/RBAC, settings/audit, contract review, final verification, channel sidebar, and second backend verification passes.
- Loaded `devplan/master_execution_prompt.md`.
- Fixed OMS compile blocker.
- Replaced OMS invalid schema writes with valid `order_status` enum values and real payment state transitions.
- Removed `time.Sleep` payment stub from OMS.
- Added org+tenant-scoped OMS list/detail APIs.
- Added actual inventory reservation derivation from order items with variant/location IDs and org-scoped reservation/release queries.
- Replaced placeholder fulfillment/return workflows with Temporal workflows that call real activities, verify backing Postgres fulfillment/return rows, update order status, emit order events, and write outbox events.
- Hardened Pub/Sub event publishing to fail if identity pillars are missing.
- Hardened OMS outbox forwarding to publish `DomainEvent` envelopes with tenant/org/user/role and aggregate attributes.
- Added tenant-scoped sales channel list/create APIs.
- Replaced hardcoded Unified provider persistence with provider verification/fallback.
- Added org+tenant-scoped Unified connection listing.
- Hardened Unified legacy webhook to verify active `commerce_connections` and carry system tenant/org context into Temporal.
- Replaced hardcoded Unified fulfillment tracking payload with real `fulfillments` carrier/tracking data and provider-aware passthrough behavior.
- Removed hardcoded Unified polling environment fallback; inbound polling now requires `UNIFIED_ENV` and uses connection-qualified workflow IDs.
- Persisted public Unified webhook receipts to the OMS outbox with org/tenant context.
- Fixed settings integrations callback syntax error.
- Replaced settings integrations mock timeout with real API fetching.
- Replaced orders hard-coded UI with real OMS API data and contextual right sidebar.
- Replaced channels hard-coded UI with real channels/integrations API data and live left/right sidebar state.
- Added org+tenant-scoped tenant settings GET/PUT API.
- Fixed organization invite RBAC role-case mismatch.
- Added Team Workspace Postgres schema with forced RLS, composite org/tenant team/task constraints, and protected API handlers.
- Replaced static Teams UI with real `/api/v1/team-workspace` data and empty states.
- Added settings audit and developer pages wired into settings sidebar navigation.
- Added idempotent `infrastructure/postgres/core_org_tenant_rls.sql` to force org+tenant RLS for in-scope OMS, channels, settings, Unified, and audit tables.
- Applied Team Workspace and core org+tenant RLS migrations to remote Mac Mini Docker-hosted Postgres via SSH using `/usr/local/bin/docker`.
- Removed fake server bootstrap defaults for Unified workspace/token and GCP Pub/Sub project; startup now requires real environment values.

## Verification So Far

- `GOCACHE=/private/tmp/synq-go-cache go test ./cmd/server ./internal/api ./internal/middleware ./internal/oms ./internal/unified ./internal/pubsub ./internal/telemetry ./internal/service` passes in `services/ops-api` with remote Postgres network access.
- `npm run build` passes in `dashboard`.
- Targeted ESLint passes for modified dashboard files.
- Remote Postgres RLS verification shows `relrowsecurity=true` and `relforcerowsecurity=true` for `ai_tasks`, `audit_events`, `commerce_connections`, `human_ai_teams`, `orders`, `sales_channels`, and `tenant_settings`.
- Full `go test ./...` still fails only in pre-existing/out-of-scope scratch/PIM surfaces: root and scratch duplicate `main` declarations, `cmd/test-flow` duplicate declarations, and `internal/pim/service_test.go` test signature.

## Active Delegations

- None active.
