resource "ansible_navigator_run" "test" {
  ansible_navigator_binary = var.ansible_navigator_binary
  playbook                 = <<-EOT
  - hosts: localhost
    gather_facts: false
    become: false
    tasks:
    - ansible.builtin.assert:
        that:
        - extra_var_string == "hello"
        - extra_var_number == 42
        - extra_var_bool == true
        - extra_var_map.key_one == "value_one"
        - extra_var_map.key_two == "value_two"
  EOT
  inventory                = "# localhost"
  ansible_options = {
    extra_vars = yamlencode({
      extra_var_string = "hello"
      extra_var_number = 42
      extra_var_bool   = true
      extra_var_map = {
        key_one = "value_one"
        key_two = "value_two"
      }
    })
  }
}
