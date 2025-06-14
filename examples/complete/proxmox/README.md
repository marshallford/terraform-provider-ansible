# Proxmox Example

This Terraform configuration creates a Fedora CoreOS Proxmox virtual machine [`bpg/proxmox`](https://registry.terraform.io/providers/bpg/proxmox/latest/docs) provider, constructs an Ansible inventory containing the virtual machine, and runs a playbook against said inventory.

## Prerequisites

1. `docker`
2. `python3` with `venv`
3. `terraform`
4. Proxmox host and the necessary [credentials](https://registry.terraform.io/providers/bpg/proxmox/latest/docs#authentication)

## Steps

1. Run `make` to install the `ansible-navigator` package into a Python virtual environment
2. Download, decompress, and upload the latest stable [Fedora CoreOS IBM Cloud image](https://fedoraproject.org/coreos/download?stream=stable) to the Proxmox host. In order to upload via the Proxmox UI, `.img` may need to be appended to the filename. In addition, depending on the desired virtual machine disk size, the image may need to be resized via `qemu-img resize --shrink fedora-coreos-... 10G` as the IBM Cloud images are preconfigured with a large virtual disk size (`100 GiB`).
3. Copy `terraform.auto.tfvars.example` to `terraform.auto.tfvars` and update the values
4. Run `terraform init` and `terraform apply`

## References

1. https://blog.cloudbending.dev/posts/fedora-coreos-on-proxmox/
2. https://frank.villaro-dixon.eu/posts/coreos-proxmox/
