resource "ansible_navigator_run" "test" {
  ansible_navigator_binary = "%s"
  playbook                 = <<-EOT
  - hosts: localhost
    become: false
    tasks:
    - ansible.builtin.assert:
        that:
        - lookup('ansible.builtin.env', 'ANSIBLE_TF_OPERATION') == 'create'
        - lookup('ansible.builtin.env', 'TF_ACC') == '1'
        - lookup('ansible.builtin.env', 'TESTING') == 'abc'
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
}
