# Availability Domain and Image Datasources
data "oci_identity_availability_domains" "ad" {
  compartment_id = var.oci_compartment_ocid
}

data "oci_core_images" "oracle_linux" {
  compartment_id           = var.oci_compartment_ocid
  operating_system         = "Oracle Linux"
  operating_system_version = "8"
  shape                    = var.oci_instance_shape
}

# Virtual Cloud Network (VCN)
resource "oci_core_vcn" "synq_vcn" {
  compartment_id = var.oci_compartment_ocid
  cidr_blocks    = ["10.0.0.0/16"]
  display_name   = "synq-vcn-${var.environment}"
  dns_label      = "synqvcn"
}

# Internet Gateway
resource "oci_core_internet_gateway" "synq_ig" {
  compartment_id = var.oci_compartment_ocid
  vcn_id         = oci_core_vcn.synq_vcn.id
  display_name   = "synq-internet-gateway"
}

# Route Table routing outbound traffic to Internet Gateway
resource "oci_core_route_table" "synq_rt" {
  compartment_id = var.oci_compartment_ocid
  vcn_id         = oci_core_vcn.synq_vcn.id
  display_name   = "synq-route-table"

  route_rules {
    destination       = "0.0.0.0/0"
    destination_type  = "CIDR_BLOCK"
    network_entity_id = oci_core_internet_gateway.synq_ig.id
  }
}

# Security List governing VM access
resource "oci_core_security_list" "synq_sl" {
  compartment_id = var.oci_compartment_ocid
  vcn_id         = oci_core_vcn.synq_vcn.id
  display_name   = "synq-security-list"

  # Outbound rule (allow all traffic out)
  egress_security_rules {
    destination = "0.0.0.0/0"
    protocol    = "all"
  }

  # Inbound Rule: SSH (Port 22)
  ingress_security_rules {
    protocol    = "6" # TCP
    source      = "0.0.0.0/0"
    description = "Allow inbound SSH access"
    tcp_options {
      min = 22
      max = 22
    }
  }

  # Inbound Rule: PostgreSQL (Port 5432)
  ingress_security_rules {
    protocol    = "6" # TCP
    source      = "0.0.0.0/0"
    description = "Allow inbound Postgres connection (secure with SSL/passwords)"
    tcp_options {
      min = 5432
      max = 5432
    }
  }

  # Inbound Rule: Valkey/Redis (Port 6379)
  ingress_security_rules {
    protocol    = "6" # TCP
    source      = "0.0.0.0/0"
    description = "Allow inbound Valkey/Redis connection"
    tcp_options {
      min = 6379
      max = 6379
    }
  }

  # Inbound Rule: Temporal Frontend (Port 7233)
  ingress_security_rules {
    protocol    = "6" # TCP
    source      = "0.0.0.0/0"
    description = "Allow inbound Temporal gRPC client connection"
    tcp_options {
      min = 7233
      max = 7233
    }
  }

  # Inbound Rule: Temporal Web UI (Port 8080)
  ingress_security_rules {
    protocol    = "6" # TCP
    source      = "0.0.0.0/0"
    description = "Allow inbound Temporal Dashboard access"
    tcp_options {
      min = 8080
      max = 8080
    }
  }

  # Inbound Rule: ICMP Ping
  ingress_security_rules {
    protocol    = "1" # ICMP
    source      = "0.0.0.0/0"
    description = "Allow ICMP ping requests for networking checks"
  }
}

# Public Subnet in VCN
resource "oci_core_subnet" "synq_subnet" {
  compartment_id    = var.oci_compartment_ocid
  vcn_id            = oci_core_vcn.synq_vcn.id
  cidr_block        = "10.0.1.0/24"
  display_name      = "synq-subnet"
  dns_label         = "synqsub"
  route_table_id    = oci_core_route_table.synq_rt.id
  security_list_ids = [oci_core_security_list.synq_sl.id]
}

# Compute Instance hosting stateful databases (ARM64 VM)
resource "oci_core_instance" "synq_db_server" {
  # Select first Availability Domain in the region
  availability_domain = data.oci_identity_availability_domains.ad.availability_domains[0].name
  compartment_id      = var.oci_compartment_ocid
  shape               = var.oci_instance_shape
  display_name        = "synq-db-server-${var.environment}"

  shape_config {
    ocpus         = var.oci_instance_ocpus
    memory_in_gbs = var.oci_instance_memory_gbs
  }

  create_vnic_details {
    subnet_id        = oci_core_subnet.synq_subnet.id
    display_name     = "primaryvnic"
    assign_public_ip = true
    hostname_label   = "synq-db-server"
  }

  source_details {
    # Dynamically select standard ARM64 Oracle Linux image
    source_id   = data.oci_core_images.oracle_linux.images[0].id
    source_type = "image"
  }

  metadata = {
    ssh_authorized_keys = var.oci_ssh_public_key
  }

  # Ensure resources are created sequentially
  lifecycle {
    ignore_changes = [
      # Ignore change in source_id if updated image becomes available later
      source_details[0].source_id
    ]
  }
}

# Outputs for OCI infrastructure
output "oci_server_public_ip" {
  value       = oci_core_instance.synq_db_server.public_ip
  description = "The public IP address of the Oracle DB instance"
}
