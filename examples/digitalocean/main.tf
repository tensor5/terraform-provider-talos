terraform {
  required_providers {
    digitalocean = {
      source  = "digitalocean/digitalocean"
      version = "~> 2.11"
    }
    talos = {
      source  = "tensor5/talos"
      version = "~> 0.1"
    }
  }
}

provider "digitalocean" {
  token = var.do_token
}

provider "talos" {}
