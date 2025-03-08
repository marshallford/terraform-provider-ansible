resource "ansible_navigator_run" "test" {
  ansible_navigator_binary = var.ansible_navigator_binary
  playbook                 = <<-EOT
  - hosts: some_group
    become: false
    tasks:
    - ansible.builtin.debug:
        msg: "{{ some_var }}"
    - ansible.builtin.assert:
        that: inventory_hostname == "local_container"
  EOT
  inventory = yamlencode({
    all = {
      children = {
        some_group = {
          hosts = {
            local_container = {
              ansible_connection = "local"
              some_var           = "hello world!"
            }
          }
        }
      }
    }
  })
  run_on_destroy = true
}
