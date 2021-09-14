terraform {
  required_providers {
    openstack = {
      source  = "terraform-provider-openstack/openstack"
      version = "~> 1.43"
    }
    talos = {
      source  = "tensor5/talos"
      version = "~> 0.1"
    }
  }
}

provider "openstack" {
  auth_url                      = var.auth_url
  application_credential_id     = var.application_credential_id
  application_credential_secret = var.application_credential_secret
  region                        = var.region
  use_octavia                   = true
}

provider "talos" {}
