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

# Vercel Project Custom Domain Mapping
resource "vercel_project_domain" "dashboard_domain" {
  project_id = vercel_project.synq_dashboard.id
  domain     = "dashboard.${var.custom_domain}"
}

# Environment Variable: API URL pointing to GCP Cloud Run
resource "vercel_project_environment_variable" "api_url" {
  project_id = vercel_project.synq_dashboard.id
  key        = "NEXT_PUBLIC_API_URL"
  value      = "https://${google_cloud_run_domain_mapping.ops_api_domain.name}"
  target     = ["production", "preview", "development"]
}

# Environment Variable: WS URL for Real-Time Event Stream
resource "vercel_project_environment_variable" "ws_url" {
  project_id = vercel_project.synq_dashboard.id
  key        = "NEXT_PUBLIC_WS_URL"
  value      = "wss://${google_cloud_run_domain_mapping.ops_api_domain.name}"
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

# Environment Variable: Firebase API Key
resource "vercel_project_environment_variable" "firebase_api_key" {
  project_id = vercel_project.synq_dashboard.id
  key        = "NEXT_PUBLIC_FIREBASE_API_KEY"
  value      = var.firebase_api_key
  target     = ["production", "preview", "development"]
}

# Environment Variable: Firebase Auth Domain
resource "vercel_project_environment_variable" "firebase_auth_domain" {
  project_id = vercel_project.synq_dashboard.id
  key        = "NEXT_PUBLIC_FIREBASE_AUTH_DOMAIN"
  value      = var.firebase_auth_domain
  target     = ["production", "preview", "development"]
}

# Environment Variable: Firebase Storage Bucket
resource "vercel_project_environment_variable" "firebase_storage_bucket" {
  project_id = vercel_project.synq_dashboard.id
  key        = "NEXT_PUBLIC_FIREBASE_STORAGE_BUCKET"
  value      = var.firebase_storage_bucket
  target     = ["production", "preview", "development"]
}

# Environment Variable: Firebase Messaging Sender ID
resource "vercel_project_environment_variable" "firebase_messaging_sender_id" {
  project_id = vercel_project.synq_dashboard.id
  key        = "NEXT_PUBLIC_FIREBASE_MESSAGING_SENDER_ID"
  value      = var.firebase_messaging_sender_id
  target     = ["production", "preview", "development"]
}

# Environment Variable: Firebase App ID
resource "vercel_project_environment_variable" "firebase_app_id" {
  project_id = vercel_project.synq_dashboard.id
  key        = "NEXT_PUBLIC_FIREBASE_APP_ID"
  value      = var.firebase_app_id
  target     = ["production", "preview", "development"]
}

# Outputs for Vercel deployment
output "vercel_dashboard_url" {
  value       = "https://${vercel_project.synq_dashboard.name}.vercel.app"
  description = "The default Vercel URL of the Dashboard project"
}
