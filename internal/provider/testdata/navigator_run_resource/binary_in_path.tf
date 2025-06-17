resource "ansible_navigator_run" "test" {
  playbook  = <<-EOT
  - hosts: localhost
    gather_facts: false
    become: false
  EOT
  inventory = "# localhost"
}
