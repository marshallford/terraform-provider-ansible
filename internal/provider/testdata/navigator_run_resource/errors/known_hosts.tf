resource "ansible_navigator_run" "test" {
  ansible_navigator_binary = var.ansible_navigator_binary
  playbook                 = <<-EOT
  - hosts: localhost
    become: false
  EOT
  inventory                = "# localhost"
  ansible_options = {
    known_hosts = [
      "",
      "10.0.0.1 ssh-ed25519 invalid-public-key",
      "10.0.0.2 ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIL5L/oFlqWdgPhttkYs63jnETm6LQHlUen9G/CIbxR8p\nadditional-data",
    ]
  }
}
