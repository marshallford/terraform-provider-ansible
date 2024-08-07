resource "ansible_navigator_run" "test" {
  ansible_navigator_binary = var.ansible_navigator_binary
  playbook                 = <<-EOT
  - hosts: localhost
    become: false
  EOT
  inventory                = "# localhost"
  ansible_options = {
    force_handlers = true
    skip_tags      = ["tag1", "tag2"]
    start_at_task  = "task name"
    limit          = ["host1", "host2"]
    tags           = ["tag3", "tag4"]
  }
}

variable "ansible_navigator_binary" {
  type     = string
  nullable = false
}
