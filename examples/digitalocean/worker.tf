resource "digitalocean_droplet" "worker" {
  count = var.worker_number

  image              = digitalocean_custom_image.talos.id
  name               = "talos-worker-${count.index}"
  private_networking = true
  region             = var.region
  size               = var.worker_size
  ssh_keys           = [var.ssh_key_fingerprint]

  user_data = yamlencode(merge(
    local.worker_config,
    {
      cluster = merge(
        local.worker_config.cluster,
        {
          controlPlane = {
            endpoint = "https://${digitalocean_loadbalancer.control_plane.ip}"
          }
        }
      )
    }
  ))
}
