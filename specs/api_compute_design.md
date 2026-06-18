# API & Compute Layer Design for GCP-Native Proactive Multi-Agent System

## Overview
This document outlines the API and Compute layer architecture for a multi-tenant, GCP-native proactive multi-agent system. The system is designed using a **Go-first approach** for high concurrency, low latency, and efficient compute resource utilization.

## Compute Layer Architecture
The compute layer is composed of the following GCP services, utilizing Go 1.22+ as the primary runtime:
- **Cloud Run**: Hosts the synchronous API layer and stateless agent orchestration routines. Its rapid scaling (scale-to-zero and burst-to-1000s) makes it ideal for fluctuating agent request loads.
- **Google Kubernetes Engine (GKE) Autopilot**: Used for persistent, stateful, or long-running proactive agents that require background loops (e.g., watching a stream of events).
- **Cloud Pub/Sub & Cloud Tasks**: Enable asynchronous, proactive behaviors. Cloud Tasks handles delayed/scheduled agent invocations, while Pub/Sub drives event-driven agent reactions.
- **Go Routines & Channels**: Within each container, Go's native concurrency model is heavily utilized to dispatch sub-agent tasks and stream partial responses back to the client.

## API Routing and Design
The API is exposed via the **GCP Global External Application Load Balancer** utilizing Serverless Network Endpoint Groups (NEGs) to route incoming HTTP/gRPC requests directly to the appropriate Cloud Run services. Heavyweight API Gateways (like Apigee or GCP API Gateway) are explicitly avoided to eliminate redundant per-request costs and bottlenecks.
- **Protocol**: gRPC for inter-service communication; REST/JSON for external clients (using grpc-gateway for Go).
- **Routing & Edge Security**: The Load Balancer provides path-based routing (e.g., mapping `/api/v1/ops/*` to the Ops API container). It natively integrates with Cloud Armor (WAF) to absorb DDoS attacks and block malicious traffic at the edge. 
- **Validation**: Because API Gateway is bypassed, JWT validation and rate limiting are handled directly within the Go middleware and Cloud Tasks, streamlining the architecture.

### Context and Isolation (`tenant_id`, `org_id`, `role`)
Multi-tenancy and data isolation are strictly enforced at the API boundary and propagated through the compute layer via Go `context.Context`.

1. **Authentication & Identity Extraction**:
   The Go API receives the request directly from the Load Balancer. A custom Go middleware interceptor validates the JWT (minted via Firebase Auth or Cloud Identity) using Google's public JWKS certificates and extracts the claims (`tenant_id`, `org_id`, `role`).
2. **Context Propagation**:
   The Go middleware injects these identifiers into the context:
   ```go
   type authKey struct{}
   type AuthContext struct {
       TenantID string
       OrgID    string
       Role     string
   }
   func WithAuth(ctx context.Context, auth *AuthContext) context.Context {
       return context.WithValue(ctx, authKey{}, auth)
   }
   ```
   Every subsequent compute function, database query, and downstream service call receives this `ctx`.
3. **Data Isolation (Row-Level Security)**:
   When making queries to Cloud Spanner or Cloud SQL, the Go data access layer appends `tenant_id` and `org_id` to all queries, effectively enforcing logical isolation.
4. **Role-Based Access Control (RBAC)**:
   The `role` extracted from the context is validated against policy definitions (e.g., using Casbin or Google Cloud IAM bindings) before an agent can execute a sensitive tool.

## Architectural Decisions

- **Go-First Approach**: Go's low memory footprint and high concurrency are ideal for orchestrating thousands of simultaneous agent interactions, allowing a single Cloud Run container to manage many active agent sessions.
- **Logical Data Isolation vs. Physical Isolation**: We opted for logical isolation using `tenant_id` and `org_id` over physical database-per-tenant, to reduce infrastructure overhead. Strict enforcement via Go middleware and Spanner Row-Level Security mitigates data bleed risks.
- **Proactive Execution Engine**: We rely on Cloud Tasks and Pub/Sub rather than persistent idle loops. A "scheduler" agent publishes a task to Cloud Tasks, which invokes a webhook on our Cloud Run Go service. This ensures we only pay for compute when agents are actively processing.
- **Context-Aware Logging**: Using `slog` in Go, all log entries automatically extract `tenant_id` and `org_id` from the context, ensuring operational visibility at the tenant level.

## Edge Cases

- **Agent Runaway Loops**: Proactive agents might trigger infinite loops (e.g., continually responding to their own outputs). We mitigate this by appending a `HopCount` to the Go context, which increments on each asynchronous task creation. If `HopCount` exceeds a threshold, execution is terminated.
- **Context Cancellation & Timeout**: If a user disconnects or a global timeout is reached, the Go context is cancelled (`ctx.Done()`). Our agent orchestration logic must listen to this channel to gracefully halt LLM generation and tool execution, preventing orphaned compute cycles.
- **Cross-Tenant Confused Deputy**: A compromised or hallucinating agent might attempt to access another tenant's data. Because data access layers enforce `tenant_id` from the secure Go context (which the agent cannot modify), the query will safely fail or return empty.

## Production Readiness

- **Observability**: OpenTelemetry is integrated natively into the Go services. Distributed tracing tracks a request from API Gateway -> Cloud Run -> Pub/Sub -> Cloud Functions. Spans are enriched with `tenant_id` to trace per-tenant latency.
- **Resiliency & Retries**: All inter-service gRPC calls and GCP API invocations use exponential backoff and retry mechanisms, leveraging Go libraries like `cenkalti/backoff`.
- **Load Testing**: Synthetic traffic simulates 10,000+ concurrent proactive agents waking up simultaneously to ensure the GKE and Cloud Run autoscalers perform optimally under thundering-herd scenarios.
- **Security Scanning**: Automated CI/CD pipelines run `govulncheck` and static analysis (e.g., `gosec`) to prevent injection vulnerabilities in dynamic SQL or agent tool executions.
- **Deployment**: We use Cloud Deploy for progressive rollouts (Canary), routing 5% of traffic for a specific `tenant_id` to new Go binary versions before full promotion.
