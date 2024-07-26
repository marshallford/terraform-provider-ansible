resource "ansible_navigator_run" "test" {
  ansible_navigator_binary = var.ansible_navigator_binary
  playbook                 = <<-EOT
  - hosts: localhost
    become: false
  EOT
  inventory                = "# localhost"
  execution_environment = {
    pull_arguments = var.pull_arguments
  }
  artifact_queries = {
    pull_args = {
      jsonpath = "$.settings_entries.ansible-navigator.execution-environment.pull.arguments"
    }
  }
}

variable "ansible_navigator_binary" {
  type     = string
  nullable = false
}

variable "pull_arguments" {
  type     = list(string)
  nullable = false
}
