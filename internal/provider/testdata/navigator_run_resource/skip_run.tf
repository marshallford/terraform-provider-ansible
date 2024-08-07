resource "ansible_navigator_run" "test" {
  ansible_navigator_binary = var.ansible_navigator_binary
  playbook                 = <<-EOT
  - hosts: localhost
    become: false
  EOT
  inventory                = "# localhost"
  run_on_destroy           = true
  timeouts = {
    create = "60m"
    update = "60m"
    delete = "60m"
  }
}

variable "ansible_navigator_binary" {
  type     = string
  nullable = false
}
