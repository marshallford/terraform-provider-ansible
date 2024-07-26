resource "ansible_navigator_run" "test" {
  ansible_navigator_binary = var.ansible_navigator_binary
  playbook                 = <<-EOT
  - hosts: localhost
    become: false
  EOT
  inventory                = "# localhost"
  execution_environment = {
    environment_variables_set = {
      ""            = "EMPTY_KEY"
      "INVALID=KEY" = "value"
    }
  }
}

variable "ansible_navigator_binary" {
  type     = string
  nullable = false
}
