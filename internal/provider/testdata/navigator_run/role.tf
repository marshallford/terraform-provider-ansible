resource "ansible_navigator_run" "test" {
  ansible_navigator_binary = "%s"
  playbook                 = <<-EOT
  - hosts: localhost
    become: false
    tasks:
    - ansible.builtin.include_role:
        name: test_role
    - ansible.builtin.assert:
        that: test_role == "test"
  EOT
  inventory                = "# localhost"
  working_directory        = "%s"
}
