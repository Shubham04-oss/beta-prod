# Service Account for Cloud Run
resource "google_service_account" "cloud_run_sa" {
  account_id   = "synq-cloud-run-sa-${var.environment}"
  display_name = "Synq Cloud Run Service Account"
}

# Secret Manager: Postgres Database URL
resource "google_secret_manager_secret" "db_url" {
  secret_id = "synq-db-url-${var.environment}"
  replication {
    auto {}
  }
}

resource "google_secret_manager_secret_version" "db_url_version" {
  secret      = google_secret_manager_secret.db_url.id
  secret_data = "postgres://dev:${var.db_password}@${oci_core_instance.synq_db_server.public_ip}:5432/synq_db?sslmode=disable"
}

# Secret Manager: Valkey URL
resource "google_secret_manager_secret" "valkey_url" {
  secret_id = "synq-valkey-url-${var.environment}"
  replication {
    auto {}
  }
}

resource "google_secret_manager_secret_version" "valkey_url_version" {
  secret      = google_secret_manager_secret.valkey_url.id
  secret_data = var.valkey_password != "" ? "redis://:${var.valkey_password}@${oci_core_instance.synq_db_server.public_ip}:6379" : "redis://${oci_core_instance.synq_db_server.public_ip}:6379"
}

# IAM Role: Secret Accessor permission for Cloud Run Service Account
resource "google_secret_manager_secret_iam_member" "db_url_accessor" {
  secret_id = google_secret_manager_secret.db_url.secret_id
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:${google_service_account.cloud_run_sa.email}"
}

resource "google_secret_manager_secret_iam_member" "valkey_url_accessor" {
  secret_id = google_secret_manager_secret.valkey_url.secret_id
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:${google_service_account.cloud_run_sa.email}"
}

# IAM Role: Pub/Sub Publisher & Subscriber for Cloud Run
resource "google_project_iam_member" "pubsub_publisher" {
  project = var.gcp_project_id
  role    = "roles/pubsub.publisher"
  member  = "serviceAccount:${google_service_account.cloud_run_sa.email}"
}

resource "google_project_iam_member" "pubsub_subscriber" {
  project = var.gcp_project_id
  role    = "roles/pubsub.subscriber"
  member  = "serviceAccount:${google_service_account.cloud_run_sa.email}"
}

# Pub/Sub Topics
resource "google_pubsub_topic" "pim_events" {
  name = "pim-events-${var.environment}"
}

resource "google_pubsub_subscription" "pim_events_sub" {
  name  = "pim-events-sub-${var.environment}"
  topic = google_pubsub_topic.pim_events.name

  # 10 minutes message retention
  message_retention_duration = "600s"
  retain_acked_messages      = false
  ack_deadline_seconds       = 60
}

# Cloud Run: Ops-API Service (Go API Backend)
resource "google_cloud_run_v2_service" "ops_api" {
  name     = "ops-api-${var.environment}"
  location = var.gcp_region
  ingress  = "INGRESS_TRAFFIC_ALL"

  template {
    service_account = google_service_account.cloud_run_sa.email

    containers {
      image = "us-central1-docker.pkg.dev/${var.gcp_project_id}/synq-repo/ops-api:latest"

      ports {
        container_port = 8080
      }

      resources {
        limits = {
          cpu    = "1"
          memory = "1024Mi"
        }
      }

      # Standard Environment Variables
      env {
        name  = "ENV"
        value = var.environment
      }
      env {
        name  = "PORT"
        value = "8080"
      }
      env {
        name  = "GCLOUD_PROJECT"
        value = var.gcp_project_id
      }
      env {
        name  = "TEMPORAL_HOST_PORT"
        value = "${oci_core_instance.synq_db_server.public_ip}:7233"
      }
      env {
        name  = "UNIFIED_WORKSPACE_ID"
        value = "placeholder-workspace-id"
      }

      # Secrets mounted from GCP Secret Manager
      env {
        name = "DATABASE_URL"
        value_source {
          secret_key_ref {
            secret  = google_secret_manager_secret.db_url.secret_id
            version = "latest"
          }
        }
      }
      env {
        name = "VALKEY_URL"
        value_source {
          secret_key_ref {
            secret  = google_secret_manager_secret.valkey_url.secret_id
            version = "latest"
          }
        }
      }
    }
  }

  # Ensure secrets exist before service deployment
  depends_on = [
    google_secret_manager_secret_version.db_url_version,
    google_secret_manager_secret_version.valkey_url_version,
    google_secret_manager_secret_iam_member.db_url_accessor,
    google_secret_manager_secret_iam_member.valkey_url_accessor
  ]
}

# Cloud Run public invoker IAM policy (allow public access to HTTP endpoints)
resource "google_cloud_run_v2_service_iam_member" "public_invoker" {
  project    = google_cloud_run_v2_service.ops_api.project
  location   = google_cloud_run_v2_service.ops_api.location
  name       = google_cloud_run_v2_service.ops_api.name
  role       = "roles/run.invoker"
  member     = "allUsers"
}

# Custom Domain Mapping (using Cloud Run direct domain CNAME, bypassing ALB)
resource "google_cloud_run_domain_mapping" "ops_api_domain" {
  name      = "api.synq.app"
  location  = var.gcp_region
  project   = var.gcp_project_id

  metadata {
    namespace = var.gcp_project_id
  }

  spec {
    route_name = google_cloud_run_v2_service.ops_api.name
  }
}

# Outputs for GCP infrastructure
output "gcp_cloud_run_url" {
  value       = google_cloud_run_v2_service.ops_api.uri
  description = "The direct serverless URL of the Cloud Run API service"
}
