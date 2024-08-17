resource "ansible_navigator_run" "test" {
  ansible_navigator_binary = var.ansible_navigator_binary
  playbook                 = <<-EOT
  - hosts: localhost
    become: false
    tasks:
    - ansible.builtin.fail:
        msg: test
  EOT
  inventory                = "# localhost"
}
