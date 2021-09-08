variable "cluster_name" {
  description = "Cluster name"
  default     = "talos-digitalocean"
}

variable "control_plane_number" {
  description = "Number of control plane nodes"
  default     = 3
}

variable "control_plane_size" {
  description = "Size of a control plane node"
  default     = "s-2vcpu-4gb"
}

variable "control_plane_tag" {
  description = "Tag assigned to control plane nodes"
  default     = "control-plane"
}

variable "worker_number" {
  description = "Number of worker nodes"
  default     = 1
}

variable "worker_size" {
  description = "Size of a worker node"
  default     = "s-2vcpu-4gb"
}

variable "do_token" {
  description = "DigitalOcean authentication token"
  sensitive   = true
  type        = string
}

variable "region" {
  description = "Region"
  type        = string
}

variable "ssh_key_fingerprint" {
  description = "SSH key fingerprint"
  type        = string
}

variable "talos_image" {
  description = "URL of Talos raw disk image"
  type        = string
}
