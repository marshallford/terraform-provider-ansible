resource "ansible_navigator_run" "test" {
  ansible_navigator_binary = var.ansible_navigator_binary
  playbook                 = <<-EOT
  - hosts: localhost
    become: false
  EOT
  inventory                = "# localhost"
  working_directory        = var.working_directory
}

variable "working_directory" {
  type     = string
  nullable = false
}
