variable "environment" {
  type        = string
  description = "Deployment environment (e.g., beta, production)"
  default     = "beta"
}

# GCP Variables
variable "gcp_project_id" {
  type        = string
  description = "The GCP Project ID to deploy resources to"
}

variable "gcp_region" {
  type        = string
  description = "The GCP region for Cloud Run and other resources"
  default     = "us-central1"
}

# OCI Variables
variable "oci_tenancy_ocid" {
  type        = string
  description = "OCI Tenancy OCID"
}

variable "oci_user_ocid" {
  type        = string
  description = "OCI User OCID"
}

variable "oci_fingerprint" {
  type        = string
  description = "Fingerprint of the OCI API private key"
}

variable "oci_private_key_path" {
  type        = string
  description = "Path to the OCI API private key on local system"
}

variable "oci_region" {
  type        = string
  description = "OCI region where resources will be provisioned"
}

variable "oci_compartment_ocid" {
  type        = string
  description = "Compartment OCID where the instance and network will reside"
}

variable "oci_ssh_public_key" {
  type        = string
  description = "SSH public key content to install on the compute instance"
}

variable "oci_instance_shape" {
  type        = string
  description = "Compute instance shape for the stateful server"
  default     = "VM.Standard.A1.Flex"
}

variable "oci_instance_ocpus" {
  type        = number
  description = "Number of OCPUs for the VM (Free tier allows up to 4)"
  default     = 4
}

variable "oci_instance_memory_gbs" {
  type        = number
  description = "Amount of RAM in GBs (Free tier allows up to 24)"
  default     = 24
}

# Vercel Variables
variable "vercel_api_token" {
  type        = string
  description = "Vercel API Token for provisioning frontend resources"
}

variable "vercel_team_id" {
  type        = string
  description = "Vercel Team ID (required if deploying under a team account)"
  default     = null
}

variable "vercel_github_repo" {
  type        = string
  description = "Vercel Git repository name (e.g., Shubham04-oss/beta-prod)"
  default     = "Shubham04-oss/beta-prod"
}

# Database Credentials
variable "db_password" {
  type        = string
  description = "Password for the Postgres database user"
  sensitive   = true
}

variable "valkey_password" {
  type        = string
  description = "Password for Valkey (Redis) cache"
  sensitive   = true
  default     = ""
}
