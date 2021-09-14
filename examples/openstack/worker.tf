resource "openstack_compute_instance_v2" "talos_worker" {
  count = var.worker_number

  name        = "talos-worker-${count.index}"
  flavor_name = var.worker_flavor
  image_id    = openstack_images_image_v2.talos.id

  network {
    name = var.network_name
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
    }
  ))
}
