resource "ansible_navigator_run" "test" {
  ansible_navigator_binary = var.ansible_navigator_binary
  playbook                 = <<-EOT
  - hosts: localhost
    become: false
  EOT
  inventory                = "# localhost"
  run_on_destroy           = false
  timeouts = {
    create = "90m"
    update = "90m"
    delete = "90m"
  }
}

variable "ansible_navigator_binary" {
  type     = string
  nullable = false
}
