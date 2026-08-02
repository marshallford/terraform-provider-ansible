resource "ansible_navigator_run" "test" {
  ansible_navigator_binary = var.ansible_navigator_binary
  playbook                 = <<-EOT
  - hosts: localhost
    gather_facts: false
    become: false
  EOT
  inventory                = "# localhost"
  artifact_queries = {
    # Parses, so the schema validator accepts it, but fails against the artifact.
    "runtime" = {
      jq_filter = ".stdout | tonumber"
    }
  }
}
