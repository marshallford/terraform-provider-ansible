resource "ansible_navigator_run" "test" {
  ansible_navigator_binary = "${path.module}/${var.ansible_navigator_binary}"
  playbook                 = <<-EOT
  - hosts: localhost
    become: false
  EOT
  inventory                = "# localhost"
  artifact_queries = {
    "test" = {
      jq_filter = "now"
    }
  }
  working_directory = ".."
  run_on_destroy = true
  timeouts = {
    create = "90m"
    update = "90m"
    delete = "90m"
  }
}
