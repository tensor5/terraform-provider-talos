# Terraform Provider for Talos

This is a [Terraform](https://www.terraform.io) provider for [Talos](https://www.talos.dev), a minimal and modern OS for running Kubernetes clusters. It helps to bootstrap a Talos based Kubernetes cluster, and returns the `kubeconfig` data that can be used with the [Kubernetes](https://registry.terraform.io/providers/hashicorp/kubernetes/latest) and [Helm](https://registry.terraform.io/providers/hashicorp/helm/latest) providers.

The `examples` folder contains working Terraform configurations to deploy Talos clusters on [DigitalOcean](https://www.digitalocean.com) and [OpenStack](https://www.openstack.org), and can be used as a starting point to write configurations for other cloud providers.
