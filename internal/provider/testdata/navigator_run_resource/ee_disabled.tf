resource "ansible_navigator_run" "test" {
  ansible_navigator_binary = var.ansible_navigator_binary
  playbook                 = <<-EOT
  - hosts: localhost
    gather_facts: false
    become: false
    tasks:
    - ansible.builtin.command:
        cmd: docker info
  EOT
  inventory                = "# localhost"
  execution_environment = {
    container_engine = "docker"
    enabled          = false
  }
}
