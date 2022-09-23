resource "digitalocean_droplet" "worker" {
  count = var.worker_number

  image     = digitalocean_custom_image.talos.id
  name      = "talos-worker-${count.index}"
  region    = var.region
  size      = var.worker_size
  ssh_keys  = [var.ssh_key_fingerprint]
  user_data = talos_gen_config.config.worker_config
}
