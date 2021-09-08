resource "talos_bootstrap" "digitalocean" {
  endpoint    = digitalocean_droplet.control_plane[0].ipv4_address
  machine_ca  = base64decode(local.talos_config.contexts.digitalocean.ca)
  machine_crt = base64decode(local.talos_config.contexts.digitalocean.crt)
  machine_key = base64decode(local.talos_config.contexts.digitalocean.key)
}

output "kubeconfig" {
  value = talos_bootstrap.digitalocean.raw
}

provider "kubernetes" {
  host                   = digitalocean_loadbalancer.control_plane.ip
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
