resource "ansible_navigator_run" "test" {
  ansible_navigator_binary = "%s"
  working_directory        = "%s"
  playbook                 = <<-EOT
  - hosts: localhost
    become: false
  EOT
  inventory                = "# localhost"
  execution_environment = {
    pull_policy = "missing" # speeds up tests
  }
  ansible_options = {
    force_handlers = true
    limit          = ["host1", "host2"]
    tags           = ["tag1", "tag2"]
  }
}
