resource "talos_bootstrap" "openstack" {
  endpoint    = openstack_compute_instance_v2.talos_control_plane[0].access_ip_v4
  machine_ca  = base64decode(local.talos_config.contexts.openstack.ca)
  machine_crt = base64decode(local.talos_config.contexts.openstack.crt)
  machine_key = base64decode(local.talos_config.contexts.openstack.key)
}

output "kubeconfig" {
  value = talos_bootstrap.openstack
}

provider "kubernetes" {
  host                   = openstack_lb_loadbalancer_v2.talos_control_plane.vip_address
  client_certificate     = talos_bootstrap.openstack.client_certificate
  client_key             = talos_bootstrap.openstack.client_key
  cluster_ca_certificate = talos_bootstrap.openstack.cluster_ca_certificate
}

resource "time_sleep" "wait_60_seconds" {
  depends_on = [talos_bootstrap.openstack]

  create_duration = "60s"
}

data "kubernetes_all_namespaces" "allns" {
  depends_on = [time_sleep.wait_60_seconds]
}

output "all-ns" {
  value = data.kubernetes_all_namespaces.allns.namespaces
}
