# Proxmox Example

This Terraform configuration creates a Fedora CoreOS Proxmox virtual machine using the [`bpg/proxmox`](https://registry.terraform.io/providers/bpg/proxmox/latest/docs) provider, constructs an Ansible inventory containing the virtual machine, and runs a playbook against that inventory.

## Prerequisites

1. `docker`
2. `python3` with `uv`
3. `terraform`
4. Proxmox host and the necessary [credentials](https://registry.terraform.io/providers/bpg/proxmox/latest/docs#authentication)

## Steps

1. Run `make` to install the `ansible-navigator` package into a Python virtual environment
2. Copy `terraform.auto.tfvars.example` to `terraform.auto.tfvars` and update the values
3. Run `terraform init` and `terraform apply`

## References

1. https://blog.cloudbending.dev/posts/fedora-coreos-on-proxmox/
2. https://frank.villaro-dixon.eu/posts/coreos-proxmox/
