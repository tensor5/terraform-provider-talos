resource "openstack_networking_secgroup_v2" "talos_allow_internal" {
  name = "talos-allow-internal"
}

resource "openstack_networking_secgroup_rule_v2" "talos_allow_internal_ipv4" {
  direction         = "ingress"
  ethertype         = "IPv4"
  remote_group_id   = openstack_networking_secgroup_v2.talos_allow_internal.id
  security_group_id = openstack_networking_secgroup_v2.talos_allow_internal.id
}

resource "openstack_networking_secgroup_rule_v2" "talos_allow_internal_ipv6" {
  direction         = "ingress"
  ethertype         = "IPv6"
  remote_group_id   = openstack_networking_secgroup_v2.talos_allow_internal.id
  security_group_id = openstack_networking_secgroup_v2.talos_allow_internal.id
}

resource "openstack_networking_secgroup_v2" "talos_external" {
  for_each = local.node_types

  name = "talos-external-${each.value}"
}

resource "openstack_networking_secgroup_rule_v2" "kubernetes_api_ipv4" {
  direction         = "ingress"
  ethertype         = "IPv4"
  protocol          = "tcp"
  port_range_min    = 6443
  port_range_max    = 6443
  security_group_id = openstack_networking_secgroup_v2.talos_external["control_plane"].id
}

resource "openstack_networking_secgroup_rule_v2" "kubernetes_api_ipv6" {
  direction         = "ingress"
  ethertype         = "IPv6"
  protocol          = "tcp"
  port_range_min    = 6443
  port_range_max    = 6443
  security_group_id = openstack_networking_secgroup_v2.talos_external["control_plane"].id
}

resource "openstack_networking_secgroup_rule_v2" "talos_api_ipv4" {
  for_each = local.node_types

  direction         = "ingress"
  ethertype         = "IPv4"
  protocol          = "tcp"
  port_range_min    = 50000
  port_range_max    = 50000
  security_group_id = openstack_networking_secgroup_v2.talos_external[each.key].id
}

resource "openstack_networking_secgroup_rule_v2" "talos_api_ipv6" {
  for_each = local.node_types

  direction         = "ingress"
  ethertype         = "IPv6"
  protocol          = "tcp"
  port_range_min    = 50000
  port_range_max    = 50000
  security_group_id = openstack_networking_secgroup_v2.talos_external[each.key].id
}
