resource "openstack_images_image_v2" "talos" {
  container_format = "bare"
  disk_format      = "raw"
  image_source_url = var.talos_image
  name             = "Talos"
}
