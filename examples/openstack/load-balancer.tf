resource "openstack_lb_loadbalancer_v2" "talos_control_plane" {
  name           = "talos-control-plane"
  vip_network_id = var.external_network
}

resource "openstack_lb_listener_v2" "talos_control_plane" {
  name            = "talos-control-plane"
  loadbalancer_id = openstack_lb_loadbalancer_v2.talos_control_plane.id
  protocol        = "TCP"
  protocol_port   = 443
}

resource "openstack_lb_pool_v2" "talos_control_plane" {
  name        = "talos-control-plane"
  lb_method   = "ROUND_ROBIN"
  listener_id = openstack_lb_listener_v2.talos_control_plane.id
  protocol    = "TCP"
}

resource "openstack_lb_monitor_v2" "talos_control_plane" {
  pool_id     = openstack_lb_pool_v2.talos_control_plane.id
  delay       = 5
  max_retries = 4
  timeout     = 10
  type        = "TCP"
}

resource "openstack_lb_member_v2" "talos_control_plane" {
  count = var.control_plane_number

  name          = "talos-control-plane-${count.index}"
  address       = openstack_networking_port_v2.talos_control_plane[count.index].all_fixed_ips[0]
  pool_id       = openstack_lb_pool_v2.talos_control_plane.id
  protocol_port = 6443
  subnet_id     = openstack_networking_subnet_v2.talos_internal_ipv4.id
}
