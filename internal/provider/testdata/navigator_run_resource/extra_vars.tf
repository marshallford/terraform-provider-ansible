resource "ansible_navigator_run" "test" {
  ansible_navigator_binary = var.ansible_navigator_binary
  playbook                 = <<-EOT
  - hosts: localhost
    become: false
    tasks:
    - ansible.builtin.assert:
        that: testing.extra_vars_test == "abc"
  EOT
  inventory                = "# localhost"
  execution_environment = {
    enabled = var.ee_enabled
  }
  ansible_options = {
    extra_vars_wo = yamlencode({
      "testing" = {
        "extra_vars_test" = "abc"
      }
    })
    extra_vars_wo_revision = var.revision
  }
}

variable "ee_enabled" {
  type     = bool
  nullable = false
}

variable "revision" {
  type     = number
  nullable = false
}
