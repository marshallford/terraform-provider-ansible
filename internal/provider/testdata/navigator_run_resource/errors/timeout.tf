resource "ansible_navigator_run" "test" {
  ansible_navigator_binary = var.ansible_navigator_binary
  playbook                 = <<-EOT
  - hosts: localhost
    gather_facts: false
    become: false
    tasks:
    - ansible.builtin.pause:
        seconds: 5
  EOT
  inventory                = "# localhost"
  timeouts = {
    create = "3s"
  }
}
