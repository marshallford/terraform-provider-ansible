resource "ansible_navigator_run" "test" {
  ansible_navigator_binary = "%s"
  working_directory        = "%s"
  playbook                 = <<-EOT
  - hosts: localhost
    become: false
    tasks:
    - ansible.builtin.pause:
        seconds: 10
  EOT
  inventory                = "# localhost"
  timeouts = {
    create = "5s"
  }
}
