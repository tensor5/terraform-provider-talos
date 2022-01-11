resource "openstack_compute_instance_v2" "talos_worker" {
  count = var.worker_number

  name        = "talos-worker-${count.index}"
  flavor_name = var.worker_flavor
  image_id    = openstack_images_image_v2.talos.id

  network {
    port = openstack_networking_port_v2.talos_worker[count.index].id
  }

  user_data = yamlencode(merge(
    local.worker_config,
    {
      cluster = merge(
        local.worker_config.cluster,
        {
          controlPlane = {
            endpoint = "https://${openstack_lb_loadbalancer_v2.talos_control_plane.vip_address}"
          }
        }
      )
      machine = merge(
        local.worker_config.machine,
        {
          certSANs = [
            openstack_networking_floatingip_v2.talos_worker[count.index].address,
          ]
        }
      )
    }
  ))
}
