resource "openstack_networking_network_v2" "talos_internal" {
  name = "talos-internal"
}

resource "openstack_networking_subnet_v2" "talos_internal_ipv4" {
  name       = "talos-internal-ipv4"
  cidr       = "192.168.1.0/24"
  ip_version = 4
  network_id = openstack_networking_network_v2.talos_internal.id
  dns_nameservers = [
    "1.1.1.1",
    "1.0.0.1",
  ]
}

resource "openstack_networking_router_v2" "talos" {
  external_network_id = var.floating_ip_network_id
}

resource "openstack_networking_router_interface_v2" "talos" {
  router_id = openstack_networking_router_v2.talos.id
  subnet_id = openstack_networking_subnet_v2.talos_internal_ipv4.id
}
