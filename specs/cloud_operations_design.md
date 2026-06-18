# Cloud Operations Design: GCP-Native Proactive Multi-Agent System

## Executive Summary
This document outlines the technical design for a proactive, multi-tenant multi-agent system built natively on Google Cloud Platform (GCP). The architecture prioritizes secure SaaS tenant isolation, robust asynchronous event orchestration, elastic scalability, and seamless integration of human oversight via Human-in-the-Loop workflows.

## Multi-Tenant Data Isolation Strategy
A primary concern for a multi-tenant SaaS is preventing cross-tenant data leakage, especially when autonomous agents have access to a shared data store.
- **Service Offering:** Cloud SQL for PostgreSQL or AlloyDB for PostgreSQL.
- **Isolation Mechanism:** We will utilize PostgreSQL Row-Level Security (RLS). Each agent session will be scoped to a specific tenant ID via a temporary session variable established upon database connection. RLS policies will ensure that database queries implicitly append tenant filters, preventing agents from accidentally reading or modifying data belonging to other tenants.
- **Connection Security:** Cloud SQL IAM database authentication will be employed.
- **Strict DB Access Guardrails:** ADK Agents must NEVER write to or query the database directly. All database interactions MUST be routed through the central Go Ops API. The Go Ops API enforces the JWT validation, tenant context injection (for RLS), and business logic. Agents are treated as untrusted clients regarding core infrastructure.

## Scalable Operational Flows
Multi-agent systems exhibit highly variable workloads. We need compute layers that can scale rapidly to handle bursty agent activity and scale down during idle periods.
- **Stateless Agent Execution:** **Google Cloud Run** will serve as the exclusive execution environment for all agent tasks. Cloud Run provides rapid autoscaling (including scale-to-zero), allowing the system to handle concurrent agent operations without over-provisioning.
- **No Kubernetes Requirement:** There are no GKE or stateful pods in this architecture. All agent persistence and memory are durably stored in PostgreSQL between invocations, allowing any stateless Cloud Run container to resume a paused agent workflow.

## Event Orchestration & Communication
Agents must react to asynchronous events (system triggers, external APIs, user interactions) and coordinate with each other efficiently.
- **Event Bus:** Google Cloud Pub/Sub will act as the backbone for agent-to-agent communication and fan-out message distribution.
- **Event Routing:** Eventarc will route system events directly to Cloud Run endpoints, triggering proactive agent behaviors.
- **Complex Agent Workflows:** Cloud Workflows will be used to orchestrate complex multi-step processes involving multiple specialized agents, handling state management, retries, and conditional routing natively.

## Human-in-the-Loop (HITL) Workflows
Not all decisions can or should be fully autonomous. Human-in-the-Loop (HITL) mechanisms are required for high-stakes actions, quality assurance, and edge-case resolution.
- **Workflow Callbacks:** Using Cloud Workflows, an agent can pause its execution stream and wait for human approval. The workflow generates a callback URL and sends it to a human operator (e.g., via Slack, Email, or a web dashboard). Once the operator approves, rejects, or modifies the action via the callback, the workflow resumes and passes the human feedback back to the agent.
- **Stateful Tracking:** Cloud Firestore will act as the real-time state database for HITL requests, allowing front-end dashboards to reactively display pending approvals to human operators.

## Architectural Decisions

1. **Cloud Run exclusively for Agents:** Chosen to minimize operational overhead (Zero-DevOps) and take advantage of rapid scale-from-zero capabilities for unpredictable agent invocations. Kubernetes (GKE) is explicitly avoided.
2. **Postgres RLS over Database-per-Tenant:** RLS provides a strong security boundary at a fraction of the cost and maintenance overhead compared to provisioning separate database instances or managing hundreds of schemas.
3. **Cloud Workflows over Airflow/Composer:** Cloud Workflows provides lower latency for real-time agent orchestration and natively supports the callback patterns necessary for seamless Human-in-the-Loop integration, without the heavy infrastructure of Cloud Composer.

## Edge Cases

- **Agent Infinite Loops:** Malfunctioning proactive agents could trigger infinite execution loops, racking up Cloud Run and LLM API costs. **Mitigation:** Implement strict execution timeouts on Cloud Run, use Pub/Sub dead-letter queues (DLQs), and enforce rate-limiting via Cloud Armor and API Gateway.
- **RLS Policy Bypass:** Misconfigured application code connecting with superuser privileges could bypass RLS. **Mitigation:** Enforce strict least-privilege IAM roles; application execution roles will not have schema ownership or superuser rights.
- **Stale HITL Requests:** Human operators may ignore or miss HITL approval requests, leaving workflows hanging indefinitely. **Mitigation:** Cloud Workflows will implement a maximum wait time (e.g., 24 hours). If the timeout is reached, a default fallback action (such as soft-fail or escalate to an admin) will be executed automatically.

## Production Readiness

To ensure the architecture is ready for enterprise multi-tenant traffic, the following operational measures are required:
- **Observability:** Centralized logging and distributed tracing across all agents using Google Cloud Operations Suite (Cloud Logging, Cloud Trace). Each log entry MUST inject the `tenant_id` and `agent_run_id` for isolated debugging.
- **CI/CD & IaC:** All infrastructure (Cloud SQL, Pub/Sub topics, Cloud Run services) will be managed via Terraform. Deployments will follow a canary rollout strategy using Google Cloud Deploy.
- **Security Posture:** Regular vulnerability scanning of agent container images via Artifact Registry. VPC Service Controls will be enabled to prevent data exfiltration from the managed project environment.
