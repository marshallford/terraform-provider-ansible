resource "ansible_navigator_run" "test" {
  working_directory        = "%s"
  playbook                 = <<-EOT
  - hosts: localhost
    become: false
  EOT
  inventory                = "# localhost"
}
