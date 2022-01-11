variable "control_plane_number" {
  description = "Number of control plane nodes"
  default     = 3
}

variable "control_plane_flavor" {
  description = "Flavor of a control plane node"
  type        = string
}

variable "worker_number" {
  description = "Number of worker nodes"
  default     = 1
}

variable "worker_flavor" {
  description = "Flavor of a worker node"
  type        = string
}

variable "external_network" {
  description = "ID of external network (for loadbalancer)"
  type        = string
}

variable "floating_ip_network_id" {
  description = "ID of network to allocate floating IP from"
  type        = string
}

variable "floating_ip_network_name" {
  description = "Name of network to allocate floating IP from"
  type        = string
}

variable "auth_url" {
  description = "OpenStack authentication URL"
  type        = string
}

variable "application_credential_id" {
  type = string
}

variable "application_credential_secret" {
  sensitive = true
  type      = string
}

variable "region" {
  description = "Region"
  type        = string
}

variable "talos_image" {
  description = "URL of Talos raw disk image"
  type        = string
}
