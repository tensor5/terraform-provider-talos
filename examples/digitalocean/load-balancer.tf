resource "digitalocean_loadbalancer" "control_plane" {
  name   = "control-plane"
  region = var.region

  forwarding_rule {
    entry_protocol  = "tcp"
    entry_port      = 443
    target_protocol = "tcp"
    target_port     = 6443
  }

  healthcheck {
    protocol = "tcp"
    port     = 6443
  }

  droplet_tag = var.control_plane_tag
}
