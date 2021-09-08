resource "digitalocean_droplet" "control_plane" {
  count = var.control_plane_number

  image              = digitalocean_custom_image.talos.id
  name               = "talos-control-plane-${count.index}"
  private_networking = true
  region             = var.region
  size               = var.control_plane_size
  ssh_keys           = [var.ssh_key_fingerprint]
  tags               = [var.control_plane_tag]

  user_data = yamlencode(merge(
    local.control_plane_config,
    {
      cluster = merge(
        local.control_plane_config.cluster,
        {
          controlPlane = {
            endpoint = "https://${digitalocean_loadbalancer.control_plane.ip}"
          }
        }
      )
    }
  ))
}
