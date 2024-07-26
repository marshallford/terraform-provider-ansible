resource "ansible_navigator_run" "test" {
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
          ansible_ssh_common_args = "-o UserKnownHostsFile=/dev/null"
        }
      }
    }
  })
  execution_environment = {
    container_options = [
      "--net=host",
    ]
  }
  ansible_options = {
    private_keys = [
      {
        name = "test"
        data = var.private_key_data
      }
    ]
  }
}

variable "ansible_navigator_binary" {
  type     = string
  nullable = false
}

variable "private_key_data" {
  type     = string
  nullable = false
}

variable "ssh_port" {
  type     = number
  nullable = false
}
