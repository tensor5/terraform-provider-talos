data "talos_kubeconfig" "example" {
  endpoint    = "<ip address>"
  machine_ca  = "cert autority"
  machine_crt = "cert"
  machine_key = "key"
}
