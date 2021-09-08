resource "talos_bootstrap" "example" {
  endpoint    = "<ip address>"
  machine_ca  = "cert autority"
  machine_crt = "cert"
  machine_key = "key"
}
