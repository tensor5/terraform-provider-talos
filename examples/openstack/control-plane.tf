resource "openstack_compute_instance_v2" "talos_control_plane" {
  count = var.control_plane_number

  name        = "talos-control-plane-${count.index}"
  flavor_name = var.control_plane_flavor
  image_id    = openstack_images_image_v2.talos.id

  network {
    port = openstack_networking_port_v2.talos_control_plane[count.index].id
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
      machine = merge(
        local.control_plane_config.machine,
        {
          certSANs = [
            openstack_networking_floatingip_v2.talos_control_plane[count.index].address,
          ]
        }
      )
    }
  ))
}
