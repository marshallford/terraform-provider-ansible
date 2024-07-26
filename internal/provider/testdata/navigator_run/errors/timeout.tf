resource "ansible_navigator_run" "test" {
  ansible_navigator_binary = var.ansible_navigator_binary
  playbook                 = <<-EOT
  - hosts: localhost
    become: false
    tasks:
    - ansible.builtin.pause:
        seconds: 10
  EOT
  inventory                = "# localhost"
  timeouts = {
    create = "5s"
  }
}

variable "ansible_navigator_binary" {
  type     = string
  nullable = false
}
