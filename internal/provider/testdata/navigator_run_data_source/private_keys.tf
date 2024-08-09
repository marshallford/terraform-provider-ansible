data "ansible_navigator_run" "test" {
  ansible_navigator_binary = var.ansible_navigator_binary
  playbook                 = <<-EOT
  - hosts: test
    gather_facts: false
    become: false
    tasks:
    - ansible.builtin.raw: test
      register: connect
    - ansible.builtin.assert:
        that: connect.stdout == 'hello world!'
  EOT
  inventory = yamlencode({
    all = {
      hosts = {
        test = {
          ansible_host            = "127.0.0.1"
          ansible_port            = var.ssh_port
          ansible_ssh_common_args = "-o StrictHostKeyChecking=yes -o UserKnownHostsFile={{ ansible_ssh_known_hosts_file }}"
        }
      }
    }
  })
  execution_environment = {
    enabled = var.ee_enabled
    container_options = [
      "--net=host",
    ]
  }
  ansible_options = {
    private_keys = [
      {
        name = "test"
        data = var.client_private_key_data
      },
    ]
    known_hosts = [
      "[127.0.0.1]:${var.ssh_port} ${var.server_public_key_data}",
    ]
  }
}

variable "ansible_navigator_binary" {
  type     = string
  nullable = false
}

variable "ee_enabled" {
  type     = bool
  nullable = false
}

variable "client_private_key_data" {
  type     = string
  nullable = false
}

variable "server_public_key_data" {
  type     = string
  nullable = false
}

variable "ssh_port" {
  type     = number
  nullable = false
}
