resource "ansible_navigator_run" "test" {
  ansible_navigator_binary = var.ansible_navigator_binary
  playbook                 = <<-EOT
  - hosts: localhost
    gather_facts: false
    become: false
  EOT
  inventory                = "# localhost"
  ansible_options = {
    extra_vars = <<-EOT
    not: valid
      yaml: {{{
    EOT
  }
}
