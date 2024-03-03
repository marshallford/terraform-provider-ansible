resource "ansible_navigator_run" "test" {
  ansible_navigator_binary = "%s"
  working_directory = "%s"
  playbook          = <<-EOT
  - hosts: some_group
    become: false
    tasks:
    - ansible.builtin.debug:
        msg: "{{ some_var }}"
  EOT
  inventory         = <<-EOT
  all:
    children:
      some_group:
        hosts:
          local_container:
            ansible_connection: local
            some_var: hello world!
  EOT
}
