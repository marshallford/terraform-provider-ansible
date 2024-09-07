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
          ansible_ssh_common_args = "-o StrictHostKeyChecking=accept-new -o UserKnownHostsFile={{ ansible_ssh_known_hosts_file }}"
        }
      }
    }
  })
  triggers = {
    known_hosts = var.trigger
  }
  execution_environment = {
    container_options = [
      "--net=host",
    ]
  }
}

variable "ssh_port" {
  type     = number
  nullable = false
}

variable "trigger" {
  type     = string
  nullable = false
}

output "known_host" {
  value = ansible_navigator_run.test.ansible_options.known_hosts[0]
}
