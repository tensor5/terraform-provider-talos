resource "openstack_networking_floatingip_v2" "talos_control_plane" {
  count      = var.control_plane_number
  depends_on = [openstack_networking_router_interface_v2.talos]

  pool    = var.floating_ip_network_name
  port_id = openstack_networking_port_v2.talos_control_plane[count.index].id
}

resource "openstack_networking_floatingip_v2" "talos_worker" {
  count      = var.worker_number
  depends_on = [openstack_networking_router_interface_v2.talos]

  pool    = var.floating_ip_network_name
  port_id = openstack_networking_port_v2.talos_worker[count.index].id
}
