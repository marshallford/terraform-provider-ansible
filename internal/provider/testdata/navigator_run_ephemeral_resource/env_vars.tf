ephemeral "ansible_navigator_run" "test" {
  ansible_navigator_binary = var.ansible_navigator_binary
  playbook                 = <<-EOT
  - hosts: localhost
    gather_facts: false
    become: false
    tasks:
    - ansible.builtin.assert:
        that:
        - lookup('ansible.builtin.env', 'ANSIBLE_TF_INVENTORY') != ''
        - lookup('ansible.builtin.env', 'ANSIBLE_TF_OPERATION') == 'open'
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
