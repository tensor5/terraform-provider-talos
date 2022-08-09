resource "talos_bootstrap" "openstack" {
  depends_on = [
    openstack_compute_instance_v2.talos_control_plane[0],
  ]

  endpoint    = openstack_networking_floatingip_v2.talos_control_plane[0].address
  machine_ca  = base64decode(local.talos_config.contexts.openstack.ca)
  machine_crt = base64decode(local.talos_config.contexts.openstack.crt)
  machine_key = base64decode(local.talos_config.contexts.openstack.key)
}

resource "local_file" "kube_config" {
  content  = talos_bootstrap.openstack.raw
  filename = "${path.module}/kubeconfig"
}

provider "kubernetes" {
  host                   = "https://${openstack_lb_loadbalancer_v2.talos_control_plane.vip_address}"
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
