# Team Workspace (Human-AI Collaboration) Architecture

## 1. Core Vision (The Misunderstanding Corrected)
The "Team Workspace" is **not** a traditional human-to-human collaboration tool like Slack or Google Docs. 

Instead, it is a **Human-AI Agentic Workspace**. An employee acts as the "Core Team Manager" and is tied to a squad of specialized AI Agents (ADK teams). The human delegates tasks, reviews AI outputs, and accomplishes complex workflows with the AI acting as the workforce.

## 2. Architecture & Existing Tech Stack
This architecture builds strictly **ON TOP** of the existing eternal tech stack (Go, Postgres, Temporal, Next.js, GCP).

*   **Database (PostgreSQL via `sqlc`)**: 
    *   `human_ai_teams`: Maps a human `user_id` to an `agent_squad_id`.
    *   `ai_tasks`: Tracks tasks delegated by the human to the AI team, including status, input context, and the AI's proposed output.
*   **Orchestration (Temporal)**: 
    *   Temporal is the eternal orchestrator for the AI Team. When a human manager assigns a massive task (e.g., "Update prices for all winter apparel"), a Temporal workflow manages the execution, dividing sub-tasks among different specialized agents and pausing for Human-in-the-Loop (HITL) approval when necessary.
*   **Backend (Go/chi)**: 
    *   Serves as the broker between the human manager's Next.js dashboard and the Temporal AI workflows.

## 3. Frontend & UI Constraints (STRICT)
**CRITICAL RULE:** The overarching frontend design layout is **FIXED**. We will absolutely **NOT** alter the core layout structure.

*   **Strict Design Language:** We will only polish custom components using existing libraries to respect the current design language.
*   **Right-Sidebar Standard:** Every domain (including this Team Workspace) will utilize the standard right-sidebar drawer for contextual information (e.g., clicking on an AI task opens the right-sidebar showing the AI's execution log or request for approval).
*   **Visual Polish:** We will utilize approved visual effects (like background blur and glassmorphism on modals/sidebars) to ensure the application feels premium without breaking the underlying grid or routing structure.

## 4. User's Point of View (POV) & Customizability
*   **Squad Customization:** The human manager can configure their AI squad, selecting which specialized agents (e.g., "PIM Editor Agent", "Pricing Analyst Agent") are in their team.
*   **Delegation Dashboard:** A centralized view where the manager can see all active tasks being worked on by the AI team.
*   **Approval Gateways (HITL):** Configurable thresholds where the AI must stop and ask the human manager for permission (via the right-sidebar) before committing a destructive action (like deleting a product or issuing a refund).


## Frontend & UI Constraints (STRICT)
**CRITICAL RULE:** The overarching frontend design layout is **FIXED**. We will absolutely **NOT** alter the core layout structure.
*   **Strict Design Language:** We will only polish custom components using existing libraries to respect the current design language.
*   **Right-Sidebar Standard:** Every domain must utilize the standard right-sidebar drawer for contextual information.
*   **Visual Polish:** We will utilize approved visual effects (like background blur and glassmorphism on modals/sidebars) to ensure the application feels premium without breaking the underlying grid or routing structure.
