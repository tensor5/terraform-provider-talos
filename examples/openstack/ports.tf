resource "openstack_networking_port_v2" "talos_control_plane" {
  count = var.control_plane_number

  name       = "talos-control-plane-${count.index}"
  network_id = openstack_networking_network_v2.talos_internal.id
  security_group_ids = [
    openstack_networking_secgroup_v2.talos_external["control_plane"].id,
    openstack_networking_secgroup_v2.talos_allow_internal.id,
  ]

  fixed_ip {
    subnet_id = openstack_networking_subnet_v2.talos_internal_ipv4.id
  }
}

resource "openstack_networking_port_v2" "talos_worker" {
  count = var.worker_number

  name       = "talos-worker-${count.index}"
  network_id = openstack_networking_network_v2.talos_internal.id
  security_group_ids = [
    openstack_networking_secgroup_v2.talos_external["worker"].id,
    openstack_networking_secgroup_v2.talos_allow_internal.id,
  ]

  fixed_ip {
    subnet_id = openstack_networking_subnet_v2.talos_internal_ipv4.id
  }
}
