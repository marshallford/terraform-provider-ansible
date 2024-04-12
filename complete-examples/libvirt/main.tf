terraform {
  required_version = ">= 1.6.0"
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
      version = "0.7.0"
    }
  }
}
provider "libvirt" {
  uri = "qemu:///system"
}

provider "tls" {}

provider "ansible" {}

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
  source = "https://cloud-images.ubuntu.com/releases/22.04/release-20240319/ubuntu-22.04-server-cloudimg-amd64.img"
  format = "qcow2"
}

resource "libvirt_cloudinit_disk" "this" {
  name = "example-cloudinit.iso"
  pool = libvirt_pool.this.name
  user_data = templatefile("${path.module}/cloud_init.cfg.tpl", {
    hostname           = "example",
    ssh_authorized_key = tls_private_key.this.public_key_openssh
  })
  network_config = file("${path.module}/network.cfg.tpl")
}

resource "libvirt_volume" "this" {
  name             = "example.qcow2"
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
  name      = "example"
  vcpu      = 2
  memory    = 1024
  autostart = true
  cloudinit = libvirt_cloudinit_disk.this.id
  disk {
    volume_id = libvirt_volume.this.id
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
          hosts = {
            example = {
              ansible_host = libvirt_domain.this.network_interface[0].addresses[0]
              ansible_user = "ubuntu"
            }
          }
          vars = {
            hello_msg = "hello world"
          }
        }
      }
    }
  })
}

resource "ansible_navigator_run" "this" {
  working_directory        = abspath("${path.root}/working-directory")
  playbook                 = <<-EOT
  - hosts: all
    gather_facts: false
    tasks:
    - name: wait for hosts
      ansible.builtin.wait_for_connection:
        timeout: 600
    - name: gather facts
      ansible.builtin.setup:
    - name: hello
      ansible.builtin.debug:
        msg: "{{ hello_msg }}! Distribution: {{ ansible_facts.distribution }}, System Vendor: {{ ansible_facts.system_vendor }}"
  EOT
  inventory                = local.inventory
  ansible_navigator_binary = abspath("${path.root}/.venv/bin/ansible-navigator")
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
      jsonpath = "$.stdout"
    }
  }
}

output "playbook_stdout" {
  value = join("\n", jsondecode(ansible_navigator_run.this.artifact_queries.stdout.result))
}
