resource "ansible_navigator_run" "test" {
  ansible_navigator_binary = var.ansible_navigator_binary
  playbook                 = <<-EOT
  - hosts: localhost
    become: false
  EOT
  inventory                = "# localhost"
  execution_environment = {
    pull_arguments = var.pull_args
  }
  artifact_queries = {
    "pull_args" = {
      jq_filter = <<-EOT
      .settings_entries."ansible-navigator"."execution-environment".pull.arguments
      EOT
    }
  }
}

variable "pull_args" {
  type     = list(string)
  nullable = false
}
