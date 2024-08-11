terraform {
  required_version = ">= 1.7.0"
  required_providers {
    libvirt = {
      source  = "dmacvicar/libvirt"
      version = "0.7.6"
    }
    tls = {
      source  = "hashicorp/tls"
      version = "4.0.5"
    }
    ansible = {
      source  = "marshallford/ansible"
      version = "0.10.1"
    }
  }
}
provider "libvirt" {
  uri = "qemu:///system"
}

provider "tls" {}

provider "ansible" {}

locals {
  machines = toset(["a", "b"])
}

resource "tls_private_key" "this" {
  algorithm = "ED25519"
}

resource "libvirt_pool" "this" {
  name = "example"
  type = "dir"
  path = "/var/lib/libvirt/images/example"
}

resource "libvirt_volume" "ubuntu" {
  name   = "ubuntu.qcow2"
  pool   = libvirt_pool.this.name
  source = "https://cloud-images.ubuntu.com/releases/22.04/release-20240416/ubuntu-22.04-server-cloudimg-amd64.img"
  format = "qcow2"
}

resource "libvirt_cloudinit_disk" "this" {
  for_each = local.machines
  name     = "example-cloudinit-${each.key}.iso"
  pool     = libvirt_pool.this.name
  user_data = templatefile("${path.module}/cloud_init.cfg.tpl", {
    hostname           = "example-${each.key}",
    ssh_authorized_key = tls_private_key.this.public_key_openssh
  })
  network_config = file("${path.module}/network.cfg.tpl")
}

resource "libvirt_volume" "this" {
  for_each         = local.machines
  name             = "example-${each.key}.qcow2"
  pool             = libvirt_pool.this.name
  base_volume_name = libvirt_volume.ubuntu.name
  size             = (1024 * 1024 * 1024) * 10
}

resource "libvirt_network" "this" {
  name      = "example"
  mode      = "nat"
  autostart = true
  addresses = ["172.16.1.0/24"]

  dhcp {
    enabled = true
  }

  dns {
    enabled = false
  }
}

resource "libvirt_domain" "this" {
  for_each  = local.machines
  name      = "example-${each.key}"
  vcpu      = 2
  memory    = 1024
  autostart = true
  cloudinit = libvirt_cloudinit_disk.this[each.key].id
  disk {
    volume_id = libvirt_volume.this[each.key].id
  }
  network_interface {
    network_id     = libvirt_network.this.id
    wait_for_lease = true
  }
}

locals {
  inventory = yamlencode({
    all = {
      children = {
        example_group = {
          hosts = { for domain in libvirt_domain.this : domain.name => {
            ansible_host = domain.network_interface[0].addresses[0]
            ansible_user = "ubuntu"
          } }
          vars = {
            hello_msg = "hello world (Libvirt)"
          }
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
  execution_environment = {
    container_options = [
      "--net=host", # required because libvirt nat network is on same host as EE
    ]
  }
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
  value = join("\n", jsondecode(ansible_navigator_run.this.artifact_queries.stdout.result[0]))
}
