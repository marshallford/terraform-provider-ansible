resource "ansible_navigator_run" "test" {
  ansible_navigator_binary = var.ansible_navigator_binary
  playbook                 = <<-EOT
  - hosts: localhost
    gather_facts: false
    become: false
    tasks:
    - ansible.builtin.assert:
        that: lookup('ansible.builtin.env', 'ANSIBLE_TF_OPERATION') != 'delete'
  EOT
  inventory                = "# localhost"
  run_on_destroy           = true
  destroy_playbook         = <<-EOT
  - hosts: localhost
    become: false
    tasks:
    - ansible.builtin.assert:
        that: lookup('ansible.builtin.env', 'ANSIBLE_TF_OPERATION') == 'delete'
  EOT
}
