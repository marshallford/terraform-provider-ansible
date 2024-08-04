resource "ansible_navigator_run" "test" {
  playbook  = <<-EOT
  - hosts: localhost
    become: false
  EOT
  inventory = "# localhost"
}
