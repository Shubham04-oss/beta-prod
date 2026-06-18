# Event-Driven Architecture Specification

## Overview
This document outlines the definitive event-based architecture for the Synq GCP-native system. It details the mechanisms used to decouple microservices, trigger autonomous agents, and handle human-in-the-loop (HITL) workflows using Google Cloud Platform (GCP) services and Temporal.

## Core Components

*   **Cloud Pub/Sub:** The asynchronous event backbone of the system. It handles the ingestion, routing, and delivery of domain events across various microservices.
*   **Eventarc:** The event routing service that connects Pub/Sub events to specific compute endpoints, primarily Cloud Run.
*   **Cloud Run (`agent-server`):** Serverless compute containers that host our autonomous agents. They are designed to scale to zero when idle and wake up instantly upon receiving an HTTP request triggered by an event.
*   **Temporal:** A microservice orchestration platform used to manage complex, long-running agent workflows and state transitions, including HITL approvals.

## Architecture Deep Dive

### 1. The Asynchronous Event Backbone: Cloud Pub/Sub
Cloud Pub/Sub serves as the central nervous system for our microservices. By acting as a decoupling layer, it ensures that producers of events (e.g., an Order Service) do not need to know about the consumers of those events (e.g., a Notification Service, or autonomous agents).

#### Publishing Domain Events
When a significant state change occurs within a domain, the responsible microservice publishes a domain event to a designated Pub/Sub topic.

**Example: `OrderCreated`**
1.  A user places an order via the API.
2.  The Order Service successfully persists the order to the database.
3.  The Order Service immediately publishes an `OrderCreated` event to the `orders-topic`.
    *   *Payload:* Contains essential details like `order_id`, `customer_id`, `timestamp`, and `total_amount`.
    *   *Attributes:* Used for message filtering (e.g., `event_type="OrderCreated"`, `region="us-central1"`).

### 2. Waking Up Agents: Eventarc and Cloud Run
Our autonomous agents reside within Cloud Run containers (`agent-server`). To optimize costs and resources, these containers scale to zero when there is no activity. We use Eventarc to bridge the gap between Pub/Sub events and these serverless containers.

#### The Eventarc Push Mechanism
1.  **Subscription:** Eventarc creates a push subscription to the relevant Pub/Sub topics (e.g., `orders-topic`).
2.  **Filtering:** Eventarc can filter events based on attributes, ensuring an agent only wakes up for events it is designed to handle.
3.  **Triggering:** When a matching event arrives in the Pub/Sub topic, Eventarc pushes the event payload via an HTTP POST request to the configured Cloud Run `agent-server` endpoint.
4.  **Wake From Zero:** If the `agent-server` has zero active instances, Cloud Run immediately spins up a new container to handle the incoming request. The agent processes the event, initiates any necessary actions, and then the container can eventually scale back down if no further requests arrive.

### 3. Workflow Management and HITL: Temporal
While Pub/Sub handles the immediate, fire-and-forget event routing, many agent tasks involve complex, multi-step processes that may take hours or days, particularly those requiring human intervention. Temporal is used to manage these long-running, stateful workflows.

#### Orchestrating Autonomous Agents
When an `agent-server` is triggered by Eventarc, it often kicks off a Temporal workflow rather than executing the entire process synchronously.
1.  **Workflow Initiation:** The `agent-server` starts a Temporal workflow execution, passing the event payload as input.
2.  **Activity Execution:** The Temporal workflow orchestrates various "Activities" (e.g., calling external APIs, performing calculations, updating databases). These activities are executed by Temporal workers (which can also be hosted on Cloud Run or GKE).
3.  **State Management:** Temporal automatically persists the state of the workflow at every step. If a worker crashes, the workflow can resume exactly where it left off on another worker.

#### Human-in-the-Loop (HITL) via Pub/Sub Callbacks
Some automated processes require explicit human approval before proceeding (e.g., executing a high-value trade, sending a sensitive email). Temporal, combined with Pub/Sub, gracefully handles these HITL states.

1.  **Workflow Pauses for Approval:** When the workflow reaches a stage requiring approval, it uses Temporal's `Signals` feature to pause execution and wait for external input.
2.  **Notification Generation:** The workflow (or an activity within it) publishes an `ApprovalRequested` event to a dedicated Pub/Sub topic.
3.  **Human Action:** A frontend system or notification service consumes this event and alerts the designated human user.
4.  **Pub/Sub Callback:** Once the human user approves or rejects the action via the UI, the frontend publishes an `ApprovalResponse` event back to another Pub/Sub topic.
5.  **Workflow Resumption:** A listener service consumes the `ApprovalResponse` event and translates it into a Temporal Signal directed at the specific, paused workflow execution.
6.  **Workflow Continuation:** The Temporal workflow receives the signal, evaluates the human response (approve/reject), and continues down the appropriate execution path.

## Summary of Interaction Flow

1.  **Producer:** Service A publishes `DomainEvent` to Pub/Sub.
2.  **Router:** Eventarc pushes `DomainEvent` to Cloud Run.
3.  **Consumer (Agent):** Cloud Run (`agent-server`) wakes up from zero, receives the event, and starts a Temporal Workflow.
4.  **Orchestrator:** Temporal manages the complex task and persists state.
5.  **HITL Pause:** Temporal pauses and publishes `ApprovalRequested` to Pub/Sub.
6.  **Human Input:** User acts, and the UI publishes `ApprovalResponse` to Pub/Sub.
7.  **Callback:** A listener service signals the Temporal Workflow to resume.
8.  **Completion:** Temporal completes the workflow.
