resource "ansible_navigator_run" "test" {
  ansible_navigator_binary = "%s"
  working_directory        = "%s"
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
