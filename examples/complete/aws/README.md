# AWS Example

This Terraform configuration creates an AL2023 AWS EC2 instance using the [`hashicorp/aws`](https://registry.terraform.io/providers/hashicorp/aws/latest/docs) provider, constructs an Ansible inventory containing the virtual machine, and runs a playbook against said inventory.

## Prerequisites

1. `docker`
2. `python3` with `uv`
3. `terraform`
4. AWS account and matching credentials (`AmazonEC2FullAccess` and `IAMFullAccess` managed policies or equivalent access required)

## Steps

1. Run `make` to install `ansible-builder` and `ansible-navigator` packages into a Python virtual environment
2. Run `make build` to build the Ansible EEI (container image)
3. Setup [AWS authentication](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#authentication-and-configuration)
4. Run `terraform init` and `terraform apply`
