resource "ansible_navigator_run" "test" {
  ansible_navigator_binary = "%s"
  working_directory        = "%s"
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
