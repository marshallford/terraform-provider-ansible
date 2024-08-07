resource "ansible_navigator_run" "test" {
  ansible_navigator_binary = var.ansible_navigator_binary
  playbook                 = <<-EOT
  - hosts: localhost
    become: false
    tasks:
    - ansible.builtin.assert:
        that:
        - lookup('ansible.builtin.env', 'ANSIBLE_TF_OPERATION') == '${var.operation}'
        - lookup('ansible.builtin.env', 'TF_ACC') == '1'
        - lookup('ansible.builtin.env', 'TESTING') == 'abc'
      when: lookup('ansible.builtin.env', 'ANSIBLE_TF_OPERATION') != 'delete'
  EOT
  inventory                = "# localhost"
  execution_environment = {
    environment_variables_pass = [
      "TF_ACC",
    ]
    environment_variables_set = {
      "TESTING" = "abc"
    }
  }
  run_on_destroy = true
}

variable "ansible_navigator_binary" {
  type     = string
  nullable = false
}

variable "operation" {
  type     = string
  nullable = false
}
