resource "ansible_navigator_run" "test" {
  ansible_navigator_binary = var.ansible_navigator_binary
  playbook                 = <<-EOT
  - hosts: localhost
    gather_facts: false
    become: false
    tasks:
    - ansible.builtin.include_role:
        name: test_role
    - ansible.builtin.assert:
        that: test_role == "test"
  EOT
  inventory                = "# localhost"
  working_directory        = var.working_directory
}

variable "working_directory" {
  type     = string
  nullable = false
}
