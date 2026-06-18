# Vercel Project for the Frontend Dashboard
resource "vercel_project" "synq_dashboard" {
  name      = "synq-dashboard-${var.environment}"
  framework = "nextjs"

  git_repository = {
    type = "github"
    repo = var.vercel_github_repo
  }

  # Build commands and output directories
  root_directory = "dashboard"
}

# Environment Variable: API URL pointing to GCP Cloud Run
resource "vercel_project_environment_variable" "api_url" {
  project_id = vercel_project.synq_dashboard.id
  key        = "NEXT_PUBLIC_API_URL"
  value      = google_cloud_run_v2_service.ops_api.uri
  target     = ["production", "preview", "development"]
}

# Environment Variable: WS URL for Real-Time Event Stream
resource "vercel_project_environment_variable" "ws_url" {
  project_id = vercel_project.synq_dashboard.id
  key        = "NEXT_PUBLIC_WS_URL"
  # Replace HTTPS with WSS / HTTP with WS for WebSockets
  value      = replace(replace(google_cloud_run_v2_service.ops_api.uri, "https://", "wss://"), "http://", "ws://")
  target     = ["production", "preview", "development"]
}

# Environment Variable: Environment Type
resource "vercel_project_environment_variable" "env" {
  project_id = vercel_project.synq_dashboard.id
  key        = "NEXT_PUBLIC_ENV"
  value      = var.environment
  target     = ["production", "preview", "development"]
}

# Environment Variable: Firebase Project ID (for client SDK auth mapping)
resource "vercel_project_environment_variable" "firebase_project" {
  project_id = vercel_project.synq_dashboard.id
  key        = "NEXT_PUBLIC_FIREBASE_PROJECT_ID"
  value      = var.gcp_project_id
  target     = ["production", "preview", "development"]
}

# Outputs for Vercel deployment
output "vercel_dashboard_url" {
  value       = "https://${vercel_project.synq_dashboard.name}.vercel.app"
  description = "The default Vercel URL of the Dashboard project"
}
