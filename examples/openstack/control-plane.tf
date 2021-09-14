resource "openstack_compute_instance_v2" "talos_control_plane" {
  count = var.control_plane_number

  name        = "talos-control-plane-${count.index}"
  flavor_name = var.control_plane_flavor
  image_id    = openstack_images_image_v2.talos.id

  network {
    name = var.network_name
  }

  user_data = yamlencode(merge(
    local.control_plane_config,
    {
      cluster = merge(
        local.control_plane_config.cluster,
        {
          controlPlane = {
            endpoint = "https://${openstack_lb_loadbalancer_v2.talos_control_plane.vip_address}"
          }
        }
      )
    }
  ))
}
