# Libvirt Example

This example creates a KVM domain using the [`dmacvicar/libvirt`](https://registry.terraform.io/providers/dmacvicar/libvirt/latest/docs) provider and connects to the virtual machine with Ansible.

## Prerequisites

1. `docker`
2. `python3` with `venv`
3. `terraform`
4. KVM host accessible [via libvirt](https://registry.terraform.io/providers/dmacvicar/libvirt/latest/docs#the-connection-uri)

## Steps

1. Run `make` to install `ansible-builder` and `ansible-navigator` packages into a Python virtual environment.
2. Run `.venv/bin/ansible-builder build --container-runtime=docker -t ansible-execution-env-libvirt-example:v1 --no-cache` to build Ansible EE container image.
3. Run `terraform init` and `terraform apply`
