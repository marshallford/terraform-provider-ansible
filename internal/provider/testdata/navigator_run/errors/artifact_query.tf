resource "ansible_navigator_run" "test" {
  ansible_navigator_binary = var.ansible_navigator_binary
  playbook                 = <<-EOT
  - hosts: localhost
    become: false
  EOT
  inventory                = "# localhost"
  artifact_queries = {
    "test" = {
      jsonpath = "!"
    }
  }
}

variable "ansible_navigator_binary" {
  type     = string
  nullable = false
}
