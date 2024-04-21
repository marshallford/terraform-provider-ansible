resource "ansible_navigator_run" "test" {
  ansible_navigator_binary = "%s"
  playbook                 = <<-EOT
  - hosts: localhost
    become: false
  EOT
  inventory                = "# localhost"
}
