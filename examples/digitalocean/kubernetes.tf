resource "talos_gen_config" "config" {
  cluster_name     = var.cluster_name
  cluster_endpoint = "https://${digitalocean_loadbalancer.control_plane.ip}"
}

resource "local_file" "talos_config" {
  content  = talos_gen_config.config.talos_config
  filename = "${path.module}/talosconfig"
}

locals {
  talos_config = yamldecode(talos_gen_config.config.talos_config).contexts[var.cluster_name]
}

resource "talos_bootstrap" "digitalocean" {
  endpoint    = digitalocean_droplet.control_plane[0].ipv4_address
  machine_ca  = base64decode(local.talos_config.ca)
  machine_crt = base64decode(local.talos_config.crt)
  machine_key = base64decode(local.talos_config.key)
}

resource "local_file" "kube_config" {
  content  = talos_bootstrap.digitalocean.raw
  filename = "${path.module}/kubeconfig"
}

provider "kubernetes" {
  host                   = "https://${digitalocean_loadbalancer.control_plane.ip}"
  client_certificate     = talos_bootstrap.digitalocean.client_certificate
  client_key             = talos_bootstrap.digitalocean.client_key
  cluster_ca_certificate = talos_bootstrap.digitalocean.cluster_ca_certificate
}

resource "time_sleep" "wait_60_seconds" {
  depends_on = [talos_bootstrap.digitalocean]

  create_duration = "60s"
}

data "kubernetes_all_namespaces" "allns" {
  depends_on = [time_sleep.wait_60_seconds]
}

output "all-ns" {
  value = data.kubernetes_all_namespaces.allns.namespaces
}
