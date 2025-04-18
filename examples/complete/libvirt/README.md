# Libvirt Example

This Terraform configuration creates a KVM domain using the [`dmacvicar/libvirt`](https://registry.terraform.io/providers/dmacvicar/libvirt/latest/docs) provider, constructs an Ansible inventory containing the virtual machine, and runs a playbook against said inventory.

## Prerequisites

1. `docker`
2. `python3` with `venv`
3. `terraform`
4. KVM host accessible [via libvirt](https://registry.terraform.io/providers/dmacvicar/libvirt/latest/docs#the-connection-uri)

## Steps

1. Run `make` to install the `ansible-navigator` package into a Python virtual environment.
2. Run `terraform init` and `terraform apply`
