# Day 4: The Monitoring Grid

**Status:** Completed
**Date:** 2026-06-12

## Tasks Completed
- [x] Disabled text masking (`maskAllInputs: false`) in the Next.js `PostHogProvider` to allow for full AI intent and chat visibility within Session Replays.
- [x] Tied Sentry issues directly to PostHog Session Replays by passing the `posthog_session_id` into `Sentry.setTag` on initialization.
- [x] Configured the Go `ops-api` to push traces directly to Sentry via OTLP (`otlptracehttp`).
- [x] Instrumented the entire `go-chi` router with the `otelchi` middleware to automatically capture metrics on every single API request, including the WebSocket handshake.

## Next Steps
Phase 1 is complete! We are now ready to move into **Phase 2: AI Orchestration & MCP Architecture**, starting with Day 5 (AI Gateway & Tracking) to deploy LiteLLM and Langfuse.
