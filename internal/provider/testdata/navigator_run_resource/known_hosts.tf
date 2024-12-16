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
  execution_environment = {
    container_options = [
      "--net=host",
    ]
  }
  ansible_options = {
    host_key_checking = true
  }
}

variable "ssh_port" {
  type     = number
  nullable = false
}
