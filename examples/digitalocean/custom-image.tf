resource "digitalocean_custom_image" "talos" {
  name    = "Talos"
  url     = var.talos_image
  regions = [var.region]
}
