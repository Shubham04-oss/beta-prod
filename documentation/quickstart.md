# Quickstart Guide

This guide will help you get Synq running locally or prepare it for staging deployment.

## Prerequisites

Before starting, ensure you have the following installed on your machine:
* [Docker](https://www.docker.com/) and Docker Compose
* [Go](https://go.dev/) (version 1.26+)
* [Node.js](https://nodejs.org/) (version 18+ for the Next.js Dashboard)
* [gcloud CLI](https://cloud.google.com/sdk/gcloud) (configured and authenticated)

## 1. Configure Environment Variables

Create a `.env` file in `services/ops-api/` with the following variables:

```env
ENV=development
PORT=8080
DATABASE_URL=postgres://dev:dev@localhost:5432/synq_db?sslmode=disable
VALKEY_URL=redis://localhost:6379
TEMPORAL_HOST_PORT=localhost:7233
UNIFIED_WORKSPACE_ID=your_unified_workspace_id
UNIFIED_TO_TOKEN=your_unified_to_token
UNIFIED_WEBHOOK_SECRET=your_unified_webhook_secret
FIREBASE_AUTH_EMULATOR_HOST=localhost:9099
GCLOUD_PROJECT=your_gcp_project_id
```

## 2. Start the Local Service Stack

We provide a Docker Compose configuration that launches all required stateful services (`postgres`, `valkey`, `temporal`, and `temporal-ui`).

Run the following command from the root of the repository:

```bash
docker-compose -f infrastructure/oracle-docker-compose.yml up -d
```

This starts:
* **Postgres** on port `5432`
* **Valkey** on port `6379`
* **Temporal Server** on port `7233`
* **Temporal Web UI** on port `8080` (mapped locally)

## 3. Run Database Migrations

Apply database schemas and Row Level Security policies:

```bash
go run services/ops-api/apply_schema.go
```

## 4. Launch the Go API Server

Navigate to the `services/ops-api` directory and run the API:

```bash
cd services/ops-api
go run cmd/server/main.go
```

The API will start listening on port `8080`. You should see logs confirming connection to both PostgreSQL and Valkey:

```stdout
Successfully established connection to Valkey/Redis at localhost:6379
Ops API starting on port 8080...
```

## 5. Launch the Next.js Dashboard

Navigate to the `dashboard` directory, install dependencies, and start the frontend:

```bash
cd dashboard
npm install
npm run dev
```

Open [http://localhost:3000](http://localhost:3000) in your browser to interact with the dashboard.
