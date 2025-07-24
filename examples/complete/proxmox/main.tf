terraform {
  required_version = ">= 1.12.0"
  required_providers {
    proxmox = {
      source  = "bpg/proxmox"
      version = "0.80.0"
    }
    ct = {
      source  = "poseidon/ct"
      version = "0.13.0"
    }
    tls = {
      source  = "hashicorp/tls"
      version = "4.1.0"
    }
    ansible = {
      source  = "marshallford/ansible"
      version = "0.31.0"
    }
  }
}
provider "proxmox" {
  endpoint = var.proxmox_endpoint
  username = var.proxmox_username
  password = var.proxmox_password
  insecure = true
}

provider "ct" {}

provider "tls" {}

provider "ansible" {}

locals {
  machines = toset(["a", "b"])
}

resource "tls_private_key" "this" {
  algorithm = "ED25519"
}

data "ct_config" "this" {
  for_each = local.machines
  content = templatefile("${path.root}/config.tftpl.yaml", {
    hostname           = "example-${each.key}"
    ssh_authorized_key = tls_private_key.this.public_key_openssh
  })
  strict       = true
  pretty_print = true
}

resource "proxmox_virtual_environment_file" "ignition" {
  for_each     = data.ct_config.this
  content_type = "snippets"
  datastore_id = var.proxmox_datastore_directory
  node_name    = var.proxmox_node

  source_raw {
    data      = each.value.rendered
    file_name = "example-${each.key}.ign"
  }
}

resource "proxmox_virtual_environment_download_file" "fcos" {
  content_type            = "iso"
  datastore_id            = var.proxmox_datastore_directory
  file_name               = "fedora-coreos-42.20250705.3.0-proxmoxve.x86_64.img"
  node_name               = var.proxmox_node
  url                     = "https://builds.coreos.fedoraproject.org/prod/streams/stable/builds/42.20250705.3.0/x86_64/fedora-coreos-42.20250705.3.0-proxmoxve.x86_64.qcow2.xz"
  checksum                = "dab2cfafe397aa96e2885d11ab89a1463bc0dd49a04c7f06bfef7246d13f0437"
  checksum_algorithm      = "sha256"
  decompression_algorithm = "zst"
}

resource "proxmox_virtual_environment_vm" "this" {
  for_each  = proxmox_virtual_environment_file.ignition
  name      = "example-${each.key}"
  node_name = var.proxmox_node
  machine   = "q35"

  agent {
    enabled = true
  }

  cpu {
    cores = 2
  }

  memory {
    dedicated = 1024
  }

  disk {
    datastore_id = var.proxmox_datastore_disk
    file_id      = proxmox_virtual_environment_download_file.fcos.id
    interface    = "virtio0"
    iothread     = true
    discard      = "on"
    size         = 10
  }

  network_device {
    bridge = "vmbr0"
  }

  initialization {
    datastore_id      = var.proxmox_datastore_disk
    user_data_file_id = each.value.id
  }
}

locals {
  inventory = yamlencode({
    all = {
      vars = {
        ansible_python_interpreter = "/usr/bin/python3"
        ansible_user               = "core"
        hello_msg                  = "hello world (Proxmox)"
      }
      children = {
        example_group = {
          hosts = { for vm in proxmox_virtual_environment_vm.this : vm.name => {
            ansible_host = vm.ipv4_addresses[1][0] # first interface is lo
          } }
        }
      }
    }
  })
}

resource "ansible_navigator_run" "this" {
  playbook                 = file("${path.root}/playbook.yaml")
  inventory                = local.inventory
  working_directory        = "${path.root}/working-directory"
  ansible_navigator_binary = "${path.root}/.venv/bin/ansible-navigator"
  ansible_options = {
    private_keys = [
      { name = "terraform", data = tls_private_key.this.private_key_openssh },
    ]
  }
  artifact_queries = {
    "stdout" = {
      jq_filter = ".stdout"
    }
  }
}

output "playbook_stdout" {
  value = join("\n", jsondecode(ansible_navigator_run.this.artifact_queries.stdout.results[0]))
}
