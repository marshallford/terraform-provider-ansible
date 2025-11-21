action "ansible_navigator_run" "test" {
  config {
    ansible_navigator_binary = var.ansible_navigator_binary
    playbook                 = <<-EOT
    - hosts: some_group
      gather_facts: false
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
  }
}

resource "terraform_data" "test" {
  lifecycle {
    action_trigger {
      events  = [after_create]
      actions = [action.ansible_navigator_run.test]
    }
  }
}
